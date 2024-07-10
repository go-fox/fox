// Package context
// MIT License
//
// # Copyright (c) 2024 go-fox
// Author https://github.com/go-fox/fox
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package context

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type mergeCtx struct {
	parent1, parent2 context.Context

	done     chan struct{}
	doneMark uint32
	doneOnce sync.Once
	doneErr  error

	cancelCh   chan struct{}
	cancelOnce sync.Once
}

// Merge merges two contexts into one.
func Merge(parent1, parent2 context.Context) (context.Context, context.CancelFunc) {
	mc := &mergeCtx{
		parent1:  parent1,
		parent2:  parent2,
		done:     make(chan struct{}),
		cancelCh: make(chan struct{}),
	}
	select {
	case <-parent1.Done():
		_ = mc.finish(parent1.Err())
	case <-parent2.Done():
		_ = mc.finish(parent2.Err())
	default:
		go mc.wait()
	}
	return mc, mc.cancel
}

func (mc *mergeCtx) finish(err error) error {
	mc.doneOnce.Do(func() {
		mc.doneErr = err
		atomic.StoreUint32(&mc.doneMark, 1)
		close(mc.done)
	})
	return mc.doneErr
}

func (mc *mergeCtx) wait() {
	var err error
	select {
	case <-mc.parent1.Done():
		err = mc.parent1.Err()
	case <-mc.parent2.Done():
		err = mc.parent2.Err()
	case <-mc.cancelCh:
		err = context.Canceled
	}
	_ = mc.finish(err)
}

func (mc *mergeCtx) cancel() {
	mc.cancelOnce.Do(func() {
		close(mc.cancelCh)
	})
}

// Done implements context.Context.
func (mc *mergeCtx) Done() <-chan struct{} {
	return mc.done
}

// Err implements context.Context.
func (mc *mergeCtx) Err() error {
	if atomic.LoadUint32(&mc.doneMark) != 0 {
		return mc.doneErr
	}
	var err error
	select {
	case <-mc.parent1.Done():
		err = mc.parent1.Err()
	case <-mc.parent2.Done():
		err = mc.parent2.Err()
	case <-mc.cancelCh:
		err = context.Canceled
	default:
		return nil
	}
	return mc.finish(err)
}

// Deadline implements context.Context.
func (mc *mergeCtx) Deadline() (time.Time, bool) {
	d1, ok1 := mc.parent1.Deadline()
	d2, ok2 := mc.parent2.Deadline()
	switch {
	case !ok1:
		return d2, ok2
	case !ok2:
		return d1, ok1
	case d1.Before(d2):
		return d1, true
	default:
		return d2, true
	}
}

// Value implements context.Context.
func (mc *mergeCtx) Value(key interface{}) interface{} {
	if v := mc.parent1.Value(key); v != nil {
		return v
	}
	return mc.parent2.Value(key)
}
