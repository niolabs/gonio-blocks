package stdlib_test

import (
	"context"
	"testing"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-blocks/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestCounterBlock_Basic(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := stdlib.CounterBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "Counter",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB"
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	{
		put(t, &b, nio.DefaultTerminal, nil, nil, nil)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		s := signals[0]
		assert.EqualValues(nio.Signal{"count": 3, "cumulative_count": 3}, s)
	}

	{
		put(t, &b, nio.DefaultTerminal, nil)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		s := signals[0]
		assert.EqualValues(nio.Signal{"count": 1, "cumulative_count": 4}, s)
	}

	{
		put(t, &b, nio.DefaultTerminal, nil, nil)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		s := signals[0]
		assert.EqualValues(nio.Signal{"count": 2, "cumulative_count": 6}, s)
	}
}

func TestCounterBlock_Grouped(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := stdlib.CounterBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "Counter",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"group_by": "{{ $group }}"
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	sigA := nio.Signal{"group": "a"}
	sibB := nio.Signal{"group": "b"}

	{
		put(t, &b, nio.DefaultTerminal, sigA, sibB, sigA)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nio.Signal{"group": "a", "count": 2, "cumulative_count": 2},
			nio.Signal{"group": "b", "count": 1, "cumulative_count": 1},
		}, signals)
	}

	{
		put(t, &b, nio.DefaultTerminal, sibB)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nio.Signal{"group": "b", "count": 1, "cumulative_count": 2},
		}, signals)
	}

	{
		put(t, &b, nio.DefaultTerminal, sigA, sigA)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nio.Signal{"group": "a", "count": 2, "cumulative_count": 4},
		}, signals)
	}
}
