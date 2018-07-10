package stdlib_test

import (
	"context"
	"testing"

	"github.com/niolabs/gonio-framework"
	. "github.com/niolabs/gonio-framework/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestModifierBlock_Basic(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := &ModifierBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "Modifier",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"fields": [{ "title": "bar", "formula": "{{ $foo + 1 }}" }]
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	put(t, b, nio.DefaultTerminal, nio.Signal{"foo": 1.0})
	signals := takeOne(t, b.ChOut, &b.Busy)
	assert.EqualValues(nio.SignalGroup{
		nio.Signal{"foo": 1.0, "bar": 2.0},
	}, signals)
}

func TestModifierBlock_MultipleSignals(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := &ModifierBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "Modifier",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"fields": [
		{ "title": "sum", "formula": "{{ $a + $b }}" }
	]
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	put(t, b, nio.DefaultTerminal,
		nio.Signal{"a": 5.0, "b": 3.0},
		nio.Signal{"a": 2.0, "b": 2.0},
		nio.Signal{"a": 5.0, "b": 5.0},
	)
	signals := takeOne(t, b.ChOut, &b.Busy)
	assert.EqualValues(nio.SignalGroup{
		nio.Signal{"a": 5.0, "b": 3.0, "sum": 8.0},
		nio.Signal{"a": 2.0, "b": 2.0, "sum": 4.0},
		nio.Signal{"a": 5.0, "b": 5.0, "sum": 10.0},
	}, signals)
}

func TestModifierBlock_MultipleFields(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := &ModifierBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "Modifier",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"fields": [
		{ "title": "sum", "formula": "{{ $a + $b }}" },
		{ "title": "prod", "formula": "{{ $a * $b }}" }
	]
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	put(t, b, nio.DefaultTerminal, nio.Signal{"a": 5.0, "b": 3.0})
	signals := takeOne(t, b.ChOut, &b.Busy)
	assert.EqualValues(nio.SignalGroup{
		nio.Signal{"a": 5.0, "b": 3.0, "sum": 8.0, "prod": 15.0},
	}, signals)
}

func TestModifierBlock_Exclude(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := &ModifierBlock{}

	if err := b.Configure(nio.RawBlockConfig(`{
	"type": "Modifier",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"exclude": true,
	"fields": [
		{ "title": "sum", "formula": "{{ $a + $b }}" }
	]
}`)); err != nil {
		t.Fatal(err)
	}

	go b.Start(ctx)

	put(t, b, nio.DefaultTerminal,
		nio.Signal{"a": 5.0, "b": 3.0},
		nio.Signal{"a": 2.0, "b": 2.0},
		nio.Signal{"a": 5.0, "b": 5.0},
	)
	signals := takeOne(t, b.ChOut, &b.Busy)
	assert.EqualValues(nio.SignalGroup{
		nio.Signal{"sum": 8.0},
		nio.Signal{"sum": 4.0},
		nio.Signal{"sum": 10.0},
	}, signals)
}
