package stdlib

import (
	"context"
	"sync"
	"encoding/json"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/mixins"
	"github.com/niolabs/gonio-framework/props"
)

type SwitchBlock struct {
	nio.DualTransformer
	Config SwitchBlockConfig
	mixins.GroupByMixin

	initialState bool
	switchState  map[mixins.Group]bool
	mutex        sync.RWMutex
}

type SwitchBlockConfig struct {
	nio.BlockConfigAtom
	StateExpr    *props.BooleanProperty `json:"state_expr"`
	InitialState *props.BooleanProperty `json:"initial_state"`
}

func (b *SwitchBlock) Configure(config nio.RawBlockConfig) error {
	SetTerminal(&b.TInLeft, "getter")
	SetTerminal(&b.TInRight, "setter")
	SetTerminal(&b.TOutLeft, "true")
	SetTerminal(&b.TOutRight, "false")

	b.DualTransformer.Configure()

	if err := json.Unmarshal(config, &b.Config); err != nil {
		return err
	}

	if err := b.GroupByMixin.Configure(config, b.Notify); err != nil {
		return err
	}

	if b.Config.StateExpr == nil {
		panic("state expression is unset")
	}

	if err := b.Config.InitialState.AssignToDefault(&b.initialState, nil, false); err != nil {
		return err
	}

	b.switchState = map[mixins.Group]bool{}

	return nil
}

func (b *SwitchBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return b.DualConsumer.Enqueue(terminal, signals, 1)
}

func (b *SwitchBlock) Start(ctx context.Context) {
	for {
		select {
		case signals := <-b.ChInLeft:
			b.GroupByMixin.Process(signals, b.processGetter)
			b.Busy.Done()
		case signals := <-b.ChInRight:
			b.GroupByMixin.Process(signals, b.processSetter)
			b.Busy.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (b *SwitchBlock) processGetter(group mixins.Group, notify nio.NotifyFunc, signals nio.SignalGroup) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	state, ok := b.switchState[group]
	if !ok {
		state = b.initialState
	}

	var tOut nio.Terminal
	if state {
		tOut = b.TOutLeft
	} else {
		tOut = b.TOutRight
	}

	return notify(tOut, signals)
}

func (b *SwitchBlock) processSetter(group mixins.Group, notify nio.NotifyFunc, signals nio.SignalGroup) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var err error

	last := signals[len(signals)-1]
	if b.switchState[group], err = b.Config.StateExpr.Invoke(last); err != nil {
		return err
	}

	return nil
}
