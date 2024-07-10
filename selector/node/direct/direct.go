// Package direct
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
package direct

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/go-fox/fox/selector"
)

const (
	defaultWeight = 100
)

var (
	_ selector.WeightedNode        = (*Node)(nil)
	_ selector.WeightedNodeBuilder = (*Builder)(nil)
)

// Node impl selector.WeightedNode
type Node struct {
	selector.Node
	// last lastPick timestamp
	lastPick int64
}

// Raw is original node
func (d *Node) Raw() selector.Node {
	return d.Node
}

// Weight is node effective weight
func (d *Node) Weight() float64 {
	if d.InitialWeight() != nil {
		return float64(*d.InitialWeight())
	}
	return defaultWeight
}

// Pick is pick a node
func (d *Node) Pick() selector.DoneFunc {
	now := time.Now().UnixNano()
	atomic.StoreInt64(&d.lastPick, now)
	return func(ctx context.Context, di selector.DoneInfo) {}
}

// PickElapsed select node time
func (d *Node) PickElapsed() time.Duration {
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&d.lastPick))
}

// Builder impl selector.WeightedNodeBuilder
type Builder struct{}

// Build is create weight node
func (b *Builder) Build(node selector.Node) selector.WeightedNode {
	return &Node{Node: node, lastPick: 0}
}
