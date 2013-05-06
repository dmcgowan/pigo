package pigo

import (
	"fmt"
	"log"
	"os"
	"sync"
	"syscall"
	"time"
)

// TODO move to non-system specific
type PinEventType byte

const (
	E_RISING_EDGE   = PinEventType(0x01)
	E_FALLING_EDGE  = PinEventType(0x02)
	E_HIGH          = PinEventType(0x04)
	E_LOW           = PinEventType(0x08)
	E_ASYNC_RISING  = PinEventType(0x10)
	E_ASYNC_FALLING = PinEventType(0x20)
)

type PinEvent struct {
	Pin   uint
	Value uint
	Types PinEventType
}

type PinWatcher struct {
	watches  map[uint]PinEventType
	channel  chan PinEvent
	pin      int
	shutdown bool
	wg       sync.WaitGroup
}

func NewWatcher(pin int) (watcher *PinWatcher) {
	watcher = &PinWatcher{shutdown: false, pin: pin}
	watcher.watches = make(map[uint]PinEventType)
	watcher.channel = make(chan PinEvent, 5)

	watcher.wg.Add(1)
	go watcher.pinHandler()

	return
}

func (watcher *PinWatcher) Close() {
	watcher.shutdown = true
	watcher.wg.Wait()
}

func (watcher *PinWatcher) pinHandler() {
	exf, exf_err := os.OpenFile("/sys/class/gpio/export", syscall.O_WRONLY, 0400)
	if exf_err != nil {
		log.Panicln("Error opening /sys/class/gpio/export:", exf_err)
	}
	_, ex_err := exf.WriteString(fmt.Sprintf("%d", watcher.pin))
	if ex_err != nil {
		log.Panicln("Error writing to /sys/class/gpio/export:", ex_err)
	}
	exf.Close()
	time.Sleep(time.Microsecond)

	edge_file := fmt.Sprintf("/sys/class/gpio/gpio%d/edge", watcher.pin)
	edgef, edgef_err := os.OpenFile(edge_file, syscall.O_WRONLY, 0400)
	if edgef_err != nil {
		log.Panicf("Error opening %s: %s\n", edge_file, edgef_err)
	}
	_, edge_err := edgef.WriteString("both")
	if edge_err != nil {
		log.Panicf("Error writing to %s: %s\n", edge_file, edge_err)
	}
	edgef.Close()
	time.Sleep(time.Microsecond)

	value_file := fmt.Sprintf("/sys/class/gpio/gpio%d/value", watcher.pin)
	irq_fd, irq_err := syscall.Open(value_file, syscall.O_RDONLY|syscall.O_NONBLOCK, syscall.S_IREAD)
	if irq_err != nil {
		log.Panicln("Error opening %s: %s\n", value_file, irq_err)
	}

	epfd, eperr := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if eperr != nil {
		log.Panicln("Error creating epoll:", eperr)
	}

	event := new(syscall.EpollEvent)
	event.Fd = int32(irq_fd)
	event.Events |= syscall.EPOLLPRI
	if ctlerr := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, irq_fd, event); ctlerr != nil {
		log.Panicln("Error on epoll control operation:", ctlerr)
	}

	var events_buffer [10]syscall.EpollEvent
	var buf [1]byte
	for !watcher.shutdown {
		n, werr := syscall.EpollWait(epfd, events_buffer[0:], 4000)
		if werr != nil {
			log.Println("Epoll error:", werr)
		} else if n == 1 {
			syscall.Seek(irq_fd, 0, 0)
			syscall.Read(irq_fd, buf[0:])
			log.Println("Interrupt!", events_buffer[0].Fd, events_buffer[0].Events, buf)
		} else {
			log.Println("Timeout")
		}
	}

	syscall.Close(irq_fd)

	unexf, unexf_err := os.OpenFile("/sys/class/gpio/unexport", syscall.O_WRONLY|syscall.O_SYNC, 0400)
	if unexf_err != nil {
		log.Panicln("Error opening /sys/class/gpio/unexport:", unexf_err)
	}
	_, unex_err := unexf.WriteString(fmt.Sprintf("%d", watcher.pin))
	if unex_err != nil {
		log.Panicln("Error writing to /sys/class/gpio/unexport:", unex_err)
	}

	unexf.Close()

	watcher.wg.Done()
}
