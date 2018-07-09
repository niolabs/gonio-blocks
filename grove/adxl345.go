package grove

import (
	"bytes"
	"context"
	"encoding/binary"

	"github.com/niolabs/gonio-framework"
	"golang.org/x/exp/io/i2c"
)

const (
	EARTH_GRAVITY_MS2 = 9.80665
	SCALE_MULTIPLIER  = 0.004

	DATA_FORMAT = 0x31
	BW_RATE     = 0x2C
	POWER_CTL   = 0x2D

	BW_RATE_1600HZ = 0x0F
	BW_RATE_800HZ  = 0x0E
	BW_RATE_400HZ  = 0x0D
	BW_RATE_200HZ  = 0x0C
	BW_RATE_100HZ  = 0x0B
	BW_RATE_50HZ   = 0x0A
	BW_RATE_25HZ   = 0x09

	RANGE_2G  = 0x00
	RANGE_4G  = 0x01
	RANGE_8G  = 0x02
	RANGE_16G = 0x03

	MEASURE   = 0x08
	SLEEP     = 0x04
	AXES_DATA = 0x32
)

type adxl345 struct {
	*i2c.Device
}

func (a adxl345) SetBandwidthRate(rate uint8) error {
	return a.Device.WriteReg(BW_RATE, []byte{rate})
}

func (a adxl345) EnableMeasurement() error {
	return a.Device.WriteReg(POWER_CTL, []byte{MEASURE})
}

func (a adxl345) DisableMeasurement() error {
	return a.Device.WriteReg(POWER_CTL, []byte{SLEEP})
}

func (a adxl345) SetRange(rangeFlag uint8) error {
	readBuffer := make([]byte, 1)
	if err := a.Device.ReadReg(DATA_FORMAT, readBuffer); err != nil {
		return err
	}
	value := readBuffer[0]
	value &^= uint8(0xf)
	value |= rangeFlag
	value |= 0x08
	return a.Device.WriteReg(DATA_FORMAT, []byte{value})
}

type RawSample struct {
	X, Y, Z int16
}

func (a adxl345) getRaw() (sample RawSample, err error) {
	buffer := make([]byte, 6)
	if err := a.ReadReg(AXES_DATA, buffer); err != nil {
		return sample, err
	}
	return sample, binary.Read(bytes.NewBuffer(buffer), binary.LittleEndian, &sample)
}

func (a adxl345) getGs() (x, y, z float64, err error) {
	sample, err := a.getRaw()
	if err != nil {
		return 0, 0, 0, err
	}
	x = float64(sample.X) * SCALE_MULTIPLIER
	y = float64(sample.Y) * SCALE_MULTIPLIER
	z = float64(sample.Z) * SCALE_MULTIPLIER
	return
}

func (a adxl345) getMps() (x, y, z float64, err error) {
	x, y, z, err = a.getGs()
	if err != nil {
		return 0, 0, 0, err
	}
	x *= EARTH_GRAVITY_MS2
	y *= EARTH_GRAVITY_MS2
	z *= EARTH_GRAVITY_MS2
	return
}

type ADXL345Block struct {
	nio.Transformer
}

func (b *ADXL345Block) Configure(config nio.RawBlockConfig) error {
	b.Transformer.Configure()
	return nil
}

func (b *ADXL345Block) Start(ctx context.Context) {
	i2cDevice, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-0"}, 0x53)
	if err != nil {
		panic(err)
	}
	defer i2cDevice.Close()

	adxl := adxl345{i2cDevice}

	if err := adxl.SetBandwidthRate(BW_RATE_25HZ); err != nil {
		panic(err)
	}

	if err := adxl.SetRange(RANGE_16G); err != nil {
		panic(err)
	}

	if err := adxl.EnableMeasurement(); err != nil {
		panic(err)
	}

	defer adxl.DisableMeasurement()

	for {
		select {
		case <-b.ChIn:
			x, y, z, err := adxl.getMps()
			if err != nil {
				b.Notify(nio.DefaultTerminal, nio.SignalGroup{
					{"x": x, "y": y, "z": z},
				})
			}
			b.Busy.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (b *ADXL345Block) Enqueue(terminal nio.Terminal, signals nio.SignalGroup) error {
	return b.Transformer.Enqueue(terminal, signals, 1)
}
