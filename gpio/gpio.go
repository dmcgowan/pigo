package gpio

import (
	"log"
	"time"
)

const (
	GPIO_REGISTER_BASE = uint32(0x20200000)
	GPIO_REGISTER_SIZE = 0x00A0
	PAGE_SIZE          = 4096

	INPUT  = 0x00 // 000 GPIO Pin is an input
	OUTPUT = 0x01 // 001 GPIO Pin is an output
	ALT0   = 0x04 // 100 GPIO Pin takes alternate function 0
	ALT1   = 0x05 // 101 GPIO Pin takes alternate function 1
	ALT2   = 0x06 // 110 GPIO Pin takes alternate function 2
	ALT3   = 0x07 // 111 GPIO Pin takes alternate function 3
	ALT4   = 0x03 // 011 GPIO Pin takes alternate function 4
	ALT5   = 0x02 // 010 GPIO Pin takes alternate function 5

	PUD_OFF  = 0x00 // 00 = Off – disable pull-up/down
	PUD_DOWN = 0x01 // 01 = Enable Pull Down control
	PUD_UP   = 0x02 // 10 = Enable Pull Up contro

	DISABLED = uint32(0x00)
	ENABLED  = uint32(0x01)
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

var gpio_map []uint32

func Dump() {
	log.Println("Mmap length:", len(gpio_map))
	log.Printf("         10987654321098765432109876543210")
	log.Printf("GPFSEL0: %032b\n", getRegister32(GPFSEL[0]))
	log.Printf("GPFSEL1: %032b\n", getRegister32(GPFSEL[1]))
	log.Printf("GPFSEL2: %032b\n", getRegister32(GPFSEL[2]))
	log.Printf("GPFSEL3: %032b\n", getRegister32(GPFSEL[3]))
	log.Printf("GPFSEL4: %032b\n", getRegister32(GPFSEL[4]))
	log.Printf("GPFSEL5: %032b\n", getRegister32(GPFSEL[5]))
	log.Printf("GPLEV0:  %032b\n", getRegister32(GPLEV[0]))
	log.Printf("GPLEV1:  %032b\n", getRegister32(GPLEV[1]))
}

// Set 32 bit read/write register value
// using mask to detemine which bits to set
func maskRegister32(addr int, value, mask uint32) {
	gpio_map[addr] = (gpio_map[addr] &^ mask) | value
}

// Set 32 bit write-only register value
func setRegister32(addr int, value uint32) {
	gpio_map[addr] = value
}

func getRegister32(addr int) uint32 {
	return gpio_map[addr]
}

func FunctionSelect(pin uint, function uint32) {
	register := pin / 10
	shift := uint32((pin % 10) * 3)
	maskRegister32(GPFSEL[register], function<<shift, 0x07<<shift)
}

func GetFunction(pin uint) uint32 {
	register := pin / 10
	shift := uint32((pin % 10) * 3)
	value := getRegister32(GPFSEL[register])
	return (value >> shift) & 0x07
}

func SetOutput(pin uint) {
	setRegister32(GPSET[pin/32], 0x01<<uint(pin%32))
}

func ClearOutput(pin uint) {
	setRegister32(GPCLR[pin/32], 0x01<<uint(pin%32))
}

// Reads the pin level, returns true if high, false if low
func PinLevel(pin uint) bool {
	register := pin / 32
	shift := uint32(pin % 32)
	mask := uint32(0x01) << shift
	value := getRegister32(GPLEV[register]) & mask
	return value != 0
}

func EventDetected(pin uint) bool {
	register := pin / 32
	shift := uint32(pin % 32)
	mask := uint32(0x01) << shift
	value := getRegister32(GPEDS[register]) & mask
	if value != 0 {
		ClearEventDetect(pin)
		return true
	}
	return false
}

func ClearEventDetect(pin uint) {
	register := pin / 32
	shift := uint32(pin % 32)

	maskRegister32(GPEDS[register], 0x01<<shift, 0x01<<shift)
	time.Sleep(500 * time.Nanosecond)
	setRegister32(GPEDS[register], 0x00)
}

func SetPullUpDown(pin uint, pud uint32) {
	register := pin / 32
	shift := uint32(pin % 32)

	// TODO Test pud is PUD_OFF, PUD_UP, or PUD_DOWN
	maskRegister32(GPPUD, pud, 0x03)

	time.Sleep(250 * time.Nanosecond)
	setRegister32(GPPUDCLK[register], 0x01<<shift)
	time.Sleep(250 * time.Nanosecond)
	maskRegister32(GPPUD, PUD_OFF, 0x03)
	setRegister32(GPPUDCLK[register], 0x00)
}

func SetRisingEdgeEvent(pin uint, enable uint32) {
	register := pin / 32
	shift := uint32(pin % 32)
	maskRegister32(GPREN[register], enable<<shift, 0x01<<shift)
}

func SetFallingEdgeEvent(pin uint, enable uint32) {
	register := pin / 32
	shift := uint32(pin % 32)
	maskRegister32(GPFEN[register], enable<<shift, 0x01<<shift)
}

func SetHighEvent(pin uint, enable uint32) {
	register := pin / 32
	shift := uint32(pin % 32)
	maskRegister32(GPHEN[register], enable<<shift, 0x01<<shift)
}

func SetLowEvent(pin uint, enable uint32) {
	register := pin / 32
	shift := uint32(pin % 32)
	maskRegister32(GPHEN[register], enable<<shift, 0x01<<shift)
}
