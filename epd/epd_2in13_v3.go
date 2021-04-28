package epd

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"image"
	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
	"time"
)

const width = 104
const height = 212

type Dev2in13v3 struct {
	resetPin gpio.PinOut
	dcPin gpio.PinOut
	busyPin gpio.PinIn
	csPin gpio.PinOut
	width int
	height int
	conn spi.Conn
	port spi.PortCloser
	log *zap.SugaredLogger
}

func NewEpd2in13v3(logger *zap.SugaredLogger) *Dev2in13v3 {
	return &Dev2in13v3{
		width: width,
		height: height,
		log: logger,
	}
}

func (e *Dev2in13v3) Init() error {
	e.log.Info("initializing host")
	_, err := host.Init()
	if err != nil {
		return errors.Wrap(err, "could not initialize host")
	}

	if _, err := driverreg.Init(); err != nil {
		return errors.Wrap(err, "could not initialize driverreg")
	}

	if !rpi.Present() {
		return errors.New("raspberry board not detected")
	}

	e.resetPin = gpioreg.ByName("17")
	e.dcPin = gpioreg.ByName("25")
	e.busyPin = gpioreg.ByName("24")
	e.csPin = gpioreg.ByName("8")

	e.log.Info("opening SPI device")
	e.port, err = spireg.Open("")
	if err != nil {
		return errors.Wrap(err, "could not open SPI")
	}

	// Convert the spi.Port into a spi.Conn so it can be used for communication.
	e.conn, err = e.port.Connect(4*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		return errors.Wrap(err, "could not connect to SPI")
	}

	if err := e.Reset(); err != nil {
		return errors.Wrap(err, "could not reset the device")
	}
	time.Sleep(10*time.Millisecond)

	if err = e.sendCommand(0x04); err != nil {
		return errors.Wrap(err, "could not send command 0x04")
	}

	if err = e.waitUntilIdle(); err != nil {
		return err
	}

	if err = e.sendCommand(0x00); err != nil {
		return errors.Wrap(err, "could not set panel setting")
	}

	if err = e.sendData([]byte{0x0f}); err != nil {
		return errors.Wrap(err, "could not set LUT")
	}

	if err = e.sendData([]byte{0x89}); err != nil {
		return errors.Wrap(err, "could not set timing settings")
	}

	if err = e.sendCommand(0x61); err != nil {
		return errors.Wrap(err, "could not set resolution")
	}

	if err = e.sendData([]byte{0x68}); err != nil {
		return errors.Wrap(err, "could not send 0x68")
	}

	if err = e.sendData([]byte{0x00}); err != nil {
		return errors.Wrap(err, "could not send 0x00")
	}

	if err = e.sendData([]byte{0xD4}); err != nil {
		return errors.Wrap(err, "could not send 0xD4")
	}

	if err = e.sendCommand(0x50); err != nil {
		return errors.Wrap(err, "could not VCOM and data interval settings")
	}

	if err = e.sendData([]byte{0x77}); err != nil {
		return errors.Wrap(err, "could not set WBmode/WBRmode")
	}

	return nil
}

func (e *Dev2in13v3) Height() int {
	return e.height
}

func (e *Dev2in13v3) Width() int {
	return e.width
}

func (e *Dev2in13v3) Close() error {
	if e.port == nil {
		return nil
	}
	return e.port.Close()
}

func (e *Dev2in13v3) Clear() error {
	buff := make([]byte, e.width * e.height / 8)
	for i:=0; i < e.width * e.height / 8; i++ {
		buff[i] = 0xFF
	}

	return e.Display(buff, buff)
}

func (e *Dev2in13v3) Display(blacks, reds []byte) error {
	err := e.sendCommand(0x10)
	if err != nil {
		return errors.Wrap(err, "could not send command 0x10 to device")
	}

	if err = e.sendData(blacks); err != nil {
		return errors.Wrap(err, "could not send black pixels data to device")
	}

	err = e.sendCommand(0x13)
	if err != nil {
		return errors.Wrap(err, "could not send command 0x13 to device")
	}

	if err = e.sendData(reds); err != nil {
		return errors.Wrap(err, "could not send red pixel data to device")
	}

	err = e.sendCommand(0x12)
	if err != nil {
		return errors.Wrap(err, "could not send command 0x12 to device")
	}

	time.Sleep(100*time.Millisecond)
	err = e.waitUntilIdle()
	if err != nil {
		return errors.Wrap(err, "could not wait for the device")
	}

	return nil
}

func (e *Dev2in13v3) Reset() error {
	if err := e.resetPin.Out(gpio.High); err != nil {
		return errors.Wrap(err, "could not set RESET pin to HIGH")
	}
	time.Sleep(200*time.Millisecond)
	if err := e.resetPin.Out(gpio.Low); err != nil {
		return errors.Wrap(err, "could not set RESET pin to low")
	}
	time.Sleep(time.Millisecond)
	if err := e.resetPin.Out(gpio.High); err != nil {
		return errors.Wrap(err, "could not set RESET pin to HIGH")
	}
	time.Sleep(200*time.Millisecond)

	return nil
}

func (e *Dev2in13v3) sendCommand(b byte) error {
	// skip noisy commands (wait for idle)
	e.log.Debug("sending command: 0x%x\n", b)
	if err := e.dcPin.Out(gpio.Low); err != nil {
		return errors.Wrap(err, "could not set DC pin to LOW")
	}
	if err := e.csPin.Out(gpio.Low); err != nil {
		return errors.Wrap(err, "could not set CS pin to LOW")
	}
	err := e.conn.Tx([]byte{b}, nil)
	if err != nil {
		return errors.Wrap(err, "failed to write command to device")
	}
	if err := e.csPin.Out(gpio.High); err != nil {
		return errors.Wrap(err, "could not set CS pin to HIGH")
	}

	return nil
}

func (e *Dev2in13v3) sendData(b []byte) error {
	if err := e.dcPin.Out(gpio.High); err != nil {
		return errors.Wrap(err, "could not set DC pin to HIGH")
	}
	if err := e.csPin.Out(gpio.Low); err != nil {
		return errors.Wrap(err, "could not set CS pin to LOW")
	}
	lower := 0
	limit := e.conn.(conn.Limits).MaxTxSize()
	upper := 0
	for {
		if lower >= len(b) {
			break
		}
		if lower+limit >= len(b) {
			upper = len(b)
		} else {
			upper = lower+limit
		}
		err := e.conn.Tx(b[lower:upper], nil)
		if err != nil {
			return errors.Wrap(err, "failed to write data to device")
		}
		lower = upper
	}

	if err := e.csPin.Out(gpio.High); err != nil {
		return errors.Wrap(err, "could not set CS pin to HIGH")
	}

	return nil
}

func (e *Dev2in13v3) waitUntilIdle() error {
	e.log.Debug("busy")
	err := e.sendCommand(0x71)
	if err != nil {
		return errors.Wrap(err, "could not send command 0x71")
	}

	for {
		if e.busyPin.Read() == gpio.High {
			break
		}
		time.Sleep(100 * time.Millisecond)

		err = e.sendCommand(0x71)
		if err != nil {
			return errors.Wrap(err, "could not send command 0x71")
		}
	}
	e.log.Debug("busy release")

	return nil
}

func (e Dev2in13v3) Bounds() image.Rectangle {
	return image.Rect(0, 0, e.Width(), e.Height())
}

func (e Dev2in13v3) BoundsHorizontal() image.Rectangle {
	return image.Rect(0, 0, e.Height(), e.Width())
}