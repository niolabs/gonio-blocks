package communications

import (
	"context"
	"encoding/json"

	"github.com/niolabs/gonio-framework"
	"github.com/pubkeeper/go-client"
)

type SubscriberBlock struct {
	nio.Producer
	client.Connection

	config SubscriberBlockConfig
}

type SubscriberBlockConfig struct {
	nio.BlockConfigAtom
	Topic string `json:"topic"`
}

func (block *SubscriberBlock) Configure(config nio.RawBlockConfig) error {
	block.Producer.Configure()

	if err := json.Unmarshal(config, &block.config); err != nil {
		return err
	}

	return nil
}

func (block *SubscriberBlock) Start(ctx context.Context) {
	p := block.Connection.RegisterPatron(block.config.Topic)
	defer block.Connection.UnregisterPatron(p)

	for {
		select {
		case bytes := <-p.Recv:
			var signals nio.SignalGroup
			json.Unmarshal(bytes, &signals)
			block.ChOut <- signals
		case <-ctx.Done():
			return
		}
	}
}

func (block *SubscriberBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return block.NoEnqueue(terminal)
}

func (block *SubscriberBlock) EachOutput(fn func(nio.Terminal, <-chan nio.SignalGroup)) {
	fn(block.TOut, block.ChOut)
}
