package stdlib

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/niolabs/gonio-framework"
)

// LoggerBlock
type LoggerBlock struct {
	nio.Consumer
	Config nio.BlockConfigAtom
	*log.Logger
}

func (lb *LoggerBlock) Configure(config nio.RawBlockConfig) error {
	if lb.Logger == nil {
		lb.Logger = log.New(os.Stderr, log.Prefix(), log.Flags())
	}

	lb.Consumer.Configure()

	if err := json.Unmarshal(config, &lb.Config); err != nil {
		return err
	}

	return nil
}

func (lb *LoggerBlock) process(signals nio.SignalGroup) {
	for _, sig := range signals {
		lb.Logger.Printf("%+v\n", sig)
	}
	lb.Busy.Done()
}

func (lb *LoggerBlock) Start(ctx context.Context) {
	for {
		select {
		case signals := <-lb.ChIn:
			lb.process(signals)
		case <-ctx.Done():
			return
		}
	}
}

func (lb *LoggerBlock) Enqueue(t nio.Terminal, signals nio.SignalGroup) error {
	return lb.Consumer.Enqueue(t, signals, 1)
}

func newLoggerBlock() nio.Block { return &LoggerBlock{} }

var Logger = nio.BlockTypeEntry{
	Create: newLoggerBlock,
	Definition: nio.BlockTypeDefinition{
		Name:      "Logger",
		Version:   "1.1.0",
		Namespace: "goblocks.logger.logger_block.Logger",
		BlockAttributes: nio.BlockAttributes{
			Inputs: []nio.TerminalDefinition{
				{
					Order:   0,
					Visible: true,
					ID:      "__default_terminal_value",
					Type:    "input",
					Default: true,
					Label:   "default",
				},
			},
			Outputs: []nio.TerminalDefinition{},
		},
		Properties: map[nio.Property]nio.PropertyDefinition{
			"version": {
				"title":      "Version",
				"order":      nil,
				"type":       "StringType",
				"default":    "1.1.0",
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
			"id": {
				"title":      "Id",
				"order":      nil,
				"type":       "StringType",
				"default":    nil,
				"advanced":   false,
				"allow_none": false,
				"visible":    false,
			},
			"log_as_list": {
				"title":      "Log as a list",
				"order":      nil,
				"type":       "BoolType",
				"default":    false,
				"advanced":   false,
				"allow_none": false,
				"visible":    false,
			},
			"log_at": {
				"title":      "Log At",
				"advanced":   false,
				"allow_none": false,
				"visible":    true,
				"enum":       "LogLevel",
				"order":      nil,
				"type":       "SelectType",
				"default":    "INFO",
				"options": map[string]int{
					"CRITICAL": 50,
					"DEBUG":    10,
					"ERROR":    40,
					"WARNING":  30,
					"INFO":     20,
					"NOTSET":   0,
				},
			},
			"log_level": {
				"title":      "Log Level",
				"advanced":   false,
				"allow_none": false,
				"visible":    true,
				"enum":       "LogLevel",
				"order":      nil,
				"type":       "SelectType",
				"default":    "INFO",
				"options": map[string]int{
					"CRITICAL": 50,
					"DEBUG":    10,
					"ERROR":    40,
					"WARNING":  30,
					"INFO":     20,
					"NOTSET":   0,
				},
			},
			"log_hidden_attributes": {
				"title":      "Log Hidden Attributes",
				"order":      nil,
				"type":       "BoolType",
				"default":    false,
				"advanced":   false,
				"allow_none": false,
				"visible":    true,
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
		},
		Commands: map[nio.Command]nio.CommandDefinition{},
	},
}

func NewLogger(dst *log.Logger) nio.BlockTypeEntry {
	return nio.BlockTypeEntry{
		Create:     func() nio.Block { return &LoggerBlock{Logger: dst} },
		Definition: Logger.Definition,
	}
}
