// init.go
// functions to initialize objects
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"os"

	"golang.org/x/term"
)

func (x *browseObj) fileInit(fp *os.File, name string, fromStdin bool) {
	x.fp = fp
	x.fileName = name
	x.fromStdin = fromStdin

	x.seekMap = make(map[int]int64, 1)
	x.sizeMap = make(map[int]int64, 1)
	x.mapSiz = 1

	x.lastMatch = SEARCH_RESET
	x.hitEOF = false
	x.shownEOF = false
	x.shownMsg = false
	x.saveRC = false

	x.seekMap[0] = 0
	x.sizeMap[0] = 0
	x.shiftWidth = 0

	x.modeNumbers = false
	x.modeScrollDown = false
}

func (x *browseObj) screenInit(fp *os.File, name string) {
	x.tty = fp
	x.screenName = name

	x.dispWidth, x.dispHeight, _ = term.GetSize(int(x.tty.Fd()))
	x.dispRows = x.dispHeight - 1
}

// vim: set ts=4 sw=4 noet:
