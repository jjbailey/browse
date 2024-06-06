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

func (x *browseObj) fileInit(fp *os.File, name string, fromStdin bool) {
	x.seekMap = map[int]int64{0: 0}
	x.sizeMap = map[int]int64{0: 0}
	x.mapSiz = 1

	x.ignoreCase = false
	x.lastMatch = SEARCH_RESET
	x.hitEOF = false
	x.shownEOF = false
	x.shownMsg = false
	x.saveRC = false

	x.shiftWidth = 0

	x.modeNumbers = false
	x.modeScroll = MODE_SCROLL_NONE

	x.fp = fp
	x.fileName = name
	x.fromStdin = fromStdin
}

func (x *browseObj) screenInit(fp *os.File, name string) {
	x.tty = fp
	x.title = name

	width, height, err := term.GetSize(int(x.tty.Fd()))

	if err != nil {
		x.dispWidth = 80
		x.dispHeight = 25
	} else {
		x.dispWidth, x.dispHeight = width, height
	}

	x.dispRows = x.dispHeight - 1
}

// vim: set ts=4 sw=4 noet:
