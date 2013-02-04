PiGo
=====

Go library for Raspberry Pi GPIO

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
			"github.com/dmcgowan/pigo"
			"flag"
			"log"
		)

		var display_value = flag.Int("d", 0, "Value to display on led")

		func main() {
			flag.Parse()
                        gp, err := pigo.NewMmapGPIO() // Actually do something with this error
                        if err != nil {
                                log.Fatalln(err)
                        }
                        defer gp.Close()
                        
			output_bits := [8]pigo.PinNumber{7, 8, 9, 10, 11, 14, 15, 17}
                        
			for _, bit := range output_bits {
				gp.SetDirection(bit, pigo.Out)
			}
                        
			value := pigo.Value(*display_value)
                        
			for i := uint(0); i < 8 ; i++ {
				gp.SetValue(output_bits[i], (value >> i & 0x01))
			}
		}
