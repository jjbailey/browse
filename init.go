// init.go
// functions to initialize objects
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"os"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// fileInit initializes browseObj state for a new file.
func (br *browseObj) fileInit(fp *os.File, fileName, title string, fromStdin bool) {
	if br.rescueFd > 0 {
		_ = unix.Close(br.rescueFd)
		br.rescueFd = 0
		br.fdLink = ""
	}
	br.fileName = fileName
	br.fp = fp
	br.fromStdin = fromStdin
	br.lastMatch = SEARCH_RESET

	if br.initTitle != "" {
		// one-time use of -t option
		br.title = br.initTitle
		br.initTitle = ""
	} else {
		br.title = title
	}
}

// screenInit initializes terminal sizing information.
func (br *browseObj) screenInit(tty *os.File) {
	br.tty = tty

	var width, height int
	var err error

	if tty != nil {
		width, height, err = term.GetSize(int(tty.Fd()))
	}

	if tty == nil || err != nil {
		br.dispWidth = 80
		br.dispHeight = 25
	} else {
		br.dispWidth = width
		br.dispHeight = height
	}

	br.dispRows = br.dispHeight - 1
}

// vim: set ts=4 sw=4 noet:
