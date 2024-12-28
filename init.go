// init.go
// functions to initialize objects
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"os"

	"golang.org/x/term"
)

func (x *browseObj) fileInit(fp *os.File, fileName, title string, fromStdin bool) {
	x.fp = fp
	x.fileName = fileName
	x.title = title
	x.fromStdin = fromStdin

	x.seekMap = map[int]int64{0: 0}
	x.sizeMap = map[int]int64{0: 0}
	x.mapSiz = 1

	x.newFileSiz = 0
	x.savFileSiz = 0
}

func (x *browseObj) browseInit() {
	x.ignoreCase = false
	x.lastMatch = SEARCH_RESET
	x.hitEOF = false
	x.shownEOF = false
	x.shownMsg = false
	x.shiftWidth = 0
	x.modeNumbers = false
	x.modeScroll = MODE_SCROLL_NONE

	x.saveRC = false
	x.exit = false
}

func (x *browseObj) screenInit(tty *os.File) {
	x.tty = tty

	width, height, err := term.GetSize(int(tty.Fd()))

	if err != nil {
		x.dispWidth = 80
		x.dispHeight = 25
	} else {
		x.dispWidth = width
		x.dispHeight = height
	}

	x.dispRows = x.dispHeight - 1
}

// vim: set ts=4 sw=4 noet:
