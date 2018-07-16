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

var Modifier = nio.BlockTypeEntry{
	Create: func() nio.Block { return &ModifierBlock{} },
	Definition: nio.BlockTypeDefinition{
		Version: "1.1.0",
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
		Namespace: "blocks.modifier.modifier_block.Modifier",
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
			"exclude": {
				"order":      0,
				"type":       "BoolType",
				"advanced":   false,
				"visible":    true,
				"default":    false,
				"allow_none": false,
				"title":      "Exclude existing fields?",
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
			"fields": {
				"order":         1,
				"advanced":      false,
				"visible":       true,
				"list_obj_type": "ObjectType",
				"title":         "Fields",
				"type":          "ListType",
				"obj_type":      "SignalField",
				"allow_none":    false,
				"template": map[string]interface{}{
					"title": map[string]interface{}{
						"order":      0,
						"type":       "Type",
						"advanced":   false,
						"visible":    true,
						"default":    "",
						"allow_none": false,
						"title":      "Attribute Name",
					},
					"formula": map[string]interface{}{
						"order":      1,
						"type":       "Type",
						"advanced":   false,
						"visible":    true,
						"default":    "",
						"allow_none": true,
						"title":      "Attribute Value",
					},
				},
				"default": []interface{}{},
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
		Name:     "Modifier",
	},
}
