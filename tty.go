// tty.go
// tty line discipline functions
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

var (
	savedTermios *unix.Termios
)

func ttySaveTerm() {
	// Get the current terminal settings
	termios, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TCGETS)
	if err == nil {
		savedTermios = termios
	}
}

func ttyRestore() {
	if savedTermios != nil {
		unix.IoctlSetTermios(int(os.Stdout.Fd()), unix.TCSETSF, savedTermios)
	}
}

func ttyBrowser() {
	// Get the current terminal settings
	termios, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TCGETS)
	if err != nil {
		return
	}

	// Save a copy of the original termios for ttyRestore
	if savedTermios == nil {
		savedTermios = termios
	}

	// Set input flags - clear all except INLCR
	termios.Iflag &^= ^uint32(unix.INLCR)

	// Set local flags - clear ISIG, ICANON, ECHO, ECHOK, ECHONL
	lflag := unix.ISIG | unix.ICANON | unix.ECHO | unix.ECHOK | unix.ECHONL
	termios.Lflag &^= uint32(lflag)

	// Set VMIN and VTIME
	termios.Cc[unix.VMIN] = 0
	termios.Cc[unix.VTIME] = 1

	// Set VERASE to backspace
	termios.Cc[unix.VERASE] = '\b'

	// Apply the settings
	unix.IoctlSetTermios(int(os.Stdout.Fd()), unix.TCSETSF, termios)
}

func ttyPrompter() {
	// Get the current terminal settings
	termios, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TCGETS)
	if err != nil {
		return
	}

	// Set input flags - enable INLCR
	termios.Iflag |= unix.INLCR

	// Set local flags - enable ISIG, disable ICANON, ECHO, ECHOK, ECHONL
	lflag := unix.ICANON | unix.ECHO | unix.ECHOK | unix.ECHONL
	termios.Lflag |= unix.ISIG
	termios.Lflag &^= uint32(lflag)

	// Set VMIN and VTIME
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0

	// Apply the settings
	unix.IoctlSetTermios(int(os.Stdout.Fd()), unix.TCSETSF, termios)
}

// vim: set ts=4 sw=4 noet:
