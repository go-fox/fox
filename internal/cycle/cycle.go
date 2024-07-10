// Package cycle
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
package cycle

import (
	"sync"
	"sync/atomic"
)

// Cycle 生命周期管理
type Cycle struct {
	locker  *sync.Mutex
	wg      *sync.WaitGroup
	done    chan struct{}
	quit    chan error
	closing uint32
	waiting uint32
}

// NewCycle 创建
func NewCycle() *Cycle {
	return &Cycle{
		locker:  &sync.Mutex{},
		wg:      &sync.WaitGroup{},
		done:    make(chan struct{}),
		quit:    make(chan error),
		closing: 0,
		waiting: 0,
	}
}

// Run 运行
func (c *Cycle) Run(fn func() error) {
	c.locker.Lock()
	defer c.locker.Unlock()
	c.wg.Add(1)
	go func(c *Cycle) {
		defer c.wg.Done()
		if err := fn(); err != nil {
			c.quit <- err
		}
	}(c)
}

// Done 结束通道
func (c *Cycle) Done() <-chan struct{} {
	if atomic.CompareAndSwapUint32(&c.waiting, 0, 1) {
		go func(c *Cycle) {
			c.locker.Lock()
			defer c.locker.Unlock()
			c.wg.Wait()
			close(c.done)
		}(c)
	}
	return c.done
}

// Close 手动关闭
func (c *Cycle) Close() {
	c.locker.Lock()
	defer c.locker.Unlock()
	if atomic.CompareAndSwapUint32(&c.closing, 0, 1) {
		close(c.quit)
	}
}

// DoneAndClose 结束并关闭
func (c *Cycle) DoneAndClose() {
	<-c.Done()
	c.Close()
}

// Wait 一直等待，一直到有错误
func (c *Cycle) Wait() <-chan error {
	return c.quit
}
