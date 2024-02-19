// print.go
// printing and some support functions
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
	"time"
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

	if lineno <= x.mapSiz {
		// replaceMatch adds line numbers if applicable
		output := x.replaceMatch(lineno, input)
		// depends on linewrap=false
		fmt.Printf("\r\n%s%s%s\r", output, VIDOFF, CLEARLINE)
	}

	if windowAtEOF(lineno, x.mapSiz) {
		x.hitEOF = true
		printSEOF("EOF")
	} else {
		x.hitEOF = false
	}

	// scrollDown needs this
	x.shownEOF = x.hitEOF
}

func (x *browseObj) printPage(lineno int) {
	// print a page -- full screen if possible
	// lineno is the top line

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
	} else if lineno > x.mapSiz {
		lineno = x.mapSiz
	}

	sop := lineno
	// +1 for EOF
	eop := minimum((sop + x.dispRows), x.mapSiz+1)

	// scroll if
	//   - more than one page of data
	//   - current position is <= 1/4 page to target
	if x.mapSiz > x.dispRows {
		if sop > x.firstRow && sop-x.firstRow <= (x.dispRows>>2) {
			x.scrollDown(sop - x.firstRow)
			return
		} else if x.firstRow > sop && x.firstRow-sop <= (x.dispRows>>2) {
			x.scrollUp(x.firstRow - sop)
			return
		}
	}

	fmt.Print(LINEWRAPOFF)
	movecursor(2, 1, false)

	for i = sop; i < eop; i++ {
		x.printLine(i)
	}

	movecursor(2, 1, false)

	// reset these
	x.firstRow = sop
	x.lastRow = i
}

func (x *browseObj) timedMessage(msg string) {
	x.printMessage(msg)
	// sleep time is arbitrary
	time.Sleep(1250 * time.Millisecond)
}

// func (x *browseObj) debugMessage(msg string) {
// 	x.printMessage(msg)
// 	// sleep time is arbitrary
// 	time.Sleep(5 * time.Second)
// }

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

	if x.shownMsg {
		movecursor(x.dispHeight, 1, true)

		if x.lastRow > x.dispHeight {
			fmt.Print(CURUP)
			x.printLine(x.lastRow - 1)
		}

		fmt.Print(CURRESTORE)
		x.shownMsg = false
	}
}

// vim: set ts=4 sw=4 noet:
