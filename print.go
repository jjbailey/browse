// print.go
// printing and some support functions
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
)

func (x *browseObj) printLine(lineno int) {
	var prLine string

	if lineno == 0 {
		// no line numbers
		movecursor(2, 1, true)
		printSEOF("SOF")
		return
	}

	data, lineSiz := x.readFromMap(lineno)

	if !x.modeNumbers || windowAtEOF(lineno, x.mapSiz) {
		// no line numbers
		prLine = string(data)
	} else {
		// 7 columns for line numbers
		prLine = fmt.Sprintf("%6d %s", lineno, string(data))
		lineSiz += 7
	}

	fmt.Printf("\r\n%.*s%s\r", minimum(lineSiz, x.dispWidth), prLine, CLEARLINE)

	if windowAtEOF(lineno, x.mapSiz) {
		printSEOF("EOF")
		x.hitEOF = true
		x.shownEOF = true
	} else {
		x.hitEOF = false
		x.shownEOF = false
	}
}

func (x *browseObj) printMessage(msg string) {
	// print a message on the bottom line of the display

	movecursor(x.dispHeight, 1, true)
	fmt.Printf("%s %s %s", VIDBOLDREV, msg, VIDOFF)
	movecursor(2, 1, false)

	// scrollDown needs this
	x.shownMsg = true
}

func (x *browseObj) printPage(lineno int) {
	var eop int
	var i int

	if lineno+x.dispRows > x.mapSiz {
		lineno -= (lineno - x.mapSiz)
		lineno -= x.dispRows
		// +1 for EOF
		lineno++
	}

	if lineno < 0 {
		lineno = 0
	}

	// (re)set some vars
	x.firstRow = lineno
	x.hitEOF = false

	// +1 for EOF
	eop = minimum((x.firstRow + x.dispRows), x.mapSiz+1)

	movecursor(2, 1, false)

	for i = x.firstRow; i < eop; i++ {
		// printLine finds EOF, sets hitEOF
		x.printLine(i)
	}

	x.lastRow = i
	movecursor(2, 1, false)
}

func (x *browseObj) restoreLast() {
	// restore the last (prompt) line
	// printLine starts with \r\n

	movecursor(x.dispRows, 1, false)
	x.printLine(x.lastRow - 1)
	movecursor(2, 1, false)
	x.shownMsg = false
}

// vim: set ts=4 sw=4 noet:
