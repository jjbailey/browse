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
	fileInitDefaults(x)

	x.fp = fp
	x.fileName = name
	x.fromStdin = fromStdin
}

func fileInitDefaults(x *browseObj) {
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
	setScreenValues(x, fp, name)
}

func setScreenValues(x *browseObj, fp *os.File, name string) {
	x.tty = fp
	x.title = name

	if width, height, err := term.GetSize(int(x.tty.Fd())); err == nil {
		x.dispWidth = width
		x.dispHeight = height
	} else {
		x.dispWidth = 80
		x.dispHeight = 25
	}

	x.dispRows = x.dispHeight - 1
}

// vim: set ts=4 sw=4 noet:
