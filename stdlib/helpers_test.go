package stdlib_test

import (
	"sync"
	"testing"

	"github.com/niolabs/gonio-framework"
)

func put(t *testing.T, b nio.Block, terminal string, signals ...nio.Signal) {
	if err := b.Enqueue(nio.Terminal(terminal), signals); err != nil {
		t.Error(err)
	}
}

func takeOne(t *testing.T, b <-chan nio.SignalGroup, wg *sync.WaitGroup) nio.SignalGroup {
	wg.Wait()
	select {
	case signals := <-b:
		return signals
	default:
		t.Error("channel has no signals")
	}

	return nil
}

func takeNone(t *testing.T, b <-chan nio.SignalGroup, wg *sync.WaitGroup) nio.SignalGroup {
	wg.Wait()
	select {
	case signals := <-b:
		t.Errorf("channel has %d signals", len(signals))
		return signals
	default:
		return nil
	}
}
