// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package time

import "time"

type Time = time.Time
type Duration = time.Duration

var impl Impl = StandardImpl

func SetImpl(i Impl) { impl = i }

type Impl interface {
	Now() Time
	AfterFunc(d Duration, f func()) *Timer
	NewTimer(d Duration) *Timer
	NewTicker(d Duration) *Ticker
}

type TimerImpl interface {
	Reset(d Duration) bool
	Stop() bool
}

type Timer struct {
	TimerImpl
	C <-chan Time
}

type TickerImpl interface {
	Reset(d Duration)
	Stop()
}

type Ticker struct {
	TickerImpl
	C <-chan Time
}

func Now() Time                             { return impl.Now() }
func AfterFunc(d Duration, f func()) *Timer { return impl.AfterFunc(d, f) }
func NewTimer(d Duration) *Timer            { return impl.NewTimer(d) }
func NewTicker(d Duration) *Ticker          { return impl.NewTicker(d) }

func After(d Duration) <-chan Time { return impl.NewTimer(d).C }
func Sleep(d Duration)             { <-impl.NewTimer(d).C }
func Tick(d Duration) <-chan Time  { return impl.NewTicker(d).C }
func Since(t Time) Duration        { return impl.Now().Sub(t) }
func Until(t Time) Duration        { return t.Sub(impl.Now()) }
