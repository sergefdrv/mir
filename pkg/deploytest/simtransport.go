// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package deploytest

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/testsim"
	t "github.com/filecoin-project/mir/pkg/types"
)

type MessageDelayFn func(from, to t.NodeID) time.Duration

type SimTransport struct {
	*testsim.Simulation
	delayFn  MessageDelayFn
	nodes    map[t.NodeID]*simTransportModule
	stopChan chan struct{}
}

func NewSimTransport(s *testsim.Simulation, delayFn MessageDelayFn) *SimTransport {
	return &SimTransport{
		Simulation: s,
		delayFn:    delayFn,
		nodes:      make(map[t.NodeID]*simTransportModule),
		stopChan:   make(chan struct{}),
	}
}

func (t *SimTransport) NodeModule(id t.NodeID, node *modules.SimNode) modules.ActiveModule {
	m := t.nodes[id]
	if m == nil {
		m = t.newModule(id, node)
		t.nodes[id] = m
	}
	return m
}

func (t *SimTransport) Stop() {
	close(t.stopChan)
}

type simTransportModule struct {
	*SimTransport
	*modules.SimNode
	id t.NodeID
	// inChan  chan *events.EventList
	// errChan chan error
	outChan chan *events.EventList
	simChan *testsim.Chan
}

func (t *SimTransport) newModule(id t.NodeID, node *modules.SimNode) *simTransportModule {
	m := &simTransportModule{
		SimTransport: t,
		SimNode:      node,
		id:           id,
		// inChan:       make(chan *events.EventList, 1),
		// errChan:      make(chan error, 1),
		outChan: make(chan *events.EventList, 1),
		simChan: testsim.NewChan(),
	}

	//go m.handleInChan(t.Simulation.Spawn())
	go m.handleOutChan(t.Simulation.Spawn())

	return m
}

func (m *simTransportModule) ImplementsModule() {}

func (m *simTransportModule) ApplyEvents(ctx context.Context, eventList *events.EventList) error {
	_, err := modules.ApplyEventsSequentially(eventList, func(e *eventpb.Event) (*events.EventList, error) {
		return events.EmptyList(), m.applyEvent(ctx, e)
	})
	return err
	// select {
	// case m.inChan <- eventList:
	// 	return <-m.errChan
	// case <-ctx.Done():
	// }
	// return nil
}

func (m *simTransportModule) applyEvent(ctx context.Context, e *eventpb.Event) error {
	switch e := e.Type.(type) {
	case *eventpb.Event_Init:
		// do nothing
	case *eventpb.Event_SendMessage:
		targets := t.NodeIDSlice(e.SendMessage.Destinations)
		m.multicastMessage(ctx, e.SendMessage.Msg, targets)
	default:
		return fmt.Errorf("unexpected type of Net event: %T", e)
	}

	return nil
}

func (m *simTransportModule) multicastMessage(ctx context.Context, msg *messagepb.Message, targets []t.NodeID) {
	for _, target := range targets {
		m.sendMessage(ctx, msg, target)
	}
}

func (m *simTransportModule) sendMessage(ctx context.Context, msg *messagepb.Message, target t.NodeID) {
	proc := m.Spawn()

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			proc.Kill()
		case <-done:
		}
	}()

	go func() {
		defer close(done)

		if !proc.Delay(m.delayFn(m.id, target)) {
			return
		}

		destModule := m.SimTransport.nodes[t.NodeID(target)]
		proc.Send(destModule.simChan, message{m.id, t.NodeID(target), msg})

		proc.Exit()
	}()
}

func (m *simTransportModule) EventsOut() <-chan *events.EventList {
	return m.outChan
}

// func (m *simTransportModule) handleInChan(proc *testsim.Process) {
// 	for {
// 		select {
// 		case eventsIn := <-m.inChan:
// 			it := eventsIn.Iterator()
// 			for e := it.Next(); e != nil; e = it.Next() {
// 				switch e := e.Type.(type) {
// 				case *eventpb.Event_Init:
// 					// do nothing
// 				case *eventpb.Event_SendMessage:
// 					for _, dest := range e.SendMessage.Destinations {
// 						destModule := m.SimTransport.nodes[t.NodeID(dest)]
// 						msg := message{m.id, t.NodeID(dest), e.SendMessage.Msg}
// 						proc.Send(destModule.simChan, msg)
// 					}
// 					m.errChan <- nil
// 				default:
// 					m.errChan <- fmt.Errorf("unexpected type of Net event: %T", e)
// 				}
// 			}
// 		case <-m.SimTransport.stopChan:
// 			return
// 		}
// 	}
// }

func (m *simTransportModule) handleOutChan(proc *testsim.Process) {
	defer proc.Exit()

	for {
		v, ok := proc.Recv(m.simChan)
		if !ok {
			panic("Module simulation process died")
		}
		msg := v.(message)

		// ok = proc.Delay(m.SimTransport.delayFn(msg.from, msg.to))
		// if !ok {
		// 	return // simulation stopped
		// }

		destModule := t.ModuleID(msg.message.DestModule)
		e := events.MessageReceived(destModule, msg.from, msg.message)

		m.SimNode.SendEvents(proc, events.ListOf(e))
		select {
		case m.outChan <- events.ListOf(e):
			// m.SimNode.SendEvent(proc, e)
		case <-m.SimTransport.stopChan:
			return
		}
	}
}

type message struct {
	from    t.NodeID
	to      t.NodeID
	message *messagepb.Message
}
