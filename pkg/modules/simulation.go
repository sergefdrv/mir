// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"context"
	"math/rand"
	"time"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/testsim"
	t "github.com/filecoin-project/mir/pkg/types"
)

type Simulation struct {
	*testsim.Simulation
	//nodes   map[t.NodeID]*SimNode
	//stopChan chan struct{}
}

func NewSimulation(rnd *rand.Rand) *Simulation {
	return &Simulation{
		Simulation: testsim.NewSimulation(rnd),
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

func (n *SimNode) SendEventList(proc *testsim.Process, eventList *events.EventList) {
	// eventsMap := make(map[t.ModuleID][]*eventpb.Event)

	it := eventList.Iterator()
	for e := it.Next(); e != nil; e = it.Next() {
		n.SendEvent(proc, e)
		// eventsMap[m] = append(eventsMap[m], e)
	}

	// for m, v := range eventsMap {
	// 	proc.Send(n.moduleChans[m], v)
	// }
}

func (n *SimNode) SendEvent(proc *testsim.Process, e *eventpb.Event) {
	m := t.ModuleID(e.DestModule)
	proc.Send(n.moduleChans[m], e)
}

func (n *SimNode) sendFollowUps(proc *testsim.Process, origEvents []*eventpb.Event) {
	for _, e := range origEvents {
		for _, e := range e.Next {
			n.SendEvent(proc, e)
		}
	}
}

func (n *SimNode) recvEvent(proc *testsim.Process, simChan *testsim.Chan) (e *eventpb.Event, ok bool) {
	v, ok := proc.Recv(simChan)
	if !ok {
		return nil, false
	}
	return v.(*eventpb.Event), true
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
		return ActiveModule(&activeSimModule{&simModule{m, n, moduleChan}})
	case PassiveModule:
		// applyFn = func(ctx context.Context, eventList *events.EventList) (*events.EventList, error) {
		// 	return m.ApplyEvents(eventList)
		// }
		//return passiveFromActive(n.newModule(activeFromPassive(m), moduleChan))
		return PassiveModule(&passiveSimModule{&simModule{m, n, moduleChan}})
	default:
		panic("Unexpected module type")
	}

	// return n.newModule(applyFn, moduleChan)
}

// type simApplyEventsFn func(ctx context.Context, eventList *events.EventList) (*events.EventList, error)

type simModule struct {
	Module
	*SimNode
	simChan *testsim.Chan
}

func (m *simModule) applyEvents(ctx context.Context, eventList *events.EventList) (eventsOut *events.EventList, err error) {
	proc := m.Simulation.Spawn()

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			proc.Kill()
		case <-done:
		}
	}()

	// fmt.Println(eventList.Slice())

	origEvents := make([]*eventpb.Event, eventList.Len())
	for i := range origEvents {
		e, ok := m.SimNode.recvEvent(proc, m.simChan)
		if !ok {
			panic("Module simulation process died")
		}
		origEvents[i] = e
	}

	for _, e := range origEvents {
		if !proc.Delay(m.SimNode.delayFn(e)) {
			return nil, nil
		}
	}

	switch m := m.Module.(type) {
	case PassiveModule:
		eventsOut, err = m.ApplyEvents(eventList)
	case ActiveModule:
		eventsOut, err = events.EmptyList(), m.ApplyEvents(ctx, eventList)
	default:
		panic("Unexpected module type")
	}

	if err != nil {
		return nil, err
	}

	go func() {
		defer close(done)

		for _, e := range origEvents {
			followUps := e.Next
			for _, e := range followUps {
				m.SimNode.SendEvent(proc, e)
			}
		}

		it := eventsOut.Iterator()
		for e := it.Next(); e != nil; e = it.Next() {
			m.SimNode.SendEvent(proc, e)
		}

		proc.Exit()
	}()

	return eventsOut, nil
}

type passiveSimModule struct{ *simModule }

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
	*simModule
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

func (m *activeSimModule) EventsOut() <-chan *events.EventList {
	return m.Module.(ActiveModule).EventsOut()
}

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
