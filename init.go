// init.go
// functions to initialize objects
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"os"

	"golang.org/x/term"
)

func (br *browseObj) fileInit(fp *os.File, fileName, title string, fromStdin bool) {
	// file initializations

	br.fp = fp
	br.fileName = fileName
	br.title = title
	br.fromStdin = fromStdin
}

func (br *browseObj) browseInit() {
	// screen initializations

	br.ignoreCase = false
	br.lastMatch = SEARCH_RESET
	br.hitEOF = false
	br.shownEOF = false
	br.shownMsg = false
	br.shiftWidth = 0
	br.modeNumbers = false
	br.modeScroll = MODE_SCROLL_NONE

	br.saveRC = false
	br.exit = false
}

func (br *browseObj) screenInit(tty *os.File) {
	// tty initializations

	br.tty = tty

	width, height, err := term.GetSize(int(tty.Fd()))
	if err != nil {
		br.dispWidth = 80
		br.dispHeight = 25
	} else {
		br.dispWidth = width
		br.dispHeight = height
	}

	br.dispRows = br.dispHeight - 1
}

// vim: set ts=4 sw=4 noet:
