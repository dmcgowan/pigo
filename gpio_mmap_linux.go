package pigo

import (
	"log"
	"time"
)

// GPIO object directly using mmapped io
type MmapGPIO struct {
	mmapPtr uintptr
	registers []uint32
}

const (
	GPIO_REGISTER_BASE = uint32(0x20200000)
	GPIO_REGISTER_SIZE = 0x00A0
	PAGE_SIZE          = 4096

	PUD_OFF  = 0x00 // 00 = Off - disable pull-up/down
	PUD_DOWN = 0x01 // 01 = Enable Pull Down control
	PUD_UP   = 0x02 // 10 = Enable Pull Up contro
)

var (
	// GPIO Registers(32bit each) based at GPIO_REGISTER_BASE
	GPFSEL   [6]int = [6]int{0, 1, 2, 3, 4, 5} // GPIO Function select
	GPSET    [2]int = [2]int{7, 8}             // GPIO Pin Output Set
	GPCLR    [2]int = [2]int{10, 11}           // GPIO Pin Output Clear
	GPLEV    [2]int = [2]int{13, 14}           // GPIO Pin Level
	GPEDS    [2]int = [2]int{16, 17}           // GPIO Pin Event Detect Status
	GPREN    [2]int = [2]int{19, 20}           // GPIO Pin Rising Edge Detect Enable
	GPFEN    [2]int = [2]int{22, 23}           // GPIO Pin Falling Edge Detect Enable
	GPHEN    [2]int = [2]int{25, 26}           // GPIO Pin High Detect Enable
	GPLEN    [2]int = [2]int{28, 29}           // GPIO Pin Low Detect Enable
	GPAREN   [2]int = [2]int{31, 32}           // GPIO Pin Async Rising Edge Detect
	GPAFEN   [2]int = [2]int{34, 35}           // GPIO Pin Async Falling Edge Detect
	GPPUD           = 37                       // GPIO Pin Pull-up/down enable
	GPPUDCLK [2]int = [2]int{38, 39}           // GPIO Pin Pull-up/down Enable Clock
)

func (gpio * MmapGPIO) Dump() {
	log.Println("Mmap length:", len(gpio.registers))
	log.Printf("         10987654321098765432109876543210")
	log.Printf("GPFSEL0: %032b\n", gpio.getRegister32(GPFSEL[0]))
	log.Printf("GPFSEL1: %032b\n", gpio.getRegister32(GPFSEL[1]))
	log.Printf("GPFSEL2: %032b\n", gpio.getRegister32(GPFSEL[2]))
	log.Printf("GPFSEL3: %032b\n", gpio.getRegister32(GPFSEL[3]))
	log.Printf("GPFSEL4: %032b\n", gpio.getRegister32(GPFSEL[4]))
	log.Printf("GPFSEL5: %032b\n", gpio.getRegister32(GPFSEL[5]))
	log.Printf("GPLEV0:  %032b\n", gpio.getRegister32(GPLEV[0]))
	log.Printf("GPLEV1:  %032b\n", gpio.getRegister32(GPLEV[1]))
}

// Sets the direction for the GPIO pin
// GPIO interface implementation
func (gpio * MmapGPIO) SetDirection(pin PinNumber, direction Direction) {
   gpio.SetFunction(pin, PinFunction(direction))
}

// Gets the currently pin level value for the GPIO pin
// GPIO interface implementation
func (gpio * MmapGPIO) GetValue(pin PinNumber) Value {
	register := pin / 32
	shift := uint32(pin % 32)
	mask := uint32(0x01) << shift
	value := gpio.getRegister32(GPLEV[register]) & mask
	return Value(value)
}

// Sets the pin level value for the GPIO pin
// GPIO interface implementation
func (gpio * MmapGPIO) SetValue(pin PinNumber, value Value) {
	if value == High {
		gpio.setRegister32(GPSET[pin/32], 0x01<<uint(pin%32))
	} else if value == Low {
		gpio.setRegister32(GPCLR[pin/32], 0x01<<uint(pin%32))
	}
}

// Sets the function for the GPIO pin
// FunctionalGPIO interface implementation
func (gpio * MmapGPIO) SetFunction(pin PinNumber, function PinFunction) {
	register := pin / 10
	shift := uint32((pin % 10) * 3)
	gpio.maskRegister32(GPFSEL[register], uint32(function)<<shift, 0x07<<shift)
}

// Gets the current function for the GPIO pin
// FunctionalGPIO interface implementation
func (gpio * MmapGPIO) GetFunction(pin PinNumber) PinFunction {
	register := pin / 10
	shift := uint32((pin % 10) * 3)
	value := gpio.getRegister32(GPFSEL[register])
	return PinFunction((value >> shift) & 0x07)
}

// Set 32 bit read/write register value
// using mask to detemine which bits to set
func (gpio * MmapGPIO) maskRegister32(addr int, value, mask uint32) {
	gpio.registers[addr] = (gpio.registers[addr] &^ mask) | value
}

// Set 32 bit write-only register value
func (gpio * MmapGPIO) setRegister32(addr int, value uint32) {
	gpio.registers[addr] = value
}

func (gpio * MmapGPIO) getRegister32(addr int) uint32 {
	return gpio.registers[addr]
}

func (gpio * MmapGPIO) EventDetected(pin uint) bool {
	register := pin / 32
	shift := uint32(pin % 32)
	mask := uint32(0x01) << shift
	value := gpio.getRegister32(GPEDS[register]) & mask
	if value != 0 {
		gpio.ClearEventDetect(pin)
		return true
	}
	return false
}

func (gpio * MmapGPIO) ClearEventDetect(pin uint) {
	register := pin / 32
	shift := uint32(pin % 32)

	gpio.maskRegister32(GPEDS[register], 0x01<<shift, 0x01<<shift)
	time.Sleep(500 * time.Nanosecond)
	gpio.setRegister32(GPEDS[register], 0x00)
}

func (gpio * MmapGPIO) SetPullUpDown(pin uint, pud uint32) {
	register := pin / 32
	shift := uint32(pin % 32)

	// TODO Test pud is PUD_OFF, PUD_UP, or PUD_DOWN
	gpio.maskRegister32(GPPUD, pud, 0x03)

	time.Sleep(250 * time.Nanosecond)
	gpio.setRegister32(GPPUDCLK[register], 0x01<<shift)
	time.Sleep(250 * time.Nanosecond)
	gpio.maskRegister32(GPPUD, PUD_OFF, 0x03)
	gpio.setRegister32(GPPUDCLK[register], 0x00)
}

func (gpio * MmapGPIO) SetRisingEdgeEvent(pin uint, enable uint32) {
	register := pin / 32
	shift := uint32(pin % 32)
	gpio.maskRegister32(GPREN[register], enable<<shift, 0x01<<shift)
}

func (gpio * MmapGPIO) SetFallingEdgeEvent(pin uint, enable uint32) {
	register := pin / 32
	shift := uint32(pin % 32)
	gpio.maskRegister32(GPFEN[register], enable<<shift, 0x01<<shift)
}

func (gpio * MmapGPIO) SetHighEvent(pin uint, enable uint32) {
	register := pin / 32
	shift := uint32(pin % 32)
	gpio.maskRegister32(GPHEN[register], enable<<shift, 0x01<<shift)
}

func (gpio * MmapGPIO) SetLowEvent(pin uint, enable uint32) {
	register := pin / 32
	shift := uint32(pin % 32)
	gpio.maskRegister32(GPHEN[register], enable<<shift, 0x01<<shift)
}
