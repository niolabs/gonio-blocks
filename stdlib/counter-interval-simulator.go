package stdlib

import (
	"context"
	"encoding/json"
	"time"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/props"
)

type CounterIntervalSimulatorBlock struct {
	nio.Producer
	Config CounterIntervalSimulatorConfig

	duration time.Duration
	limit    int64
	total    int64
	count    int64

	start int64
	end   int64
	step  int64

	key string
}

type CounterIntervalSimulatorConfig struct {
	nio.BlockConfigAtom

	Interval props.TimeDeltaProperty `json:"interval"`
	Limit    *props.IntProperty      `json:"limit"`
	Count    *props.IntProperty      `json:"num_signals"`

	Key   *props.StringProperty `json:"attr_name"`
	Range struct {
		Start *props.IntProperty `json:"start"`
		End   *props.IntProperty `json:"end"`
		Step  *props.IntProperty `json:"step"`
	} `json:"attr_value"`
}

func (cis *CounterIntervalSimulatorBlock) Configure(config nio.RawBlockConfig) error {
	cis.Producer.Configure()

	if err := json.Unmarshal(config, &cis.Config); err != nil {
		return err
	}

	if err := cis.Config.Interval.Assign(&cis.duration, nil); err != nil {
		return err
	}

	if err := cis.Config.Count.AssignToDefault(&cis.count, nil, 1); err != nil {
		return err
	}

	if err := cis.Config.Limit.AssignToDefault(&cis.limit, nil, 0); err != nil {
		return err
	}

	if err := cis.Config.Range.Start.AssignToDefault(&cis.start, nil, 0); err != nil {
		return err
	}

	if err := cis.Config.Range.End.AssignToDefault(&cis.end, nil, 1); err != nil {
		return err
	}

	if err := cis.Config.Range.Step.AssignToDefault(&cis.step, nil, 1); err != nil {
		return err
	}

	if err := cis.Config.Key.AssignToDefault(&cis.key, nil, "sim"); err != nil {
		return err
	}
	return nil
}

func (cis *CounterIntervalSimulatorBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return cis.NoEnqueue(terminal)
}

func (cis *CounterIntervalSimulatorBlock) Start(ctx context.Context) {
	counter := int64(0)
	t := time.NewTicker(cis.duration)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			num := cis.count
			hasLimit := cis.limit > 0
			isComplete := hasLimit && (cis.total+num) > cis.limit
			if isComplete {
				num = cis.limit - cis.total
			}

			cis.total += num

			outSignals := make(nio.SignalGroup, num)

			for i := int64(0); i < num; i++ {
				outSignals[i] = nio.Signal{cis.key: counter + cis.start}
				counter += cis.step
				switch {
				case cis.step < 0 && counter < (cis.end-cis.start):
					counter = 0
				case cis.step > 0 && counter > (cis.end-cis.start):
					counter = 0
				}
			}

			cis.ChOut <- outSignals

			if isComplete {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func newCounterIntervalSimulatorBlock() nio.Block {
	return &CounterIntervalSimulatorBlock{}
}

var CounterIntervalSimulator = nio.BlockTypeEntry{
	Create: newCounterIntervalSimulatorBlock,
	Definition: nio.BlockTypeDefinition{
		Namespace: "goblocks.simulator.blocks.CounterIntervalSimulator",
		Commands:  map[nio.Command]nio.CommandDefinition{},
		Version:   "1.3.0",
		Name:      "CounterIntervalSimulator",
		Properties: map[nio.Property]nio.PropertyDefinition{
			"version": {
				"title":      "Version",
				"order":      nil,
				"type":       "StringType",
				"default":    "1.2.0",
				"advanced":   true,
				"allow_none": false,
				"visible":    true,
			},
			"name": {
				"title":      "Name",
				"order":      nil,
				"type":       "StringType",
				"default":    nil,
				"advanced":   false,
				"allow_none": true,
				"visible":    false,
			},
			"interval": {
				"title": "Interval",
				"order": nil,
				"type":  "TimeDeltaType",
				"default": map[string]float64{
					"seconds": 1,
				},
				"advanced":   false,
				"allow_none": false,
				"visible":    true,
			},
			"id": {
				"title":      "Id",
				"order":      nil,
				"type":       "StringType",
				"default":    nil,
				"advanced":   false,
				"allow_none": false,
				"visible":    false,
			},
			"type": {
				"title":      "Type",
				"readonly":   true,
				"visible":    false,
				"advanced":   false,
				"allow_none": false,
				"order":      nil,
				"type":       "StringType",
				"default":    nil,
			},
			"log_level": {
				"title":      "Log Level",
				"advanced":   true,
				"allow_none": false,
				"visible":    true,
				"enum":       "LogLevel",
				"order":      nil,
				"type":       "SelectType",
				"default":    "NOTSET",
				"options": map[string]int{
					"CRITICAL": 50,
					"DEBUG":    10,
					"ERROR":    40,
					"WARNING":  30,
					"INFO":     20,
					"NOTSET":   0,
				},
			},
			"num_signals": {
				"title":      "Number of Signals",
				"order":      nil,
				"type":       "IntType",
				"default":    1,
				"advanced":   false,
				"allow_none": false,
				"visible":    true,
			},
			"total_signals": {
				"title":      "Total Number of Signals",
				"order":      nil,
				"type":       "IntType",
				"default":    -1,
				"advanced":   false,
				"allow_none": false,
				"visible":    true,
			},
		},
		BlockAttributes: nio.BlockAttributes{
			Inputs: []nio.TerminalDefinition{},
			Outputs: []nio.TerminalDefinition{
				{
					Order:   0,
					Visible: true,
					ID:      "__default_terminal_value",
					Type:    "output",
					Default: true,
					Label:   "output",
				},
			},
		},
	},
}
