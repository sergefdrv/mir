// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package testsim

import (
	"github.com/filecoin-project/mir/pkg/common/time"
)

type timer struct {
	*Simulation
	*event
	fn     func()
	timeC  chan time.Time
	active bool
}

func (s *Simulation) newTimer(d time.Duration, f func()) (t *timer) {
	t = &timer{Simulation: s}
	if f != nil {
		t.fn = f
	} else {
		t.timeC = make(chan time.Time, 1)
	}
	t.schedule(d)

	return t
}

func (t *timer) schedule(d time.Duration) {
	s := t.Simulation
	t.event = s.newEvent(d, func() {
		t.active = false
		if t.fn != nil {
			t.fn()
		} else {
			t.timeC <- s.Now()
		}
	})
	t.active = true
}

func (t *timer) cancel() {
	t.Simulation.removeEvent(t.event)
	t.event = nil
	t.active = false
}

func (t *timer) Reset(d time.Duration) (ok bool) {
	ok = t.Stop()
	t.schedule(d)
	return ok
}

func (t *timer) Stop() (ok bool) {
	ok = t.active
	t.cancel()
	return ok
}

type ticker struct {
	timer
}

func (s *Simulation) newTicker(d time.Duration) (t *ticker) {
	return &ticker{*s.newTimer(d, nil)}
}

func (t *ticker) Reset(d time.Duration) { t.timer.Reset(d) }
func (t *ticker) Stop()                 { t.timer.Stop() }
