package stdlib_test

import (
	"context"
	"testing"

	"github.com/niolabs/gonio-framework"
	. "github.com/niolabs/gonio-blocks/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestFilterBlock_1(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := FilterBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "Filter",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"operator": "ALL",
	"conditions": [{ "expr": true }]
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	b.Enqueue(nio.DefaultTerminal, nio.SignalGroup{
		nio.Signal{},
	})

	b.Busy.Wait()

	select {
	case signals := <-b.ChOutLeft:
		assert.Len(signals, 1)
	default:
		t.Error("should have signals on the `true' terminal")
	}

	select {
	case <-b.ChOutRight:
		t.Error("should have no channels on the `false` terminal")
	default:
	}
}

func TestFilterBlock_Multiple(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := FilterBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "Filter",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"operator": "ALL",
	"conditions": [
		{ "expr": true }, 
		{ "expr": false }
	]
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	b.Enqueue(nio.DefaultTerminal, nio.SignalGroup{
		nio.Signal{},
	})

	b.Busy.Wait()
	// TODO refactor this
	select {
	case <-b.ChOutLeft:
		t.Error("should have no signals on the `true' terminal")
	case signals := <-b.ChOutRight:
		assert.Len(signals, 1)
	default:
	}

	select {
	case <-b.ChOutLeft:
		t.Error("should have no signals on the `true' terminal")
	case signals := <-b.ChOutRight:
		assert.Len(signals, 1)
	default:
	}

	select {
	case <-b.ChOutLeft:
		t.Error("should have no signals on the `true' terminal")
	case <-b.ChOutRight:
		t.Error("should have no signals on the `false' terminal")
	default:
	}
}

func TestFilterBlock_Dynamic(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := FilterBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "Filter",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"conditions": [
		{ "expr": "{{ $bool }}" } 
	]
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	b.Enqueue(nio.DefaultTerminal, nio.SignalGroup{
		nio.Signal{"bool": true},
		nio.Signal{"bool": false},
	})

	b.Busy.Wait()

	// TODO refactor this
	select {
	case signals := <-b.ChOutLeft:
		assert.Len(signals, 1)
	case signals := <-b.ChOutRight:
		assert.Len(signals, 1)
	default:
	}

	select {
	case signals := <-b.ChOutLeft:
		assert.Len(signals, 1)
	case signals := <-b.ChOutRight:
		assert.Len(signals, 1)
	default:
	}

	select {
	case <-b.ChOutLeft:
		t.Error("should have no signals on the `true' terminal")
	case <-b.ChOutRight:
		t.Error("should have no signals on the `false' terminal")
	default:
	}
}
