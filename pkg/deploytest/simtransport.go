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
	*testsim.Runtime
	delayFn  MessageDelayFn
	nodes    map[t.NodeID]*simTransportModule
	stopChan chan struct{}
}

func NewSimTransport(r *testsim.Runtime, delayFn MessageDelayFn) *SimTransport {
	return &SimTransport{
		Runtime:  r,
		delayFn:  delayFn,
		nodes:    make(map[t.NodeID]*simTransportModule),
		stopChan: make(chan struct{}),
	}
}

func (t *SimTransport) NodeModule(id t.NodeID, node *SimNode) modules.ActiveModule {
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
	*SimNode
	id      t.NodeID
	outChan chan *events.EventList
	simChan *testsim.Chan
}

func (t *SimTransport) newModule(id t.NodeID, node *SimNode) *simTransportModule {
	m := &simTransportModule{
		SimTransport: t,
		SimNode:      node,
		id:           id,
		outChan:      make(chan *events.EventList, 1),
		simChan:      testsim.NewChan(),
	}

	//go m.handleInChan(t.Simulation.Spawn())
	go m.handleOutChan(t.Runtime.Spawn())

	return m
}

func (m *simTransportModule) ImplementsModule() {}

func (m *simTransportModule) ApplyEvents(ctx context.Context, eventList *events.EventList) error {
	_, err := modules.ApplyEventsSequentially(eventList, func(e *eventpb.Event) (*events.EventList, error) {
		return events.EmptyList(), m.applyEvent(ctx, e)
	})
	return err
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
		case <-m.SimTransport.stopChan:
		case <-done:
			return
		}
		proc.Kill()
	}()

	d := m.SimTransport.delayFn(m.id, target)
	go func() {
		defer close(done)

		if !proc.Delay(d) {
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

func (m *simTransportModule) handleOutChan(proc *testsim.Process) {
	go func() {
		<-m.SimTransport.stopChan
		proc.Kill()
	}()

	for {
		v, ok := proc.Recv(m.simChan)
		if !ok {
			return
		}
		msg := v.(message)

		destModule := t.ModuleID(msg.message.DestModule)
		eventList := events.ListOf(events.MessageReceived(destModule, msg.from, msg.message))

		select {
		case eventsOut := <-m.outChan:
			eventsOut.PushBackList(eventList)
			m.outChan <- eventsOut
		default:
			m.outChan <- eventList
		}
		m.SimNode.SendEvents(proc, eventList)
	}
}

type message struct {
	from    t.NodeID
	to      t.NodeID
	message *messagepb.Message
}
