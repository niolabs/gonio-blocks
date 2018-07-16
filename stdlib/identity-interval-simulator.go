package stdlib

import (
	"context"
	"encoding/json"
	"time"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/props"
)

// IdentityIntervalSimulatorBlock
type IdentityIntervalSimulatorBlock struct {
	nio.Producer
	Config IdentityIntervalSimulatorConfig

	duration time.Duration
	limit    int64
	total    int64
	count    int64
}

type IdentityIntervalSimulatorConfig struct {
	nio.BlockConfigAtom
	Interval props.TimeDeltaProperty `json:"interval"`
	Limit    *props.IntProperty      `json:"limit"`
	Count    *props.IntProperty      `json:"num_signals"`
}

func (iis *IdentityIntervalSimulatorBlock) Configure(config nio.RawBlockConfig) error {
	iis.Producer.Configure()

	if err := json.Unmarshal(config, &iis.Config); err != nil {
		return err
	}

	if err := iis.Config.Interval.Assign(&iis.duration, nil); err != nil {
		return err
	}

	if err := iis.Config.Limit.AssignToDefault(&iis.limit, nil, -1); err != nil {
		return err
	}

	if err := iis.Config.Count.AssignToDefault(&iis.count, nil, 1); err != nil {
		return err
	}

	return nil
}

func (iis *IdentityIntervalSimulatorBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return iis.NoEnqueue(terminal)
}

func (iis *IdentityIntervalSimulatorBlock) Start(ctx context.Context) {
	t := time.NewTicker(iis.duration)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			num := iis.count
			hasLimit := iis.limit > 0
			isComplete := hasLimit && (iis.total+num) > iis.limit
			if isComplete {
				num = iis.limit - iis.total
			}

			iis.total += num
			iis.ChOut <- make(nio.SignalGroup, num)

			if isComplete {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

var IdentityIntervalSimulator = nio.BlockTypeEntry{
	Create: func() nio.Block { return &IdentityIntervalSimulatorBlock{} },
	Definition: nio.BlockTypeDefinition{
		Version: "0.2.0",
		BlockAttributes: nio.BlockAttributes{
			Outputs: []nio.TerminalDefinition{
				{
					Label:   "default",
					Type:    "output",
					Visible: true,
					Order:   0,
					ID:      "__default_terminal_value",
					Default: true,
				},
			},
			Inputs: []nio.TerminalDefinition{},
		},
		Namespace: "goblocks.simulator.blocks.IdentityIntervalSimulator",
		Properties: map[nio.Property]nio.PropertyDefinition{
			"interval": {
				"order":    0,
				"type":     "TimeDeltaType",
				"advanced": false,
				"visible":  true,
				"default": map[string]float64{
					"seconds": 1,
				},
				"allow_none": false,
				"title":      "Interval",
			},
			"num_signals": {
				"order":      3,
				"type":       "IntType",
				"advanced":   false,
				"visible":    true,
				"default":    1,
				"allow_none": false,
				"title":      "Number of Signals",
			},
			"total_signals": {
				"order":      4,
				"type":       "IntType",
				"advanced":   false,
				"visible":    true,
				"default":    -1,
				"allow_none": false,
				"title":      "Total Number of Signals",
			},
			"version": {
				"order":      nil,
				"type":       "StringType",
				"advanced":   true,
				"visible":    true,
				"default":    "0.2.0",
				"allow_none": false,
				"title":      "Version",
			},
			"type": {
				"order":      nil,
				"advanced":   false,
				"visible":    false,
				"title":      "Type",
				"type":       "StringType",
				"readonly":   true,
				"allow_none": false,
				"default":    nil,
			},
			"id": {
				"order":      nil,
				"type":       "StringType",
				"advanced":   false,
				"visible":    false,
				"default":    nil,
				"allow_none": false,
				"title":      "Id",
			},
			"name": {
				"order":      nil,
				"type":       "StringType",
				"advanced":   false,
				"visible":    false,
				"default":    nil,
				"allow_none": true,
				"title":      "Name",
			},
			"log_level": {
				"order": nil,
				"options": map[string]int{
					"WARNING":  30,
					"NOTSET":   0,
					"ERROR":    40,
					"INFO":     20,
					"DEBUG":    10,
					"CRITICAL": 50,
				},
				"advanced":   true,
				"visible":    true,
				"title":      "Log Level",
				"type":       "SelectType",
				"enum":       "LogLevel",
				"allow_none": false,
				"default":    "NOTSET",
			},
		},
		Commands: map[nio.Command]nio.CommandDefinition{},
		Name:     "IdentityIntervalSimulator",
	},
}
