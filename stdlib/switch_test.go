package stdlib_test

import (
	"context"
	"testing"

	"github.com/niolabs/gonio-framework"
	. "github.com/niolabs/gonio-blocks/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestSwitchBlock_Basic(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := SwitchBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "Switch",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"state_expr": "{{ $state }}"
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	b.Enqueue(nio.Terminal("getter"), nio.SignalGroup{nil, nil})
	b.Busy.Wait()
	assert.Len(takeOne(t, b.ChOutRight, &b.Busy), 2)
	assert.Nil(takeNone(t, b.ChOutLeft, &b.Busy))

	b.Enqueue(nio.Terminal("setter"), nio.SignalGroup{
		nio.Signal{"state": true},
	})
	b.Busy.Wait()

	b.Enqueue(nio.Terminal("getter"), nio.SignalGroup{nil, nil})
	b.Busy.Wait()
	assert.Len(takeOne(t, b.ChOutLeft, &b.Busy), 2)
	assert.Nil(takeNone(t, b.ChOutRight, &b.Busy))

	b.Enqueue(nio.Terminal("setter"), nio.SignalGroup{
		nio.Signal{"state": false},
	})
	b.Busy.Wait()

	b.Enqueue(nio.Terminal("getter"), nio.SignalGroup{nil, nil})
	b.Busy.Wait()
	assert.Len(takeOne(t, b.ChOutRight, &b.Busy), 2)
	assert.Nil(takeNone(t, b.ChOutLeft, &b.Busy))
}

func TestSwitchBlock_Group(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := SwitchBlock{}

	if err := b.Configure([]byte(`{
	"type": "Switch",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"state_expr": "{{ $state }}",
	"group_by": "{{ $group }}",
	"group_attr": "g"
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	put(t, &b, "getter", nio.Signal{"group": "foo"}, nio.Signal{"group": "bar"})
	assert.Nil(takeNone(t, b.ChOutLeft, &b.Busy))
	assert.Len(takeOne(t, b.ChOutRight, &b.Busy), 2)

	b.Enqueue(nio.Terminal("setter"), nio.SignalGroup{
		nio.Signal{"state": true, "group": "foo"},
	})
	b.Busy.Wait()

	b.Enqueue(nio.Terminal("getter"), nio.SignalGroup{
		nio.Signal{"group": "foo"},
		nio.Signal{"group": "bar"},
	})
	b.Busy.Wait()

	assert.Len(takeOne(t, b.ChOutLeft, &b.Busy), 1)
	assert.Len(takeOne(t, b.ChOutRight, &b.Busy), 1)

	b.Enqueue(nio.Terminal("setter"), nio.SignalGroup{
		nio.Signal{"state": true, "group": "bar"},
	})
	b.Busy.Wait()

	b.Enqueue(nio.Terminal("getter"), nio.SignalGroup{
		nio.Signal{"group": "foo"},
		nio.Signal{"group": "bar"},
	})
	b.Busy.Wait()

	assert.Len(takeOne(t, b.ChOutLeft, &b.Busy), 2)
	assert.Nil(takeNone(t, b.ChOutRight, &b.Busy), 0)
}
