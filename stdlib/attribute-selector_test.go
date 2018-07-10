package stdlib_test

import (
	"context"
	"testing"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestAttributeSelectorBlock_WhiteList(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := stdlib.AttributeSelectorBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "AttributeSelector",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"mode": true,
	"attributes": ["foo", "bar"]
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	{
		put(t, &b, nio.DefaultTerminal, nio.Signal{"foo": 1, "bar": 2, "baz": 3})
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		s := signals[0]
		assert.EqualValues(nio.Signal{"foo": 1, "bar": 2}, s)
	}

	{
		put(t, &b, nio.DefaultTerminal, nio.Signal{"baz": 3, "ack": 4})
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		s := signals[0]
		assert.EqualValues(nio.Signal{}, s)
	}
}

func TestAttributeSelectorBlock_BlackList(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := stdlib.AttributeSelectorBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "AttributeSelector",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"mode": false,
	"attributes": ["foo", "bar"]
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	{
		put(t, &b, nio.DefaultTerminal, nio.Signal{"foo": 1, "bar": 2, "baz": 3})
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		s := signals[0]
		assert.EqualValues(nio.Signal{"baz": 3}, s)
	}

	{
		put(t, &b, nio.DefaultTerminal, nio.Signal{"foo": 1, "bar": 2})
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		s := signals[0]
		assert.EqualValues(nio.Signal{}, s)
	}
}
