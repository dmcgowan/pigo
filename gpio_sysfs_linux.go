package pigo

import (
	"log"
	"os"
	"strconv"
)

// GPIO object using sysfs
type SysGPIO struct {
	exportedPins map[PinNumber]string
	export *os.File
	unexport *os.File
}

func (pn PinNumber) String() string {
	return strconv.FormatUint(uint64(pn), 10)
}

func handleFileError(err error) {
	if err != nil {
		log.Fatalln("File error:", err)
	}
}

func NewSysGPIO() (*SysGPIO, error) {
	gpio := new(SysGPIO)
	gpio.exportedPins = make(map[PinNumber]string)
	
	var open_err error
	gpio.export, open_err = os.OpenFile("/sys/class/gpio/export", os.O_WRONLY|os.O_SYNC, 0600)
	if open_err != nil {
		log.Println("Error opening /sys/class/gpio/export:", open_err)
		return nil, open_err
	}
	gpio.unexport, open_err = os.OpenFile("/sys/class/gpio/unexport", os.O_WRONLY|os.O_SYNC, 0600)
	if open_err != nil {
		log.Println("Error opening /sys/class/gpio/unexport:", open_err)
		return nil, open_err
	}
	
	return gpio, nil
}

func (gpio *SysGPIO) getPinMaybeExport(pin PinNumber) string {
	prefix, is_exported := gpio.exportedPins[pin]
	if !is_exported {
		_, write_err := gpio.export.WriteString(pin.String()) // Handle error
		handleFileError(write_err)
		gpio.export.Sync()
		gpio.export.Seek(0, 0)
		prefix = "/sys/class/gpio/gpio" + pin.String() + "/"
		gpio.exportedPins[pin] = prefix
	}
	return prefix
}

func (gpio *SysGPIO) SetDirection(pin PinNumber, dir Direction) {
	prefix := gpio.getPinMaybeExport(pin)
	
	dfd, open_err := os.OpenFile(prefix + "direction", os.O_WRONLY|os.O_SYNC, 0600)
	handleFileError(open_err)
	defer dfd.Close()

	if dir == Input {
		_, write_err := dfd.WriteString("in")
		handleFileError(write_err)
	} else {
		_, write_err := dfd.WriteString("out")
		handleFileError(write_err)
	}
}

func (gpio *SysGPIO) GetValue(pin PinNumber) Value {
	prefix := gpio.getPinMaybeExport(pin)
	
	vfd, open_err := os.OpenFile(prefix + "value", os.O_RDONLY, 0600)
	handleFileError(open_err)
	defer vfd.Close()
	var buf [1]byte
	vfd.Read(buf[0:])
	if buf[0] == '0' {
		return Low
	}
	return High
}

func (gpio *SysGPIO) SetValue(pin PinNumber, val Value) {
	prefix := gpio.getPinMaybeExport(pin)

	vfd, open_err := os.OpenFile(prefix + "value", os.O_WRONLY|os.O_SYNC, 0600)
	handleFileError(open_err)
	if val == Low {
		_, write_err := vfd.WriteString("0")
		handleFileError(write_err)
	} else {
		_, write_err := vfd.WriteString("1")
		handleFileError(write_err)
	}
	vfd.Close()
}

func (gpio *SysGPIO) Close() error {
	for pin, _ := range gpio.exportedPins {
		gpio.unexport.WriteString(pin.String())
		gpio.unexport.Sync()
		gpio.unexport.Seek(0, 0)
	}
	gpio.export.Close()
	gpio.unexport.Close()

	return nil
}