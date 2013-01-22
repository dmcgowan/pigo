gorpi
=====

Go library for Raspberry Pi

Reference
---------

[http://elinux.org/RPi_Low-level_peripherals](http://elinux.org/RPi_Low-level_peripherals)

[http://elinux.org/RPi_BCM2835_GPIOs](http://elinux.org/RPi_BCM2835_GPIOs)

Example
-------

Sample program which will display an 8bit value given from command line args onto gpio pins (rev2).

sample.go (*go run sample.go -d 7* will set gpio 7, 8, 9 high and 10, 11, 14, 15, 17 low)

		package main

		import (
			"github.com/dmcgowan/gorpi/gpio"
			"flag"
		)

		var display_value = flag.Int("d", 0, "Value to display on led")

		func main() {
			flag.Parse()
			gpio.Setup()
			defer gpio.Teardown()
	
			output_bits := [8]uint{7, 8, 9, 10, 11, 14, 15, 17}

			for _, bit := range output_bits {
				gpio.FunctionSelect(bit, gpio.OUTPUT)
			}

			value := byte(*display_value)
	
			for i := uint(0); i < 8 ; i++ {
				if (value >> i & 0x01) == 0 {
					gpio.ClearOutput(output_bits[i])
				} else {
					gpio.SetOutput(output_bits[i])
				}
	
			}
		}
