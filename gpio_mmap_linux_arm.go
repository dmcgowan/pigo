package pigo

import (
	"log"
	"syscall"
	"unsafe"
)

func NewMmapGPIO() (*MmapGPIO, error) {
	mem_fd, mem_err := syscall.Open("/dev/mem", syscall.O_RDWR|syscall.O_SYNC, 0)
	if mem_err != nil {
		log.Println("Error opening /dev/mem:", mem_err)
		return nil, mem_err
	}
	
	addr, _, map_err := syscall.Syscall6(syscall.SYS_MMAP2, uintptr(0), uintptr(GPIO_REGISTER_SIZE), uintptr(syscall.PROT_READ|syscall.PROT_WRITE), uintptr(syscall.MAP_SHARED), uintptr(mem_fd), uintptr(GPIO_REGISTER_BASE/PAGE_SIZE))
	if map_err != 0 {
		log.Println("Error mmap:", map_err)
		return nil, map_err
	}
	
	// Slice memory layout
	var addr_slice = struct {
		addr uintptr
		len  int
		cap  int
	}{addr, GPIO_REGISTER_SIZE / 4, GPIO_REGISTER_SIZE / 4}
	
	gpio := new(MmapGPIO)
	
	// Use unsafe to turn sl into a []uint32.
	gpio.registers = *(*[]uint32)(unsafe.Pointer(&addr_slice))
	gpio.mmapPtr = addr
	
	return gpio, nil
}

func (gpio * MmapGPIO) Close() error {
   _, _, munmap_err := syscall.Syscall(syscall.SYS_MUNMAP, uintptr(gpio.mmapPtr), uintptr(GPIO_REGISTER_SIZE), 0)
	return munmap_err
}
