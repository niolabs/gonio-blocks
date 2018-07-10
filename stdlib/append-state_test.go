package stdlib_test

import (
	"context"
	"testing"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestAppendStateBlock_Basic(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := stdlib.AppendStateBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "AppendState",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"state_expr": "{{ $state }}"
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	{
		put(t, &b, "getter", nil)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		for _, s := range signals {
			assert.EqualValues(nil, s["state"])
		}
	}

	{
		put(t, &b, "setter", nio.Signal{"state": true})
		put(t, &b, "getter", nil)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		for _, s := range signals {
			assert.EqualValues(true, s["state"])
		}
	}

	{
		put(t, &b, "setter", nio.Signal{"state": -1.0})
		put(t, &b, "getter", nil)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		for _, s := range signals {
			assert.EqualValues(-1.0, s["state"])
		}
	}

}

func TestAppendStateBlock_InitialValue(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := stdlib.AppendStateBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "AppendState",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"state_expr": "{{ $state }}",
	"initial_state": true
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	{
		put(t, &b, "getter", nil)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		for _, s := range signals {
			assert.EqualValues(true, s["state"])
		}
	}

	{
		put(t, &b, "setter", nio.Signal{"state": false})
		put(t, &b, "getter", nil)
		signals := takeOne(t, b.ChOut, &b.Busy)
		assert.Len(signals, 1)
		for _, s := range signals {
			assert.EqualValues(false, s["state"])
		}
	}
}
