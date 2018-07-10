package stdlib_test

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/niolabs/gonio-framework"
	. "github.com/niolabs/gonio-framework/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestIdentityIntervalSimulatorBlock(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := IdentityIntervalSimulatorBlock{}
	b.Configure(nio.RawBlockConfig(`{
	"type": "IdentityIntervalSimulatorBlock",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"interval": {
		"milliseconds": 50
	}
}`))

	go b.Start(ctx)

	{
		start := time.Now()
		signals := <-b.ChOut

		assert.Len(signals, 1, "should emit one signal")
		assert.Empty(signals[0], "")
		assert.WithinDuration(
			start.Add(time.Millisecond*50),
			time.Now(),
			6*time.Millisecond,
		)
	}

	{
		start := time.Now()
		signals := <-b.ChOut

		assert.Len(signals, 1, "should emit one signal")
		assert.Empty(signals[0], "")
		assert.WithinDuration(
			start.Add(time.Millisecond*50),
			time.Now(),
			6*time.Millisecond,
		)
	}
}

func TestIdentityIntervalSimulatorBlock_Dispatch(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	b := IdentityIntervalSimulatorBlock{}
	b.Configure(nio.RawBlockConfig(`{
	"type": "IdentityIntervalSimulatorBlock",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": "",
	"interval": {
		"milliseconds": 50
	}
}`))

	go b.Start(ctx)

	assert.Error(b.Enqueue(nio.DefaultTerminal, nil))
}

func TestLoggerBlock(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
	}()

	var buffer bytes.Buffer

	b := LoggerBlock{
		Logger: log.New(&buffer, "", 0),
	}

	b.Configure(nio.RawBlockConfig(`{
	"type": "Logger",
	"id": "0787AD0A-456D-46D5-AD47-5BFE2D8CA8BB",
	"name": ""
}`))
	go b.Start(ctx)

	b.Enqueue(nio.DefaultTerminal, nio.SignalGroup{
		map[string]interface{}{"a": 1},
	})

	b.Busy.Wait()

	assert.Equal("map[a:1]\n", buffer.String())
}
