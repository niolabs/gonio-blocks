package grove

import (
	"bytes"
	"context"
	"encoding/binary"

	"github.com/niolabs/gonio-framework"
	"golang.org/x/exp/io/i2c"
	"encoding/json"
	"fmt"
)

const (
	gMps2           = 9.80665
	scaleMultiplier = 0.004

	dataFormat = 0x31
	bwRate     = 0x2C
	powerCtl   = 0x2D

	bwRate1600HZ = 0x0F
	bwRate800HZ  = 0x0E
	bwRate400HZ  = 0x0D
	bwRate200HZ  = 0x0C
	bwRate100HZ  = 0x0B
	bwRate50HZ   = 0x0A
	bwRate25HZ   = 0x09

	range16G = 0x03
	range8G  = 0x02
	range4G  = 0x01
	range2G  = 0x00

	measure  = 0x08
	sleep    = 0x04
	axesData = 0x32
)

var (
	enablePayload  = []byte{powerCtl, measure}
	disablePayload = []byte{powerCtl, sleep}
)

type adxl345 struct {
	*i2c.Device
}

func (a adxl345) SetBandwidthRate(rate uint8) error {
	return a.Device.WriteReg(bwRate, []byte{rate})
}

func (a adxl345) EnableMeasurement() error {
	return a.Device.Write(enablePayload)
}

func (a adxl345) DisableMeasurement() error {
	return a.Device.Write(disablePayload)
}

func (a adxl345) SetRange(rangeFlag uint8) error {
	buffer := make([]byte, 1)
	if err := a.Device.ReadReg(dataFormat, buffer); err != nil {
		return err
	}
	value := buffer[0]
	value &^= uint8(0xf)
	value |= rangeFlag
	value |= 0x08
	return a.Device.Write([]byte{dataFormat, value})
}

type rawSample struct {
	X, Y, Z int16
}

func (a adxl345) getRaw() (sample rawSample, err error) {
	buffer := make([]byte, 6)
	if err := a.ReadReg(axesData, buffer); err != nil {
		return sample, err
	}
	return sample, binary.Read(bytes.NewBuffer(buffer), binary.LittleEndian, &sample)
}

func (a adxl345) getGs() (x, y, z float64, err error) {
	sample, err := a.getRaw()
	if err != nil {
		return 0, 0, 0, err
	}
	x = float64(sample.X) * scaleMultiplier
	y = float64(sample.Y) * scaleMultiplier
	z = float64(sample.Z) * scaleMultiplier
	return
}

func (a adxl345) getMps() (x, y, z float64, err error) {
	x, y, z, err = a.getGs()
	if err != nil {
		return 0, 0, 0, err
	}
	x *= gMps2
	y *= gMps2
	z *= gMps2
	return
}

type ADXL345Block struct {
	nio.Transformer
	Config struct {
		nio.BlockConfigAtom
	}

	bus uint
}

func (b *ADXL345Block) Configure(config nio.RawBlockConfig) error {
	b.Transformer.Configure()

	if err := json.Unmarshal(config, &b.Config); err != nil {
		return err
	}

	return nil
}

func (b *ADXL345Block) Start(ctx context.Context) {
	device := fmt.Sprintf("/dev/i2c-%d", b.bus)
	i2cDevice, err := i2c.Open(&i2c.Devfs{Dev: device}, 0x53)
	if err != nil {
		panic(err)
	}
	defer i2cDevice.Close()

	adxl := adxl345{i2cDevice}
	if err := adxl.SetBandwidthRate(bwRate100HZ); err != nil {
		panic(err)
	}

	if err := adxl.SetRange(range16G); err != nil {
		panic(err)
	}

	if err := adxl.EnableMeasurement(); err != nil {
		panic(err)
	}

	defer adxl.DisableMeasurement()

	for {
		select {
		case <-b.ChIn:
			x, y, z, err := adxl.getGs()
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

var DefaultADXL345 = nio.BlockTypeEntry{
	Create:     func() nio.Block { return &ADXL345Block{bus: 0} },
	Definition: nio.BlockTypeDefinition{},
}

func NewADXL345(bus uint) nio.BlockTypeEntry {
	return nio.BlockTypeEntry{
		Create:     func() nio.Block { return &ADXL345Block{bus: bus} },
		Definition: nio.BlockTypeDefinition{},
	}
}
