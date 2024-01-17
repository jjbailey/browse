// tty.go
// tty line discipline functions
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"github.com/k0kubun/go-termios"
)

var (
	saneTerm termios.Termios
	prmTerm  termios.Termios
	rawTerm  termios.Termios
)

func ttySaveTerm() {
	saneTerm.GetAttr(termios.Stdout)
}

func ttyRestore() {
	saneTerm.SetAttr(termios.Stdout, termios.TCSAFLUSH)
}

func ttyBrowser() {
	rawTerm = saneTerm

	rawTerm.IFlag &^= termios.INLCR

	rawTerm.LFlag &^= termios.ISIG | termios.ICANON | termios.ECHO | termios.ECHOK | termios.ECHONL

	rawTerm.CC[termios.VMIN], rawTerm.CC[termios.VTIME] = 0, 1

	rawTerm.SetAttr(termios.Stdout, termios.TCSAFLUSH)
}

func ttyPrompter() {
	prmTerm = saneTerm

	prmTerm.IFlag |= termios.INLCR

	prmTerm.LFlag |= termios.ISIG
	prmTerm.LFlag &^= termios.ICANON | termios.ECHO | termios.ECHOK | termios.ECHONL

	prmTerm.CC[termios.VMIN], prmTerm.CC[termios.VTIME] = 1, 0

	prmTerm.SetAttr(termios.Stdout, termios.TCSAFLUSH)
}

// vim: set ts=4 sw=4 noet:
