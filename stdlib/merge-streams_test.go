package stdlib_test

import (
	"context"
	"testing"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-blocks/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestMergeStreamsBlock_Basic(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := stdlib.MergeStreamsBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "MergeStreams",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB"
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	put(t, &b, "input_1", nio.Signal{"foo": 1})
	takeNone(t, b.ChOut, &b.Busy)

	{
		put(t, &b, "input_2", nio.Signal{"bar": 1})
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nio.Signal{"foo": 1, "bar": 1},
		}, signals)
	}

	put(t, &b, "input_2", nio.Signal{"bar": 2})
	takeNone(t, b.ChOut, &b.Busy)
	put(t, &b, "input_2",
		nio.Signal{"bar": 3},
		nio.Signal{"bar": 4},
		nio.Signal{"bar": 5},
	)
	takeNone(t, b.ChOut, &b.Busy)

	{
		put(t, &b, "input_1",
			nio.Signal{"foo": 2},
			nio.Signal{"foo": 3},
			nio.Signal{"foo": 4},
		)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nio.Signal{"foo": 2, "bar": 5},
		}, signals)
	}

}

func TestMergeStreamsBlock_Multi(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := stdlib.MergeStreamsBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "MergeStreams",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"notify_once": false
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	put(t, &b, "input_1", nio.Signal{"foo": 1})
	takeNone(t, b.ChOut, &b.Busy)

	{
		put(t, &b, "input_2", nio.Signal{"bar": 1})
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nio.Signal{"foo": 1, "bar": 1},
		}, signals)
	}

	{
		put(t, &b, "input_2", nio.Signal{"bar": 2})
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nio.Signal{"foo": 1, "bar": 2},
		}, signals)
	}

	{
		put(t, &b, "input_2",
			nio.Signal{"bar": 3},
			nio.Signal{"bar": 4},
			nio.Signal{"bar": 5},
		)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nio.Signal{"foo": 1, "bar": 3},
			nio.Signal{"foo": 1, "bar": 4},
			nio.Signal{"foo": 1, "bar": 5},
		}, signals)
	}

	{
		put(t, &b, "input_1",
			nio.Signal{"foo": 2},
			nio.Signal{"foo": 3},
			nio.Signal{"foo": 4},
		)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nio.Signal{"foo": 2, "bar": 5},
			nio.Signal{"foo": 3, "bar": 5},
			nio.Signal{"foo": 4, "bar": 5},
		}, signals)
	}

}
