// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package testsim

import (
	"container/heap"
	"fmt"
	"math"
	"math/rand"
	"sync"
	stdtime "time"

	"github.com/filecoin-project/mir/pkg/common/time"
)

type Runtime struct {
	*rand.Rand

	// nextPID   PID
	// processes map[PID]*Process

	activeProcLock sync.Mutex
	nrActiveProcs  uint64
	nrProcs        uint64
	resumeCond     sync.Cond

	//resumeChan     chan struct{}
	startTime  time.Time
	elapsed    time.Duration
	eventQueue eventQueue
}

func NewRuntime(rnd *rand.Rand) *Runtime {
	r := &Runtime{
		Rand: rnd,
		// nextPID:   1,
		// processes: map[PID]*Process{},
		//resumeChan: make(chan struct{}, 1),
		startTime:  stdtime.Now(),
		elapsed:    0,
		eventQueue: []*event{},
	}
	r.resumeCond = *sync.NewCond(&r.activeProcLock)

	return r
}

// Spawn creates a new active process.
func (r *Runtime) Spawn() *Process {
	//atomic.AddInt64(&r.nrActiveProcs, 1)

	p := &Process{
		Runtime:  r,
		killChan: make(chan struct{}),
		// id:         r.nextPID,
	}
	r.activeProcLock.Lock()
	r.nrProcs++
	r.activeProcLock.Unlock()
	p.activateLocked()
	// r.processes[p.id] = p
	// r.nextPID++

	return p
}

func (r *Runtime) Now() time.Time { return r.startTime.Add(r.elapsed) }

func (r *Runtime) AfterFunc(d time.Duration, f func()) *time.Timer {
	t := r.newTimer(d, f)
	return &time.Timer{TimerImpl: t}
}

func (r *Runtime) NewTimer(d time.Duration) *time.Timer {
	t := r.newTimer(d, nil)
	return &time.Timer{TimerImpl: t, C: t.timeC}
}

func (r *Runtime) NewTicker(d time.Duration) *time.Ticker {
	t := r.newTicker(d)
	return &time.Ticker{TickerImpl: t, C: t.timeC}
}

func (r *Runtime) RandDuration(min, max time.Duration) time.Duration {
	return min + time.Duration(r.Int63n(int64(max-min+1)))
}

func (r *Runtime) RandExpDuration(min, mean time.Duration) time.Duration {
	max := float64(2 << 53) // maximum exactly representable positive integer
	scale := float64(mean - min)
	for {
		randExp := r.ExpFloat64()                   // random exponentially distributed value
		d := math.FMA(scale, randExp, float64(min)) // scale * randExp + min
		if d <= max {
			return time.Duration(d)
		}
	}
}

func (r *Runtime) Run() {
	r.waitQuiescence()
	for r.eventQueue.Len() > 0 {
		r.doStep()
	}
}

func (r *Runtime) RunFor(d time.Duration) {
	r.waitQuiescence()
	t := r.elapsedAfter(d)
	for r.eventQueue.Len() > 0 && r.eventQueue[0].deadline <= t {
		r.doStep()
	}
	//time.Sleep(d)
	fmt.Println("No more work")
	fmt.Println(r.eventQueue.Len(), "events in the queue")
}

func (r *Runtime) Step() (ok bool) {
	r.waitQuiescence()
	if r.eventQueue.Len() > 0 {
		r.doStep()
		ok = true
	}
	return
}

func (r *Runtime) elapsedAfter(d time.Duration) time.Duration {
	t := r.elapsed + d
	if t < 0 {
		panic("Time overflow")
	}
	return t
}

func (r *Runtime) schedule(d time.Duration, f func()) *event {
	fmt.Println("Scheduling at", r.elapsedAfter(d))
	// r.activeProcLock.Lock()
	// if r.nrActiveProcs == 0 && r.eventQueue.Len() == 0 {
	// 	panic("wtf")
	// }
	// r.activeProcLock.Unlock()

	e := &event{
		deadline: r.elapsedAfter(d),
		fn:       f,
	}
	heap.Push(&r.eventQueue, e)
	return e
}

func (r *Runtime) doStep() {
	e := heap.Pop(&r.eventQueue).(*event)
	r.elapsed = e.deadline
	fmt.Println("Calling scheduled at", e.deadline)
	e.fn()
	r.waitQuiescence()
}

func (r *Runtime) waitQuiescence() {
	// for atomic.LoadInt64(&r.nrActiveProcs) > 0 {
	// 	fmt.Println("Waiting for quiescence")
	// 	<-r.resumeChan
	// 	fmt.Println("Resuming simulation")
	// }
	r.activeProcLock.Lock()
	for r.nrActiveProcs > 0 {
		fmt.Println("Waiting for quiescence")
		r.resumeCond.Wait()
		fmt.Println("Resuming simulation")
	}
	r.activeProcLock.Unlock()
}

func (r *Runtime) addActiveProcess() {
	fmt.Print("addActiveProcess ")
	//	fmt.Println(atomic.AddInt64(&r.nrActiveProcs, 1))
	r.activeProcLock.Lock()
	r.nrActiveProcs++
	fmt.Println(r.nrActiveProcs)
	r.activeProcLock.Unlock()
}

func (r *Runtime) rmActiveProcess() {
	fmt.Print("rmActiveProcess ")
	//x := atomic.AddInt64(&r.nrActiveProcs, -1)
	r.activeProcLock.Lock()
	r.nrActiveProcs--
	x := r.nrActiveProcs
	fmt.Println(x)
	if x == 0 {
		// if r.eventQueue.Len() == 0 {
		// 	panic("wtf")
		// }
		fmt.Println("Signal to resume simulation")
		// select {
		// case r.resumeChan <- struct{}{}:
		// default:
		// }
		r.resumeCond.Broadcast()
	}
	r.activeProcLock.Unlock()
}

func (r *Runtime) unschedule(e *event) {
	if e.index >= 0 {
		heap.Remove(&r.eventQueue, e.index)
		e.index = -1
	}
}

type event struct {
	index    int
	deadline time.Duration
	fn       func()
}

type eventQueue []*event

func (eq eventQueue) Len() int           { return len(eq) }
func (eq eventQueue) Less(i, j int) bool { return eq[i].deadline < eq[j].deadline }

func (eq eventQueue) Swap(i, j int) {
	eq[i], eq[j] = eq[j], eq[i]
	eq[i].index = i
	eq[j].index = j
}

func (eq *eventQueue) Push(v any) {
	e := v.(*event)
	e.index = len(*eq)
	*eq = append(*eq, e)
}

func (eq *eventQueue) Pop() any {
	n := len(*eq)
	e := (*eq)[n-1]
	*eq, (*eq)[n-1] = (*eq)[:n-1], nil
	e.index = -1
	return e
}

// type PID int

type Process struct {
	*Runtime
	lock     sync.Mutex
	active   bool
	killed   bool
	killChan chan struct{}
	// id PID
}

// ID returns the process' identifier.
// func (p *Process) ID() PID { return p.id }

// Fork creates a new active process.
func (p *Process) Fork() *Process { return p.Spawn() }

// // Yield enables another waiting process to run.
// func (p *Process) Yield() { p.Delay(0) }

// Delay suspends the execution and returns after d amount of
// simulated time. It returns false in case the process was killed
// while waiting.
func (p *Process) Delay(d time.Duration) (ok bool) {
	// time.Sleep(d)

	done := make(chan struct{})
	e := p.Runtime.schedule(d, func() {
		p.activate()
		close(done)
	})
	p.deactivate()

	select {
	case <-done:
		return true
	case <-p.killChan:
		p.Runtime.unschedule(e)
		return false
	}
}

// Send enqueues the value v to the channel and resumes a process
// waiting in Recv operation on the channel.
func (p *Process) Send(c *Chan, v any) (ok bool) {
	c.lock.Lock()
	if c.waiter != nil {
		c.waiter.activate()
		c.waiter = nil
		//c.ch <- v
	} else {
		p.deactivate()
		c.waiter = p
	}
	c.lock.Unlock()

	// select {
	// case c.ch <- v:
	// 	return true
	// default:
	// 	p.deactivate()
	// }

	select {
	case c.ch <- v:
		//p.activate()
		return true
	case <-p.killChan:
		return false
	}
}

// Recv attempts to dequeue a value from the channel. If the channel
// is empty then it suspends the execution until a new value is
// enqueued to the channel or the process gets killed. It returns the
// dequeued value and true, otherwise it returns nil and false in case
// the process was killed while waiting.
func (p *Process) Recv(c *Chan) (v any, ok bool) {
	c.lock.Lock()
	if c.waiter != nil {
		c.waiter.activate()
		c.waiter = nil
	} else {
		p.deactivate()
		c.waiter = p
	}
	c.lock.Unlock()

	// select {
	// case v = <-c.ch:
	// 	return v, true
	// default:
	// 	p.deactivate()
	// }

	select {
	case v = <-c.ch:
		//p.activate()
		return v, true
	case <-p.killChan:
		return nil, false
	}
}

// Exit terminates the process normally.
func (p *Process) Exit() {
	p.deactivate()
	p.Runtime.activeProcLock.Lock()
	p.Runtime.nrProcs--
	p.Runtime.activeProcLock.Unlock()
}

// Kill immediately resumes and terminates the process.
func (p *Process) Kill() {
	p.Runtime.activeProcLock.Lock()
	p.Runtime.nrProcs--
	p.Runtime.activeProcLock.Unlock()
	p.lock.Lock()
	p.deactivateLocked()
	p.killed = true
	p.lock.Unlock()

	close(p.killChan)
}

func (p *Process) activate() bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.activateLocked()
}
func (p *Process) deactivate() bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.deactivateLocked()
}

func (p *Process) activateLocked() bool {
	if !p.active && !p.killed {
		p.active = true
		p.Runtime.addActiveProcess()
		return true
	}
	return false
}

func (p *Process) deactivateLocked() bool {
	if p.active {
		p.active = false
		p.Runtime.rmActiveProcess()
		return true
	}
	return false
}

type Chan struct {
	lock   sync.Mutex
	ch     chan any
	waiter *Process
}

func NewChan() *Chan {
	return &Chan{
		ch: make(chan any),
	}
}
