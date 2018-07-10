package stdlib_test

import (
	"context"
	"testing"
	"time"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestDebounceBlock_Basic(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := stdlib.DebounceBlock{}

	if err := b.Configure([]byte(`{
	"type": "Debounce",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"interval": {"milliseconds": 50}
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	{
		put(t, &b, nio.DefaultTerminal, nil, nil, nil)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nil,
		}, signals)
	}

	time.Sleep(25 * time.Millisecond)

	put(t, &b, nio.DefaultTerminal, nil)
	takeNone(t, b.ChOut, &b.Busy)

	time.Sleep(30 * time.Millisecond)

	{
		put(t, &b, nio.DefaultTerminal,
			nio.Signal{"foo": 1},
			nio.Signal{"bar": 2},
			nio.Signal{"baz": 4},
		)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.EqualValues(nio.SignalGroup{
			nio.Signal{"baz": 4},
		}, signals)
	}
}
