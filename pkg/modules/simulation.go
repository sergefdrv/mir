// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/testsim"
	t "github.com/filecoin-project/mir/pkg/types"
	"google.golang.org/protobuf/proto"
)

type Simulation struct {
	*testsim.Runtime
	//nodes   map[t.NodeID]*SimNode
	//stopChan chan struct{}
}

func NewSimulation(rnd *rand.Rand) *Simulation {
	return &Simulation{
		Runtime: testsim.NewRuntime(rnd),
		//nodes:      make(map[t.NodeID]*SimNode),
		//stopChan: make(chan struct{}),
	}
}

// func (s *Simulation) Stop() {
// 	close(s.stopChan)
// }

type EventDelayFn func(e *eventpb.Event) time.Duration

type SimNode struct {
	*Simulation
	delayFn EventDelayFn
	//id          t.NodeID
	moduleChans map[t.ModuleID]*testsim.Chan
}

func (s *Simulation) NewNode( /* id t.NodeID, */ delayFn EventDelayFn) *SimNode {
	n := &SimNode{
		Simulation: s,
		delayFn:    delayFn,
		//id:          id,
		moduleChans: make(map[t.ModuleID]*testsim.Chan),
	}
	//s.nodes[id] = n
	return n
}

// func (n *SimNode) SendEventList(proc *testsim.Process, eventList *events.EventList) {
// 	// eventsMap := make(map[t.ModuleID][]*eventpb.Event)

// 	it := eventList.Iterator()
// 	for e := it.Next(); e != nil; e = it.Next() {
// 		n.SendEvent(proc, e)
// 		// eventsMap[m] = append(eventsMap[m], e)
// 	}

// 	// for m, v := range eventsMap {
// 	// 	proc.Send(n.moduleChans[m], v)
// 	// }
// }

// func (n *SimNode) SendEvent(proc *testsim.Process, e *eventpb.Event) {
// 	m := t.ModuleID(e.DestModule)
// 	proc.Send(n.moduleChans[m], e)
// }

// func (n *SimNode) sendFollowUps(proc *testsim.Process, origEvents []*eventpb.Event) {
// 	for _, e := range origEvents {
// 		for _, e := range e.Next {
// 			n.SendEvent(proc, e)
// 		}
// 	}
// }

// func (n *SimNode) recvEvent(proc *testsim.Process, simChan *testsim.Chan) (e *eventpb.Event, ok bool) {
// 	v, ok := proc.Recv(simChan)
// 	if !ok {
// 		return nil, false
// 	}
// 	return v.(*eventpb.Event), true
// }

func (n *SimNode) SendEvents(proc *testsim.Process, eventList *events.EventList) {
	eventsMap := make(map[t.ModuleID]*events.EventList)

	it := eventList.Iterator()
	for e := it.Next(); e != nil; e = it.Next() {
		m := t.ModuleID(e.DestModule)
		if eventsMap[m] == nil {
			eventsMap[m] = events.EmptyList()
		}
		eventsMap[m].PushBack(e)
	}

	//proc := n.Spawn()
	//go func() {
	for m, v := range eventsMap {
		//proc.Delay(n.RandDuration(1, time.Microsecond))
		proc.Delay(1)
		fmt.Println("Sending", v.Slice())
		proc.Send(n.moduleChans[m], v)
		fmt.Println("Sent", v.Slice())
	}
	//proc.Exit()
	//}()
}

func (n *SimNode) recvEvents(proc *testsim.Process, simChan *testsim.Chan) (eventList *events.EventList, ok bool) {
	v, ok := proc.Recv(simChan)
	if !ok {
		return nil, false
	}
	return v.(*events.EventList), true
}

func (n *SimNode) WrapModules(mods Modules) Modules {
	wrapped := make(Modules, len(mods))
	for k, v := range mods {
		wrapped[k] = n.WrapModule(k, v)
	}
	return wrapped
}

func (n *SimNode) WrapModule(id t.ModuleID, m Module) Module {
	moduleChan := testsim.NewChan()
	n.moduleChans[id] = moduleChan

	// var applyFn func(ctx context.Context, eventList *events.EventList) (*events.EventList, error)

	switch m := m.(type) {
	case ActiveModule:
		// applyFn = func(ctx context.Context, eventList *events.EventList) (*events.EventList, error) {
		// 	return nil, m.ApplyEvents(ctx, eventList)
		// }
		//return n.newModule(m, moduleChan)
		//return ActiveModule(&activeSimModule{&simModule{m, n, moduleChan}})
		return n.wrapActive(m, moduleChan)
	case PassiveModule:
		// applyFn = func(ctx context.Context, eventList *events.EventList) (*events.EventList, error) {
		// 	return m.ApplyEvents(eventList)
		// }
		//return passiveFromActive(n.newModule(activeFromPassive(m), moduleChan))
		// return PassiveModule(&passiveSimModule{&simModule{m, n, moduleChan}})
		return n.wrapPassive(m, moduleChan)
	default:
		panic("Unexpected module type")
	}

	// return n.newModule(applyFn, moduleChan)
}

func (n *SimNode) Start(done chan struct{}) {
	proc := n.Spawn()
	go func() {
		//proc.Delay(n.RandDuration(1, time.Microsecond))
		//n.SendEvents( /* proc, */ initEvents)
		// initEvents = events.EmptyList()
		// if _, ok := n.moduleChans["wal"]; ok {
		// 	initEvents := events.EmptyList()
		// 	initEvents.PushBack(events.WALLoadAll("wal"))
		// 	//initEvents.PushBack(events.Init("wal"))
		// 	proc.Send(n.moduleChans["wal"], initEvents)
		// 	proc.Delay(n.RandDuration(1, time.Microsecond))
		// 	//proc.Delay(0)
		// }
		for m := range n.moduleChans {
			initEvents := events.EmptyList()
			if m == "wal" {
				// 	//initEvents.PushBack(events.WALLoadAll("wal"))
				//	continue
			}
			initEvents.PushBack(events.Init(m))
			//n.SendEvents( /* proc,  */ initEvents)
			proc.Send(n.moduleChans[m], initEvents)
			//proc.Delay(n.RandDuration(1, time.Microsecond))
			proc.Delay(0)
		}
		proc.Exit()
		close(done)
	}()
}

type applyEventsFn func(ctx context.Context, eventList *events.EventList) (*events.EventList, error)

type eventsIn struct {
	ctx       context.Context
	eventList *events.EventList
}

type eventsOut struct {
	eventList *events.EventList
	err       error
}

type simModule struct {
	// Module
	*SimNode
	applyFn applyEventsFn
	inChan  chan eventsIn
	outChan chan eventsOut
	//eventChan chan *eventpb.Event
	//	eventChan chan *eventpb.Event
	//eventChan *testsim.Chan
	simChan *testsim.Chan
}

func newSimModule(n *SimNode, m Module, simChan *testsim.Chan) *simModule {
	// sm := &simModule{
	// 	// Module:  m,
	// 	SimNode: n,
	// 	// inChan:  make(chan eventsIn, 1),
	// 	// outChan: make(chan eventsOut, 1),
	// 	origEventsChan: make(chan *eventpb.Event, 1),
	// 	//simChan:        simChan,
	// }

	//eventChan := make(chan *eventpb.Event /* , 1 */)
	//eventChan := testsim.NewChan()
	// eventChan := make(chan *eventpb.Event)
	// proc := n.Spawn()
	// go func() {
	// 	for {
	// 		e, ok := n.recvEvent(proc, simChan)
	// 		if !ok {
	// 			return
	// 		}
	// 		eventChan <- e
	// 		//proc.Send(eventChan, e)
	// 	}
	// }()

	var applyFn applyEventsFn
	switch m := m.(type) {
	case PassiveModule:
		applyFn = func(_ context.Context, eventList *events.EventList) (*events.EventList, error) {
			return m.ApplyEvents(eventList)
		}
	case ActiveModule:
		applyFn = func(ctx context.Context, eventList *events.EventList) (*events.EventList, error) {
			return events.EmptyList(), m.ApplyEvents(ctx, eventList)
		}
	default:
		panic("Unexpected module type")
	}

	sm := &simModule{
		SimNode: n,
		applyFn: applyFn,
		//eventChan: eventChan,
		inChan:  make(chan eventsIn, 1),
		outChan: make(chan eventsOut, 1),
		simChan: simChan,
	}

	//d := n.RandDuration(1, time.Millisecond)
	go func() {
		proc := n.Spawn()
		//	proc.Delay(d)
		sm.run(proc, applyFn)
	}()

	return sm
}

func (m *simModule) run(proc *testsim.Process, applyFn applyEventsFn) {
	// 	for {
	// 		e, ok := m.SimNode.recvEvent(proc, m.simChan)
	// 		if !ok {
	// 			return
	// 		}

	// 		m.origEventsChan <- e
	// 	}
	origEvents := events.EmptyList()
	// origEvents, ok := m.SimNode.recvEvents(proc, m.simChan)
	// if !ok {
	// 	fmt.Println("exiting module loop")
	// 	return
	// }
	// it := origEvents.Iterator()
	// for e := it.Next(); e != nil; e = it.Next() {
	// 	if !proc.Delay(m.SimNode.delayFn(e)) {
	// 		return
	// 	}
	// }

	// fmt.Println("Received", origEvents.Slice())
	for {
		if origEvents.Len() == 0 {
			newOrigEvents, ok := m.SimNode.recvEvents(proc, m.simChan)
			if !ok {
				fmt.Println("exiting module loop")
				return
			}
			origEvents.PushBackList(newOrigEvents)

			fmt.Println("Received", newOrigEvents.Slice())
		}

		//fmt.Println("origEvents", origEvents.Slice())

		in := <-m.inChan

		for origEvents.Len() < in.eventList.Len() {
			//fmt.Println(origEvents.Slice(), in.eventList.Slice())
			//panic("Unexpected number of events")
			fmt.Println("Waiting to receive missing messages")
			newOrigEvents, ok := m.SimNode.recvEvents(proc, m.simChan)
			if !ok {
				return
			}
			origEvents.PushBackList(newOrigEvents)
			fmt.Println("Received", newOrigEvents.Slice())

			// it := newOrigEvents.Iterator()
			// for e := it.Next(); e != nil; e = it.Next() {
			// 	if !proc.Delay(m.SimNode.delayFn(e)) {
			// 		return
			// 	}
			// }
		}

		fmt.Println("applying", in.eventList.Slice())
		var out eventsOut
		out.eventList, out.err = applyFn(in.ctx, in.eventList)

		// go m.run(proc.Fork(), applyFn)
		// defer proc.Exit()

		it := in.eventList.Iterator()
		it2 := origEvents.Iterator()
		for e := it.Next(); e != nil; e = it.Next() {
			e2 := it2.Next()
			e2 = proto.Clone(e2).(*eventpb.Event)
			e2.Next = nil
			if e.String() != e2.String() {
				panic(fmt.Sprint(e, "!=", e2))
			}
			switch e.Type.(type) {
			//case *eventpb.Event_WalLoadAll:
			//case *eventpb.Event_Init:
			default:
				// if !proc.Delay(m.SimNode.delayFn(e)) {
				// 	return
				// }
			}
		}

		//go func() {
		//m.outChan <- out
		//}()

		if out.err == nil {
			inLen := in.eventList.Len()
			origS := origEvents.Slice()
			origS, pend := origS[:inLen], origS[inLen:]
			origEvents = events.EmptyList().PushBackSlice(origS)
			_, followUps := origEvents.StripFollowUps()
			origEvents = events.EmptyList().PushBackSlice(pend)
			followUps.PushBackList(out.eventList)

			//m.SimNode.SendEvents(followUps)
			go func(proc *testsim.Process) {
				m.outChan <- out
				m.SimNode.SendEvents(proc, followUps)
				proc.Exit()
			}(proc.Fork())
		}

		// m.SimNode.SendEventList(proc, followUps)

		//m.outChan <- out
	}
}

func (m *simModule) applyEvents(ctx context.Context, eventList *events.EventList) (eventsOut *events.EventList, err error) {
	m.inChan <- eventsIn{ctx, eventList}
	select {
	case out := <-m.outChan:
		return out.eventList, out.err
	case <-time.After(1 * time.Second):
		fmt.Println("Got stuck applying", eventList.Slice())
		//panic("")
		select {}
	}
	// proc := m.Simulation.Spawn()

	// done := make(chan struct{})
	// go func() {
	// 	select {
	// 	case <-ctx.Done():
	// 		proc.Kill()
	// 	case <-done:
	// 	}
	// }()

	// // fmt.Println(eventList.Slice())

	// // origEvents := make([]*eventpb.Event, eventList.Len())
	// // for i := range origEvents {
	// // 	// e, ok := m.SimNode.recvEvent(proc, m.simChan)
	// // 	// if !ok {
	// // 	// 	panic("Module simulation process died")
	// // 	// }
	// // 	// origEvents[i] = e
	// // 	origEvents[i] = <-m.origEventsChan
	// // }
	// origEvents := events.EmptyList()
	// // for i := 0; i < eventList.Len(); i++ {
	// // v, ok := proc.Recv(m.eventChan)
	// // if !ok {
	// // 	return
	// // }
	// // origEvents.PushBack(v.(*eventpb.Event))
	// //origEvents.PushBack(<-m.eventChan)
	// // }
	// for origEvents.Len() < eventList.Len() {
	// 	origEvents.PushBack(<-m.eventChan)
	// }

	// // for _, e := range origEvents {
	// it := eventList.Iterator()
	// for e := it.Next(); e != nil; e = it.Next() {
	// 	if !proc.Delay(m.SimNode.delayFn(e)) {
	// 		return nil, nil
	// 	}
	// }

	// // switch m := m.Module.(type) {
	// // case PassiveModule:
	// // 	eventsOut, err = m.ApplyEvents(eventList)
	// // case ActiveModule:
	// // 	eventsOut, err = events.EmptyList(), m.ApplyEvents(ctx, eventList)
	// // default:
	// // 	panic("Unexpected module type")
	// // }
	// eventsOut, err = m.applyFn(ctx, eventList)
	// if err != nil {
	// 	return nil, err
	// }

	// go func() {
	// 	defer close(done)
	// 	defer proc.Exit()

	// 	// for _, e := range origEvents {
	// 	// 	followUps := e.Next
	// 	// 	for _, e := range followUps {
	// 	// 		m.SimNode.SendEvent(proc, e)
	// 	// 	}
	// 	// }
	// 	_, followUps := origEvents.StripFollowUps()
	// 	m.SimNode.SendEventList(proc, followUps)
	// 	m.SimNode.SendEventList(proc, eventsOut)

	// 	// it := eventsOut.Iterator()
	// 	// for e := it.Next(); e != nil; e = it.Next() {
	// 	// 	m.SimNode.SendEvent(proc, e)
	// 	// }

	// }()

	// return eventsOut, nil
}

type passiveSimModule struct {
	PassiveModule
	*simModule
}

func (n *SimNode) wrapPassive(m PassiveModule, simChan *testsim.Chan) PassiveModule {
	return &passiveSimModule{m, newSimModule(n, m, simChan)}
}

func (m *passiveSimModule) ApplyEvents(eventList *events.EventList) (*events.EventList, error) {
	return m.applyEvents(context.Background(), eventList)
	// proc := m.Simulation.Spawn()

	// origEvents := make([]*eventpb.Event, eventList.Len())
	// for i := range origEvents {
	// 	e, ok := m.SimNode.recvEvent(proc, m.simChan)
	// 	if !ok {
	// 		panic("Module simulation process died")
	// 	}
	// 	origEvents[i] = e
	// }

	// for _, e := range origEvents {
	// 	if !proc.Delay(m.SimNode.delayFn(e)) {
	// 		return nil, nil // simulation stopped
	// 	}
	// }

	// eventsOut, err := m.PassiveModule.ApplyEvents(eventList)
	// if err != nil {
	// 	return nil, err
	// }

	// go func() {
	// 	m.SimNode.sendFollowUps(proc, origEvents)
	// 	proc.Exit()
	// }()

	// return eventsOut, nil
}

type activeSimModule struct {
	ActiveModule
	*simModule
}

func (n *SimNode) wrapActive(m ActiveModule, simChan *testsim.Chan) ActiveModule {
	return &activeSimModule{m, newSimModule(n, m, simChan)}
}

func (m *activeSimModule) ApplyEvents(ctx context.Context, eventList *events.EventList) error {
	_, err := m.applyEvents(ctx, eventList)
	return err
	// proc := m.Simulation.Spawn()

	// done := make(chan struct{})
	// go func() {
	// 	select {
	// 	case <-ctx.Done():
	// 		proc.Kill()
	// 	case <-done:
	// 	}
	// }()

	// origEvents := make([]*eventpb.Event, eventList.Len())
	// for i := range origEvents {
	// 	e, ok := m.SimNode.recvEvent(proc, m.simChan)
	// 	if !ok {
	// 		panic("Module simulation process died")
	// 	}
	// 	origEvents[i] = e
	// }

	// for _, e := range origEvents {
	// 	if !proc.Delay(m.SimNode.delayFn(e)) {
	// 		return nil // simulation stopped
	// 	}
	// }

	// err := m.ActiveModule.ApplyEvents(ctx, eventList)
	// if err != nil {
	// 	return err
	// }

	// go func() {
	// 	defer close(done)

	// 	m.SimNode.sendFollowUps(proc, origEvents)
	// 	proc.Exit()
	// }()

	// return nil
}

