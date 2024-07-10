package wrr

import (
	"context"
	"sync"

	"github.com/go-fox/fox/selector"
	"github.com/go-fox/fox/selector/base"
	"github.com/go-fox/fox/selector/node/direct"
)

// Name selector name
const Name = "wrr"

var _ selector.Balancer = (*balancer)(nil)
var _ selector.BalancerBuilder = (*balancerBuilder)(nil)

func init() {
	selector.Register(
		base.NewSelectorBuilder(
			Name,
			&direct.Builder{},
			&balancerBuilder{},
		),
	)
}

type balancer struct {
	mu            sync.Mutex
	currentWeight map[string]float64
}

func (b *balancer) Pick(ctx context.Context, nodes []selector.WeightedNode) (selector.WeightedNode, selector.DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, selector.ErrNoAvailable
	}
	var totalWeight float64
	var selected selector.WeightedNode
	var selectWeight float64

	// nginx wrr load balancing algorithm: http://blog.csdn.net/zhangskd/article/details/50194069
	b.mu.Lock()
	for _, node := range nodes {
		totalWeight += node.Weight()
		cwt := b.currentWeight[node.Address()]
		// current += effectiveWeight
		cwt += node.Weight()
		b.currentWeight[node.Address()] = cwt
		if selected == nil || selectWeight < cwt {
			selectWeight = cwt
			selected = node
		}
	}
	b.currentWeight[selected.Address()] = selectWeight - totalWeight
	b.mu.Unlock()

	d := selected.Pick()
	return selected, d, nil
}

type balancerBuilder struct{}

func (b *balancerBuilder) Build() selector.Balancer {
	return &balancer{}
}
