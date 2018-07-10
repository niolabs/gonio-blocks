package stdlib

import (
	"context"
	"encoding/json"
	"time"

	"github.com/niolabs/gonio-framework"
	"github.com/niolabs/gonio-framework/props"
)

// IdentityIntervalSimulatorBlock
type IdentityIntervalSimulatorBlock struct {
	nio.Producer
	Config IdentityIntervalSimulatorConfig

	duration time.Duration
	limit    int64
	total    int64
	count    int64
}

type IdentityIntervalSimulatorConfig struct {
	nio.BlockConfigAtom
	Interval props.TimeDeltaProperty `json:"interval"`
	Limit    *props.IntProperty      `json:"limit"`
	Count    *props.IntProperty      `json:"num_signals"`
}

func (iis *IdentityIntervalSimulatorBlock) Configure(config nio.RawBlockConfig) error {
	iis.Producer.Configure()

	if err := json.Unmarshal(config, &iis.Config); err != nil {
		return err
	}

	if err := iis.Config.Interval.Assign(&iis.duration, nil); err != nil {
		return err
	}

	if err := iis.Config.Limit.AssignToDefault(&iis.limit, nil, -1); err != nil {
		return err
	}

	if err := iis.Config.Count.AssignToDefault(&iis.count, nil, 1); err != nil {
		return err
	}

	return nil
}

func (iis *IdentityIntervalSimulatorBlock) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return iis.NoEnqueue(terminal)
}

func (iis *IdentityIntervalSimulatorBlock) Start(ctx context.Context) {
	t := time.NewTicker(iis.duration)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			num := iis.count
			hasLimit := iis.limit > 0
			isComplete := hasLimit && (iis.total+num) > iis.limit
			if isComplete {
				num = iis.limit - iis.total
			}

			iis.total += num
			iis.ChOut <- make(nio.SignalGroup, num)

			if isComplete {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
