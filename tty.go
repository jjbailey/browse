// tty.go
// tty line discipline functions
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

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

	rawTerm.IFlag &= termios.INLCR
	rawTerm.LFlag ^= termios.ISIG | termios.ICANON | termios.ECHO | termios.ECHOK | termios.ECHONL
	rawTerm.CC[termios.VTIME] = 1
	rawTerm.CC[termios.VMIN] = 0
	// depends on key mapping
	rawTerm.CC[termios.VERASE] = '\b'

	rawTerm.SetAttr(termios.Stdout, termios.TCSAFLUSH)
}

func ttyPrompter() {
	prmTerm = saneTerm

	prmTerm.IFlag |= termios.INLCR
	prmTerm.LFlag |= termios.ISIG
	prmTerm.LFlag ^= termios.ICANON | termios.ECHO | termios.ECHOK | termios.ECHONL
	prmTerm.CC[termios.VTIME] = 0
	prmTerm.CC[termios.VMIN] = 1

	prmTerm.SetAttr(termios.Stdout, termios.TCSAFLUSH)
}

// vim: set ts=4 sw=4 noet:
