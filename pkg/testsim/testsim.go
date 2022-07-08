// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package testsim

import (
	"container/heap"
	"math"
	"math/rand"
	stdtime "time"

	"github.com/filecoin-project/mir/pkg/common/time"
)

type Simulation struct {
	*rand.Rand

	// nextPID   PID
	// processes map[PID]*Process
	startTime  time.Time
	elapsed    time.Duration
	eventQueue eventQueue
	stopChan   chan struct{}
}

func NewSimulation(rnd *rand.Rand) *Simulation {
	return &Simulation{
		Rand: rnd,
		// nextPID:   1,
		// processes: map[PID]*Process{},
		startTime:  stdtime.Now(),
		elapsed:    0,
		eventQueue: []*event{},
		stopChan:   make(chan struct{}),
	}
}

// Spawn creates a new active process.
func (s *Simulation) Spawn() *Process {
	p := &Process{
		Simulation: s,
		// id:         s.nextPID,
	}
	// s.processes[p.id] = p
	// s.nextPID++

	return p
}

func (s *Simulation) Now() time.Time { return s.startTime.Add(s.elapsed) }

func (s *Simulation) AfterFunc(d time.Duration, f func()) *time.Timer {
	t := s.newTimer(d, f)
	return &time.Timer{TimerImpl: t}
}

func (s *Simulation) NewTimer(d time.Duration) *time.Timer {
	t := s.newTimer(d, nil)
	return &time.Timer{TimerImpl: t, C: t.timeC}
}

func (s *Simulation) NewTicker(d time.Duration) *time.Ticker {
	t := s.newTicker(d)
	return &time.Ticker{TickerImpl: t, C: t.timeC}
}

func (s *Simulation) RandDuration(min, max time.Duration) time.Duration {
	return min + time.Duration(s.Int63n(int64(max-min+1)))
}

func (s *Simulation) RandExpDuration(min, mean time.Duration) time.Duration {
	max := float64(2 << 53) // maximum exactly representable positive integer
	scale := float64(mean - min)
	for {
		randExp := s.ExpFloat64()                   // random exponentially distributed value
		d := math.FMA(scale, randExp, float64(min)) // scale * randExp + min
		if d <= max {
			return time.Duration(d)
		}
	}
}

func (s *Simulation) Run() {
	for s.eventQueue.Len() > 0 {
		s.doStep()
	}
}

func (s *Simulation) RunFor(d time.Duration) {
	// t := s.elapsedAfter(d)
	// for s.eventQueue.Len() > 0 && s.eventQueue[0].deadline <= t {
	// 	s.doStep()
	// }
	time.Sleep(d)
}

func (s *Simulation) Step() (ok bool) {
	if s.eventQueue.Len() > 0 {
		s.doStep()
		ok = true
	}
	return
}

func (s *Simulation) Stop() {
	close(s.stopChan)
}

func (s *Simulation) elapsedAfter(d time.Duration) time.Duration {
	t := s.elapsed + d
	if t < 0 {
		panic("Time overflow")
	}
	return t
}

func (s *Simulation) doStep() {
	e := heap.Pop(&s.eventQueue).(*event)
	s.elapsed = e.deadline
	e.fn()
}

func (s *Simulation) newEvent(d time.Duration, f func()) (e *event) {
	e = &event{
		index:    s.eventQueue.Len(),
		deadline: s.elapsedAfter(d),
		fn:       f,
	}
	heap.Push(&s.eventQueue, e)
	return
}

func (s *Simulation) removeEvent(e *event) {
	heap.Remove(&s.eventQueue, e.index)
	e.index = -1
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

func (eq *eventQueue) Push(e any) {
	*eq = append(*eq, e.(*event))
}

func (eq *eventQueue) Pop() (e any) {
	n := len(*eq)
	e = (*eq)[n-1]
	*eq, (*eq)[n-1] = (*eq)[:n-1], nil
	return
}

// type PID int

type Process struct {
	*Simulation

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
func (p *Process) Delay(d time.Duration) (ok bool) { time.Sleep(d); return true }

// Send enqueues the value v to the channel and resumes a process
// waiting in Recv operation on the channel.
func (p *Process) Send(c *Chan, v any) { c.ch <- v }

// Recv attempts to dequeue a value from the channel. If the channel
// is empty then it suspends the execution until a new value is
// enqueued to the channel or the process gets killed. It returns the
// dequeued value and true, otherwise it returns nil and false in case
// the process was killed while waiting.
func (p *Process) Recv(c *Chan) (v any, ok bool) { return <-c.ch, true }

// Exit terminates the process normally.
func (p *Process) Exit() {}

// Kill immediately resumes and terminates the process.
func (p *Process) Kill() {}

type Chan struct {
	ch chan any
}

func NewChan() *Chan {
	return &Chan{
		ch: make(chan any),
	}
}