// func (m *activeSimModule) EventsOut() <-chan *events.EventList {
// 	return m.Module.(ActiveModule).EventsOut()
// }

// type simModuleWrapper struct {
// 	ActiveModule
// 	*SimNode
// 	//*testsim.Process
// 	//wrapped    ActiveModule
// 	// inChan  chan eventsIn
// 	// errChan chan error
// 	// outChan chan *events.EventList
// 	simChan *testsim.Chan
// }

// type eventsIn struct {
// 	ctx    context.Context
// 	events *events.EventList
// }

// func (n *SimNode) newModule(m ActiveModule, moduleChan *testsim.Chan) *simModuleWrapper {
// 	wrapper := &simModuleWrapper{
// 		ActiveModule: m,
// 		SimNode:      n,
// 		//Process:    n.Simulation.Spawn(),
// 		//wrapped:    m,
// 		// inChan:  make(chan eventsIn, 1),
// 		// errChan: make(chan error, 1),
// 		// outChan: make(chan *events.EventList, 1),
// 		simChan: moduleChan,
// 	}

// 	//go wrapper.handleEvents( /* n.Simulation.Spawn(), */ m)

// 	return wrapper
// }

// func (m *simModuleWrapper) ImplementsModule() {}

// func (m *simModuleWrapper) EventsOut() <-chan *events.EventList {
// 	return m.outChan
// }

// func (m *simModuleWrapper) ApplyEvents(ctx context.Context, eventList *events.EventList) (err error) {
// 	// m.inChan <- eventsIn{ctx, events}
// 	// return <-m.errChan

// 	return m.ActiveModule.ApplyEvents(ctx, eventList)
// }

// func (m *simModuleWrapper) handleEvents( /* proc *testsim.Process, */ wrapped ActiveModule) {
// 	eventsOut := wrapped.EventsOut()

// 	for {
// 		select {
// 		case in := <-m.inChan:
// 			origEvents := make([]*eventpb.Event, in.events.Len())
// 			for i := range origEvents {
// 				v, ok := proc.Recv(m.simChan)
// 				if !ok {
// 					panic("Module simulation process died")
// 				}
// 				e := v.(*eventpb.Event)
// 				origEvents[i] = e
// 			}

// 			for _, e := range origEvents {
// 				ok := proc.Delay(m.SimNode.delayFn(e))
// 				if !ok {
// 					return // simulation stopped
// 				}
// 			}

// 			err := wrapped.ApplyEvents(in.ctx, in.events)
// 			m.errChan <- err
// 			if err != nil {
// 				return
// 			}

// 			for _, e := range origEvents {
// 				followUps := e.Next
// 				for _, e := range followUps {
// 					proc.Send(m.moduleChans[t.ModuleID(e.DestModule)], e)
// 				}
// 			}
// 		case out := <-eventsOut:
// 			m.outChan <- out
// 			it := out.Iterator()
// 			for e := it.Next(); e != nil; e = it.Next() {
// 				proc.Send(m.moduleChans[t.ModuleID(e.DestModule)], e)
// 			}
// 		case <-m.SimNode.Simulation.stopChan:
// 			return
// 		}
// 	}
// }

// type passiveToActiveTransformer struct {
// 	PassiveModule
// 	eventsOut chan *events.EventList
// }

// func activeFromPassive(m PassiveModule) ActiveModule {
// 	return &passiveToActiveTransformer{m, make(chan *events.EventList, 1)}
// }

// func (m *passiveToActiveTransformer) ApplyEvents(ctx context.Context, events *events.EventList) error {
// 	events, err := m.PassiveModule.ApplyEvents(events)
// 	if err != nil {
// 		return err
// 	}
// 	select {
// 	case m.eventsOut <- events:
// 	case <-ctx.Done():
// 	}
// 	return nil
// }

// func (m *passiveToActiveTransformer) EventsOut() <-chan *events.EventList {
// 	return m.eventsOut
// }

// type activeToPassiveTransformer struct {
// 	ActiveModule
// 	eventsOut <-chan *events.EventList
// }

// func passiveFromActive(m ActiveModule) PassiveModule {
// 	return &activeToPassiveTransformer{m, m.EventsOut()}
// }

// func (m *activeToPassiveTransformer) ApplyEvents(events *events.EventList) (*events.EventList, error) {
// 	if err := m.ActiveModule.ApplyEvents(context.Background(), events); err != nil {
// 		return nil, err
// 	}
// 	return <-m.eventsOut, nil
// }
