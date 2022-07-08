// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package time

import "time"

var StandardImpl = standardImpl{}

type standardImpl struct{}

func (standardImpl) Now() time.Time { return time.Now() }

func (standardImpl) AfterFunc(d time.Duration, f func()) *Timer {
	return &Timer{time.AfterFunc(d, f), nil}
}

func (standardImpl) NewTimer(d time.Duration) *Timer {
	t := time.NewTimer(d)
	return &Timer{t, t.C}
}

func (standardImpl) NewTicker(d time.Duration) *Ticker {
	t := time.NewTicker(d)
	return &Ticker{t, t.C}
}
