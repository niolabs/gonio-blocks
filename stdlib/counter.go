package stdlib

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/mixins"
)

type CounterBlock struct {
	nio.Transformer
	mixins.GroupByMixin
	Config nio.BlockConfigAtom

	cumulativeCount map[mixins.Group]int
	mutex           sync.RWMutex
}

func (b *CounterBlock) Configure(config nio.RawBlockConfig) error {
	b.Transformer.Configure()

	if err := json.Unmarshal(config, &b.Config); err != nil {
		return err
	}

	if err := b.GroupByMixin.Configure(config, b.Notify); err != nil {
		return err
	}

	b.cumulativeCount = map[mixins.Group]int{}

	return nil
}

func (b *CounterBlock) Start(ctx context.Context) {
	for {
		select {
		case signals := <-b.ChIn:
			b.GroupByMixin.Process(signals, b.process)
			b.Busy.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (b *CounterBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return b.Transformer.Enqueue(terminal, signals, 1)
}

func (b *CounterBlock) process(group mixins.Group, notify nio.NotifyFunc, signals nio.SignalGroup) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	c := len(signals)
	next := b.cumulativeCount[group] + c
	b.cumulativeCount[group] = next

	outSignal := nio.Signal{
		"count":            c,
		"cumulative_count": next,
	}

	b.GroupByMixin.AddGroupToSignal(group, outSignal, false)
	return notify(b.TOut, nio.SignalGroup{outSignal})
}
