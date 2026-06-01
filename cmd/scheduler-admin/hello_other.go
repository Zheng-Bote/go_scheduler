//go:build !windows

package main

import (
	"fyne.io/fyne/v2"
)

func verifyWindowsHello(w fyne.Window) (bool, error) {
	// No-op for non-Windows platforms, just return true
	return true, nil
}
