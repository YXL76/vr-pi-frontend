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

// Device PCA9685
type Device struct {
	conn *smbus.Conn // smbus 连接
	addr uint8       // 地址
}

// Open Open
func Open() (*Device, error) {
	c, err := smbus.Open(1, addr)

	dev := Device{
		conn: c,
		addr: addr,
	}
	dev.Write(modeAdr, 0x00)

	time.Sleep(50 * time.Millisecond) // 等待一段时间
	return &dev, err
}

func (dev *Device) Write(reg, value uint8) {
	// 将8位值写入指定的寄存器/地址
	_ = dev.conn.WriteReg(dev.addr, reg, value)
}

func (dev *Device) Read(reg uint8) uint8 {
	// 从 I2C 设备读取无符号字节
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
	duty := int(pulse * 4096 / 20000) // PWM 频率为 50HZ, 则周期为 20000us
	dev.Write(lowAdr+4*channel, uint8(duty&0xFF))
	dev.Write(highAdr+4*channel, uint8(duty>>8))
}
