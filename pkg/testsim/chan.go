// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package testsim

import "sync"

// Chan represents a channel for synchronization and communication
// between processes in the simulated runtime.
type Chan struct {
	lock   sync.Mutex
	ch     chan any
	waiter *Process
}

// NewChan creates a new channel in the simulation runtime.
func NewChan() *Chan {
	return &Chan{
		ch: make(chan any),
	}
}

// send performs the Send operation by the process on the channel.
func (c *Chan) send(p *Process, v any) (ok bool) {
	c.visit(p)

	select {
	case c.ch <- v:
		return true
	case <-p.killChan:
		return false
	case <-p.Runtime.stopChan:
		return false
	}
}

// recv performs the Recv operation by the process on the channel.
func (c *Chan) recv(p *Process) (v any, ok bool) {
	c.visit(p)

	select {
	case v = <-c.ch:
		return v, true
	case <-p.killChan:
		return nil, false
	case <-p.Runtime.stopChan:
		return nil, false
	}
}

// visit prepares the channel and the process for an operation being
// performed by the process on the channel.
func (c *Chan) visit(p *Process) {
	c.lock.Lock()
	if c.waiter != nil {
		c.waiter.activate()
		c.waiter = nil
	} else {
		p.deactivate()
		c.waiter = p
	}
	c.lock.Unlock()
}
