package table

// Link against ncurses with wide character support in case goncurses doesn't

// #cgo !darwin,!freebsd,!openbsd pkg-config: ncursesw
// #cgo darwin freebsd openbsd LDFLAGS: -lncurses
// #include <stdlib.h>
// #include <locale.h>
// #include <sys/select.h>
// #include <sys/ioctl.h>
//
// static void grv_FD_ZERO(void *set) {
// 	FD_ZERO((fd_set *)set);
// }
//
// static void grv_FD_SET(int fd, void *set) {
// 	FD_SET(fd, (fd_set *)set);
// }
//
// static int grv_FD_ISSET(int fd, void *set) {
// 	return FD_ISSET(fd, (fd_set *)set);
// }
//
import "C"
import (
	"log"
	"os"
	"syscall"
	"unsafe"
)

// GetScreenSize get screen window size
func GetScreenSize() (int, int) {
	var winSize C.struct_winsize

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), C.TIOCGWINSZ, uintptr(unsafe.Pointer(&winSize)))
	if err != 0 {
		log.Fatal(err)
	}

	rows := int(winSize.ws_row)
	cols := int(winSize.ws_col)
	return cols, rows
}
