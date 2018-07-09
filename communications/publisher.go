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
