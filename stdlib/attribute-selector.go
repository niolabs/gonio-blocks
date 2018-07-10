package stdlib

import (
	"context"
	"encoding/json"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/props"
)

type AttributeSelectorBlock struct {
	nio.Transformer
	Config AttributeSelectorBlockConfig
}

type AttributeSelectorBlockConfig struct {
	nio.BlockConfigAtom
	Mode       props.BooleanProperty    `json:"mode"`
	Attributes props.StringPropertyList `json:"attributes"`
}

func (b *AttributeSelectorBlock) Configure(config nio.RawBlockConfig) error {
	b.Transformer.Configure()
	if err := json.Unmarshal(config, &b.Config); err != nil {
		return err
	}
	return nil
}

func (b *AttributeSelectorBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return b.Consumer.Enqueue(terminal, signals, 1)
}

func (b *AttributeSelectorBlock) Start(ctx context.Context) {
	for {
		select {
		case signals := <-b.ChIn:
			b.process(signals)
		case <-ctx.Done():
			return
		}
	}
}

func (b *AttributeSelectorBlock) process(inSignals nio.SignalGroup) {
	defer b.Busy.Done()

	outSignals := make(nio.SignalGroup, 0, len(inSignals))

SignalLoop:
	for _, signal := range inSignals {
		mode, err := b.Config.Mode.Invoke(signal)
		if err != nil {
			// TODO handle error?
			continue SignalLoop
		}

		selection := map[string]struct{}{}
		for _, prop := range b.Config.Attributes {
			attr, err := prop.Invoke(signal)
			if err != nil {
				// TODO handle error?
				continue SignalLoop
			}
			selection[attr] = struct{}{}
		}

		// whitelist
		outSignal := nio.Signal{}
		for k, v := range signal {
			_, ok := selection[k]
			if mode == ok {
				outSignal[k] = v
			}
		}
		outSignals = append(outSignals, outSignal)
	}
	b.ChOut <- outSignals
}
