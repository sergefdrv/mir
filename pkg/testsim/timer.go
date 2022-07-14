// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package testsim

import (
	"github.com/filecoin-project/mir/pkg/common/time"
)

type timer struct {
	*Runtime
	*event
	fn     func()
	timeC  chan time.Time
	active bool
}

func (r *Runtime) newTimer(d time.Duration, f func()) (t *timer) {
	t = &timer{Runtime: r}
	if f != nil {
		t.fn = f
	} else {
		t.timeC = make(chan time.Time, 1)
	}
	t.schedule(d)

	return t
}

func (t *timer) schedule(d time.Duration) {
	r := t.Runtime
	t.event = r.schedule(d, func() {
		t.active = false
		if t.fn != nil {
			t.fn()
		} else {
			t.timeC <- r.Now()
		}
	})
	t.active = true
}

func (t *timer) cancel() {
	t.Runtime.unschedule(t.event)
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

func (r *Runtime) newTicker(d time.Duration) (t *ticker) {
	return &ticker{*r.newTimer(d, nil)}
}

func (t *ticker) Reset(d time.Duration) { t.timer.Reset(d) }
func (t *ticker) Stop()                 { t.timer.Stop() }
