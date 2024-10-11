package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	mod                     = windows.NewLazyDLL("user32.dll")
	procGetClassNameW       = mod.NewProc("GetClassNameW")
	procGetWindowText       = mod.NewProc("GetWindowTextW")
	procGetWindowTextLength = mod.NewProc("GetWindowTextLengthW")
)

type (
	HANDLE uintptr
	HWND   HANDLE
)

func GetClassName(hwnd HWND) (name string, err error) {
	n := make([]uint16, 256)
	p := &n[0]

	r0, _, e1 := syscall.SyscallN(procGetClassNameW.Addr(), 3, uintptr(hwnd), uintptr(unsafe.Pointer(p)), uintptr(len(n)))
	if r0 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}

		return
	}

	name = syscall.UTF16ToString(n)
	return
}

func GetWindowTextLength(hwnd HWND) int {
	ret, _, _ := procGetWindowTextLength.Call(uintptr(hwnd))
	return int(ret)
}

func GetWindowText(hwnd HWND) string {
	textLen := GetWindowTextLength(hwnd) + 1

	buf := make([]uint16, textLen)
	procGetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(textLen))

	return syscall.UTF16ToString(buf)
}

func getWindow(funcName string) uintptr {
	proc := mod.NewProc(funcName)
	hwnd, _, _ := proc.Call()
	return hwnd
}

func main() {

	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)

	go func() {
		for {

			select {
			case <-done:
				return
			case <-ticker.C:
				if hwnd := getWindow("GetForegroundWindow"); hwnd != 0 {
					text := GetWindowText(HWND(hwnd))
					fmt.Println("window :", text, "# hwnd:", hwnd)
				}
			}
		}
	}()

	time.Sleep(10 * time.Second)
	ticker.Stop()
	done <- true

	fmt.Println("finished")
}
