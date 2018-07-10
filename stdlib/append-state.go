package stdlib

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/mixins"
	"github.com/niolabs/gonio-framework/props"
)

type AppendStateBlock struct {
	nio.Joiner
	Config AppendStateBlockConfig
	mixins.GroupByMixin

	initialState interface{}
	key          string

	previousState map[mixins.Group]interface{}
	mutex         sync.RWMutex
}

type AppendStateBlockConfig struct {
	nio.BlockConfigAtom
	InitialState *props.AnyProperty    `json:"initial_state"`
	StateExpr    *props.AnyProperty    `json:"state_expr"`
	StateName    *props.StringProperty `json:"state_name"`
}

func (b *AppendStateBlock) Configure(config nio.RawBlockConfig) error {
	SetTerminal(&b.TInLeft, "getter")
	SetTerminal(&b.TInRight, "setter")

	b.Joiner.Configure()
	if err := b.GroupByMixin.Configure(config, b.Notify); err != nil {
		return err
	}

	if err := json.Unmarshal(config, &b.Config); err != nil {
		return err
	}

	if b.Config.StateExpr == nil {
		panic("state expression is unset")
	}

	if err := b.Config.InitialState.AssignToDefault(&b.initialState, nil, nil); err != nil {
		return err
	}

	if err := b.Config.StateName.AssignToDefault(&b.key, nil, "state"); err != nil {
		return err
	}

	b.previousState = map[mixins.Group]interface{}{}

	return nil
}

func (b *AppendStateBlock) Start(ctx context.Context) {
	for {
		select {
		case signals := <-b.ChInLeft:
			b.GroupByMixin.Process(signals, b.processGetter)
			b.Busy.Done()
		case signals := <-b.ChInRight:
			b.GroupByMixin.Process(signals, b.processSetter)
			b.Busy.Done()
		}
	}
}

func (b *AppendStateBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return b.Joiner.Enqueue(terminal, signals, 1)
}

func (b *AppendStateBlock) processGetter(group mixins.Group, notify nio.NotifyFunc, inSignals nio.SignalGroup) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	state, ok := b.previousState[group]
	if !ok {
		state = b.initialState
	}

	var outSignals nio.SignalGroup
	for _, inSignal := range inSignals {
		outSignal := inSignal.Clone()
		outSignal[b.key] = state
		outSignals = append(outSignals, outSignal)
	}

	return notify(b.TOut, outSignals)
}

func (b *AppendStateBlock) processSetter(group mixins.Group, notify nio.NotifyFunc, signals nio.SignalGroup) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	last := signals[len(signals)-1]

	value, err := b.Config.StateExpr.Invoke(last)
	if err != nil {
		return err
	}

	b.previousState[group] = value

	return nil
}
