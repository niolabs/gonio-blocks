package stdlib

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/mixins"
	"github.com/niolabs/gonio-framework/props"
)

type DebounceBlock struct {
	nio.Transformer
	mixins.GroupByMixin
	Config DebounceBlockConfig

	mutex      sync.Mutex
	lastNotify map[mixins.Group]time.Time
	interval   time.Duration
}

type DebounceBlockConfig struct {
	nio.BlockConfigAtom
	Interval *props.TimeDeltaProperty
}

func (b *DebounceBlock) Configure(config nio.RawBlockConfig) error {
	b.Transformer.Configure()
	if err := b.GroupByMixin.Configure(config, b.Notify); err != nil {
		return err
	}

	if err := json.Unmarshal(config, &b.Config); err != nil {
		return err
	}

	b.Config.Interval.AssignDefault(&b.interval, nil, 1*time.Second)
	b.lastNotify = map[mixins.Group]time.Time{}

	return nil
}

func (b *DebounceBlock) Start(ctx context.Context) {
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

func (b *DebounceBlock) process(group mixins.Group, notify nio.NotifyFunc, signals nio.SignalGroup) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	now := time.Now()
	prev, hasPrev := b.lastNotify[group]

	if !hasPrev || now.Sub(prev) > b.interval {
		b.lastNotify[group] = now

		last := signals[len(signals)-1]
		notify(b.TOut, nio.SignalGroup{last})
	}

	return nil
}

func (b *DebounceBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return b.Transformer.Enqueue(terminal, signals, 1)
}
