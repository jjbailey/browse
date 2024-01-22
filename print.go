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
	// print a line from the map, finds EOF, sets hitEOF

	if lineno == 0 {
		movecursor(2, 1, true)
		printSEOF("SOF")
		return
	}

	// lineIsMatch reads lines from the map
	matches, input := x.lineIsMatch(lineno)

	if matches > 0 {
		// where to start search
		x.lastMatch = lineno
	}

	// replaceMatch adds line numbers if applicable
	output := x.replaceMatch(lineno, input)

	// depends on linewrap=false
	fmt.Printf("\r\n%s%s%s\r", output, VIDOFF, CLEARLINE)

	if windowAtEOF(lineno, x.mapSiz) {
		printSEOF("EOF")
		x.hitEOF = true
	} else {
		x.hitEOF = false
	}

	// scrollDown needs this
	x.shownEOF = x.hitEOF
}

func (x *browseObj) printPage(lineno int) {
	// print a page -- full screen if possible

	var i int

	if lineno+x.dispRows > x.mapSiz {
		// beyond EOF
		lineno -= (lineno - x.mapSiz)
		lineno -= x.dispRows
		// +1 for EOF
		lineno++
	}

	if lineno < 0 {
		lineno = 0
	}

	// +1 for EOF
	sop := lineno
	eop := minimum((sop + x.dispRows), x.mapSiz+1)

	// scroll if less than 1/4 page to target
	if sop > x.firstRow && sop-x.firstRow < (x.dispRows>>2) {
		x.scrollDown(sop - x.firstRow)
		return
	} else if x.firstRow > sop && x.firstRow-sop < (x.dispRows>>2) {
		x.scrollUp(x.firstRow - sop)
		return
	}

	fmt.Printf("%s", LINEWRAPOFF)
	movecursor(2, 1, false)

	for i = sop; i < eop; i++ {
		x.printLine(i)
	}

	movecursor(2, 1, false)

	// reset these
	x.firstRow = sop
	x.lastRow = i
}

func (x *browseObj) printMessage(msg string) {
	// print a message on the bottom line of the display

	movecursor(x.dispHeight, 1, true)
	fmt.Printf("%s %s %s", VIDMESSAGE, msg, VIDOFF)
	movecursor(2, 1, false)

	// scrollDown needs this
	x.shownMsg = true
}

func (x *browseObj) restoreLast() {
	// restore the last (prompt) line

	movecursor(x.dispRows, 1, false)
	x.printLine(x.lastRow - 1)
	fmt.Printf("%s", CURRESTORE)
	x.shownMsg = false
}

// vim: set ts=4 sw=4 noet:
