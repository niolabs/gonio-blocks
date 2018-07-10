package stdlib

import (
	"context"
	"encoding/json"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/props"
)

type ModifierBlock struct {
	nio.Transformer
	Config ModifierBlockConfig
}

type ModifierBlockConfig struct {
	nio.BlockConfigAtom
	Exclude props.BooleanProperty `json:"exclude"`
	Fields  []struct {
		Title   props.StringProperty `json:"title"`
		Formula *props.AnyProperty   `json:"formula"`
	} `json:"fields"`
}

func (b *ModifierBlock) Configure(config nio.RawBlockConfig) error {
	b.Transformer.Configure()
	if err := json.Unmarshal(config, &b.Config); err != nil {
		return err
	}
	return nil
}

func (b *ModifierBlock) Start(ctx context.Context) {
	for {
		select {
		case inSignals := <-b.ChIn:
			var outSignals nio.SignalGroup

			for _, inSignal := range inSignals {
				next := nio.Signal{}

				if exclude, excludeErr := b.Config.Exclude.InvokeDefault(inSignal, false); excludeErr != nil {
					// TODO handle error
					next = inSignal
					goto appendSignal
				} else if exclude {
					next = nio.Signal{}
				} else {
					next = inSignal.Clone()
				}

				for _, field := range b.Config.Fields {
					key, keyErr := field.Title.Invoke(inSignal)
					if keyErr != nil {
						// TODO handle error
						goto appendSignal
					}

					value, valueErr := field.Formula.InvokeDefault(inSignal, nil)
					if valueErr != nil {
						// TODO handle error
						goto appendSignal
					}

					next[key] = value
				}

			appendSignal:
				outSignals = append(outSignals, next)
			}

			b.ChOut <- outSignals
			b.Busy.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (b *ModifierBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return b.Transformer.Enqueue(terminal, signals, 1)
}
