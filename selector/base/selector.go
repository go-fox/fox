// Package base
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
package base

import (
	"context"

	"github.com/go-fox/sugar/container/satomic"

	"github.com/go-fox/fox/selector"
)

var (
	_ selector.Selector = (*baseSelector)(nil)
	_ selector.Builder  = (*baseBuilder)(nil)
)

type baseSelector struct {
	name        string
	nodeBuilder selector.WeightedNodeBuilder
	balancer    selector.Balancer
	nodes       *satomic.Value[[]selector.WeightedNode]
}

func (b *baseSelector) Store(nodes []selector.Node) {
	weightedNodes := make([]selector.WeightedNode, 0, len(nodes))
	for _, n := range nodes {
		weightedNodes = append(weightedNodes, b.nodeBuilder.Build(n))
	}
	b.nodes.Store(weightedNodes)
}

func (b *baseSelector) Select(ctx context.Context, opts ...selector.SelectOption) (selected selector.Node, done selector.DoneFunc, err error) {
	var (
		options    selector.SelectOptions
		candidates []selector.WeightedNode
	)
	nodes := b.nodes.Load()
	for _, o := range opts {
		o(&options)
	}
	if len(options.NodeFilters) > 0 {
		newNodes := make([]selector.Node, len(nodes))
		for i, wc := range nodes {
			newNodes[i] = wc
		}
		for _, filter := range options.NodeFilters {
			newNodes = filter(ctx, newNodes)
		}
		candidates = make([]selector.WeightedNode, len(newNodes))
		for i, n := range newNodes {
			candidates[i] = n.(selector.WeightedNode)
		}
	} else {
		candidates = nodes
	}

	if len(candidates) == 0 {
		return nil, nil, selector.ErrNoAvailable
	}
	wn, done, err := b.balancer.Pick(ctx, candidates)
	if err != nil {
		return nil, nil, err
	}
	p, ok := selector.FromPeerContext(ctx)
	if ok {
		p.Node = wn.Raw()
	}
	return wn.Raw(), done, nil
}

type baseBuilder struct {
	name        string
	nodeBuilder selector.WeightedNodeBuilder
	balancer    selector.BalancerBuilder
}

func (bb *baseBuilder) Name() string {
	return bb.name
}

func (bb *baseBuilder) Build() selector.Selector {
	return &baseSelector{
		nodeBuilder: bb.nodeBuilder,
		balancer:    bb.balancer.Build(),
		name:        bb.name,
		nodes:       satomic.New[[]selector.WeightedNode](),
	}
}
