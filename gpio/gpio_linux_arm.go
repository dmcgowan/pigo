package gpio

import (
	"log"
	"syscall"
	"unsafe"
)

var mmap_ptr uintptr

func Setup() {
	mem_fd, mem_err := syscall.Open("/dev/mem", syscall.O_RDWR|syscall.O_SYNC, 0)
	if mem_err != nil {
		log.Panicln("Error opening /dev/mem:", mem_err)
	}

	addr, _, map_err := syscall.Syscall6(syscall.SYS_MMAP2, uintptr(0), uintptr(GPIO_REGISTER_SIZE), uintptr(syscall.PROT_READ|syscall.PROT_WRITE), uintptr(syscall.MAP_SHARED), uintptr(mem_fd), uintptr(GPIO_REGISTER_BASE/PAGE_SIZE))
	if map_err != 0 {
		log.Panicln("Error mmap:", map_err)
	}

	// Slice memory layout
	var addr_slice = struct {
		addr uintptr
		len  int
		cap  int
	}{addr, GPIO_REGISTER_SIZE / 4, GPIO_REGISTER_SIZE / 4}

	// Use unsafe to turn sl into a []uint32.
	gpio_map = *(*[]uint32)(unsafe.Pointer(&addr_slice))
	mmap_ptr = addr
}

func Teardown() {
	_, _, munmap_err := syscall.Syscall(syscall.SYS_MUNMAP, uintptr(mmap_ptr), uintptr(GPIO_REGISTER_SIZE), 0)
	if munmap_err != 0 {
		log.Panicln("Error unmapping memory:", munmap_err)
	}
}
