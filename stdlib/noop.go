package stdlib

import (
	"context"
	"encoding/json"

	"github.com/niolabs/gonio-framework"
)

// NoopBlock
type NoopBlock struct {
	nio.Transformer
	Config nio.BlockConfigAtom
}

func (nb *NoopBlock) Configure(config nio.RawBlockConfig) error {
	nb.Transformer.Configure()
	return json.Unmarshal(config, &nb.Config)
}

func (nb *NoopBlock) Start(ctx context.Context) {
	for {
		select {
		case signals := <-nb.ChIn:
			nb.ChOut <- signals
			nb.Busy.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (nb *NoopBlock) Enqueue(t nio.Terminal, sg nio.SignalGroup) error {
	return nb.Consumer.Enqueue(t, sg, 1)
}

func newNoopBlock() nio.Block { return &NoopBlock{} }

var Noop = nio.BlockTypeEntry{
	Create: newNoopBlock,
	Definition: nio.BlockTypeDefinition{
		Namespace: "goblocks.noop",
		Commands:  map[nio.Command]nio.CommandDefinition{},
		Version:   "0.0.0",
		Name:      "Noop",
		Properties: map[nio.Property]nio.PropertyDefinition{
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
			"name": {
				"title":      "Name",
				"order":      nil,
				"type":       "StringType",
				"default":    nil,
				"advanced":   false,
				"allow_none": true,
				"visible":    false,
			},
			"version": {
				"title":      "Version",
				"order":      nil,
				"type":       "StringType",
				"default":    "1.2.0",
				"advanced":   true,
				"allow_none": false,
				"visible":    true,
			},
		},
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
			Outputs: []nio.TerminalDefinition{
				{
					Order:   0,
					Visible: true,
					ID:      "__default_terminal_value",
					Type:    "output",
					Default: true,
					Label:   "default",
				},
			},
		},
	},
}
