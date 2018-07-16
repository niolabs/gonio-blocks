package communications

import (
	"context"
	"encoding/json"

	"github.com/niolabs/gonio-framework"
	"github.com/pubkeeper/go-client"
)

type PublisherBlock struct {
	nio.Consumer
	client.Connection

	config PublisherBlockConfig
}

type PublisherBlockConfig struct {
	nio.BlockConfigAtom
	Topic string `json:"topic"`
}

func (block *PublisherBlock) Configure(config nio.RawBlockConfig) error {
	block.Consumer.Configure()

	if err := json.Unmarshal(config, &block.config); err != nil {
		return err
	}

	return nil
}

func (block *PublisherBlock) Start(ctx context.Context) {
	b := block.Connection.RegisterBrewer(block.config.Topic)
	defer block.Connection.UnregisterBrewer(b)

	for {
		select {
		case signals := <-block.ChIn:
			bytes, _ := json.Marshal(signals)
			b.Send <- bytes
		case <-ctx.Done():
			return
		}
	}
}

func (block *PublisherBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return block.Consumer.Enqueue(terminal, signals, 1)
}

func (block *PublisherBlock) EachOutput(func(nio.Terminal, <-chan nio.SignalGroup)) {}

var publisherDefinition = nio.BlockTypeDefinition{
	Version: "1.1.0",
	BlockAttributes: nio.BlockAttributes{
		Outputs: []nio.TerminalDefinition{},
		Inputs: []nio.TerminalDefinition{
			{
				Label:   "default",
				Type:    "input",
				Visible: true,
				Order:   0,
				ID:      "__default_terminal_value",
				Default: true,
			},
		},
	},
	Namespace: "blocks.communication.publisher.Publisher",
	Properties: map[nio.Property]nio.PropertyDefinition{
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
		"timeout": {
			"order":    nil,
			"type":     "TimeDeltaType",
			"advanced": true,
			"visible":  true,
			"default": map[string]float64{
				"seconds": 2,
			},
			"allow_none": false,
			"title":      "Connect Timeout",
		},
		"version": {
			"order":      nil,
			"type":       "StringType",
			"advanced":   true,
			"visible":    true,
			"default":    "1.1.0",
			"allow_none": false,
			"title":      "Version",
		},
		"topic": {
			"order":      nil,
			"type":       "StringType",
			"advanced":   false,
			"visible":    true,
			"default":    nil,
			"allow_none": false,
			"title":      "Topic",
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
	Name:     "Publisher",
}

func NewPublisher(connection client.Connection) nio.BlockTypeEntry {
	return nio.BlockTypeEntry{
		Create: func() nio.Block {
			return &PublisherBlock{
				Connection: connection,
			}
		},
		Definition: publisherDefinition,
	}
}
