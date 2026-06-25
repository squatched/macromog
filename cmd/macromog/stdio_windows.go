//go:build windows

package main

import (
	"os"
	"syscall"
)

func init() {
	attachInheritedHandle(syscall.STD_OUTPUT_HANDLE, func(f *os.File) { os.Stdout = f })
	attachInheritedHandle(syscall.STD_ERROR_HANDLE, func(f *os.File) { os.Stderr = f })
}

func attachInheritedHandle(which int, set func(*os.File)) {
	h, err := syscall.GetStdHandle(which)
	if err != nil || h == 0 || h == syscall.InvalidHandle {
		return
	}
	set(os.NewFile(uintptr(h), ""))
}
