package stdlib

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/mixins"
	"github.com/niolabs/gonio-framework/props"
)

type MergeStreamsBlock struct {
	nio.Joiner
	mixins.GroupByMixin
	Config MergeStreamsBlockConfig

	once       bool
	mutex      sync.Mutex
	leftCache  map[mixins.Group]nio.Signal
	rightCache map[mixins.Group]nio.Signal
}

type MergeStreamsBlockConfig struct {
	nio.BlockConfigAtom
	Once *props.BooleanProperty `json:"notify_once"`
}

func (b *MergeStreamsBlock) Configure(config nio.RawBlockConfig) error {
	SetTerminal(&b.TInLeft, "input_1")
	SetTerminal(&b.TInRight, "input_2")
	b.Joiner.Configure()

	if err := b.GroupByMixin.Configure(config, b.Notify); err != nil {
		return err
	}

	if err := json.Unmarshal(config, &b.Config); err != nil {
		return err
	}

	if err := b.Config.Once.AssignToDefault(&b.once, nil, true); err != nil {
		return err
	}

	b.leftCache = map[mixins.Group]nio.Signal{}
	b.rightCache = map[mixins.Group]nio.Signal{}

	return nil
}

func (b *MergeStreamsBlock) Start(ctx context.Context) {
	for {
		select {
		case signals := <-b.ChInLeft:
			b.GroupByMixin.Process(signals, b.processLeft)
			b.Busy.Done()
		case signals := <-b.ChInRight:
			b.GroupByMixin.Process(signals, b.processRight)
			b.Busy.Done()
		}
	}
}

func (b *MergeStreamsBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return b.DualConsumer.Enqueue(terminal, signals, 1)
}

func (b *MergeStreamsBlock) processLeft(group mixins.Group, notify nio.NotifyFunc, inSignals nio.SignalGroup) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if rightSignal, ok := b.rightCache[group]; ok {
		if b.once {
			delete(b.rightCache, group)
			firstSignal := inSignals[0]
			notify(b.TOut, nio.SignalGroup{firstSignal.CloneWith(rightSignal)})
			return nil
		}

		var outSignals nio.SignalGroup
		for _, inSignal := range inSignals {
			outSignals = append(outSignals, inSignal.CloneWith(rightSignal))
		}

		if err := notify(b.TOut, outSignals); err != nil {
			return err
		}
	}

	last := inSignals[len(inSignals)-1]
	b.leftCache[group] = last
	return nil
}

func (b *MergeStreamsBlock) processRight(group mixins.Group, notify nio.NotifyFunc, inSignals nio.SignalGroup) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if leftSignal, ok := b.leftCache[group]; ok {
		if b.once {
			delete(b.leftCache, group)
			firstSignal := inSignals[0]
			notify(b.TOut, nio.SignalGroup{leftSignal.CloneWith(firstSignal)})
			return nil
		}

		var outSignals nio.SignalGroup
		for _, inSignal := range inSignals {
			outSignals = append(outSignals, leftSignal.CloneWith(inSignal))
		}

		if err := notify(b.TOut, outSignals); err != nil {
			return err
		}
	}

	last := inSignals[len(inSignals)-1]
	b.rightCache[group] = last
	return nil
}
