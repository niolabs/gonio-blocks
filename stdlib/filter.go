package stdlib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/props"
)

type FilterBlock struct {
	nio.Splitter
	Config FilterBlockConfig

	operator string
}

type FilterBlockConfig struct {
	nio.BlockConfigAtom
	Operator   *props.StringProperty `json:"operator"`
	Conditions []struct {
		Expr props.BooleanProperty `json:"expr"`
	} `json:"conditions"`
}

func (fb *FilterBlock) Configure(config nio.RawBlockConfig) error {
	SetTerminal(&fb.TOutLeft, "true")
	SetTerminal(&fb.TOutRight, "false")

	fb.Splitter.Configure()

	if err := json.Unmarshal(config, &fb.Config); err != nil {
		return err
	}

	if len(fb.Config.Conditions) == 0 {
		return errors.New("configuration error: no conditions")
	}

	if prop := fb.Config.Operator; prop == nil {
		fb.operator = "ALL"
	} else {
		switch op, err := prop.Invoke(nil); {
		case err != nil:
			return err
		case op == "ALL":
			fb.operator = "ALL"
		case op == "ANY":
			fb.operator = "ANY"
		default:
			return fmt.Errorf("configuration error: invalid operation `%s'", op)
		}
	}

	return nil
}

func (fb *FilterBlock) Start(ctx context.Context) {
	for {
		select {
		case signals := <-fb.ChIn:
			fb.process(signals)
			fb.Busy.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (fb *FilterBlock) Enqueue(t nio.Terminal, signals nio.SignalGroup) error {
	return fb.Consumer.Enqueue(t, signals, 1)
}

func (fb *FilterBlock) process(signals nio.SignalGroup) {
	total := len(signals)

	trueSignals := make(nio.SignalGroup, 0, total)
	falseSignals := make(nio.SignalGroup, 0, total)

SignalLoop:
	for _, signal := range signals {
		var test bool

		switch fb.operator {
		case "ALL":
			test = true
		case "ANY":
			test = false
		default:
			// TODO handle invalid operator
			panic("invalid operator")
		}

		for _, prop := range fb.Config.Conditions {
			b, err := prop.Expr.Invoke(signal)
			if err != nil {
				// TODO handle expr eval error
				continue SignalLoop
			}

			switch fb.operator {
			case "ALL":
				test = test && b
			case "ANY":
				test = test || b
			}
		}

		switch test {
		case true:
			trueSignals = append(trueSignals, signal)
		case false:
			falseSignals = append(falseSignals, signal)
		}
	}

	// notify in non-deterministic order
	for ch, outSignals := range map[chan nio.SignalGroup]nio.SignalGroup{
		fb.ChOutLeft:  trueSignals,
		fb.ChOutRight: falseSignals,
	} {
		if len(outSignals) > 0 && ch != nil {
			ch <- outSignals
		}
	}
}
