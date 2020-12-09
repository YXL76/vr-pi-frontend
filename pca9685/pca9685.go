package pca9685

import (
	"math"
	"time"

	"github.com/go-daq/smbus"
)

const (
	addr        uint8 = 0x40
	modeAdr     uint8 = 0x00
	lowAdr      uint8 = 0x08
	highAdr     uint8 = 0x09
	prescaleAdr uint8 = 0xFE
)

// Device device is a PCA9685 based device.
type Device struct {
	conn *smbus.Conn // connection to smbus
	addr uint8       // address
}

// Open Open
func Open() (*Device, error) {
	c, err := smbus.Open(1, addr)

	dev := Device{
		conn: c,
		addr: addr,
	}
	dev.Write(modeAdr, 0x00)

	time.Sleep(50 * time.Millisecond) // wait required time
	return &dev, err
}

func (dev *Device) Write(reg, value uint8) {
	// writes an 8-bit value to the specified register/address
	_ = dev.conn.WriteReg(dev.addr, reg, value)
}

func (dev *Device) Read(reg uint8) uint8 {
	// read an unsigned byte from the I2C device
	result, _ := dev.conn.ReadReg(dev.addr, reg)
	return result
}

// Close Close
func (dev *Device) Close() error {
	return dev.conn.Close()
}

// SetFrequency SetFrequency
func (dev *Device) SetFrequency(frequency float64) {
	prescaleval := 25000000.0 // 25MHz
	prescaleval /= 4096.0     // 12-bit
	prescaleval /= frequency
	prescaleval -= 1.0

	prescale := math.Floor(prescaleval + 0.5)

	oldmode := dev.Read(modeAdr)
	newmode := (oldmode & 0x7F) | 0x10
	dev.Write(modeAdr, newmode)
	dev.Write(prescaleAdr, uint8(math.Floor(prescale)))
	dev.Write(modeAdr, oldmode)
	time.Sleep(5 * time.Millisecond)
	dev.Write(modeAdr, oldmode|0x80)
}

// SetPulse SetPulse
func (dev *Device) SetPulse(channel uint8, pulse float64) {
	duty := int(pulse * 4096 / 20000) // PWM frequency is 50HZ, the period is 20000us
	dev.Write(lowAdr+4*channel, uint8(duty&0xFF))
	dev.Write(highAdr+4*channel, uint8(duty>>8))
}
