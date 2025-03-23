// print.go
// printing and some support functions
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"time"
)

func (br *browseObj) printLine(lineno int) {
	// print a line from the map, finds EOF, sets hitEOF

	br.hitEOF = windowAtEOF(lineno, br.mapSiz)

	if lineno == 0 {
		moveCursor(2, 1, true)
		printSEOF("SOF")
		return
	}

	// lineIsMatch reads lines from the map
	matches, input := br.lineIsMatch(lineno)

	if matches > 0 {
		// where to start search
		br.lastMatch = lineno
	}

	if lineno <= br.mapSiz {
		// replaceMatch adds line numbers if applicable
		output := br.replaceMatch(lineno, input)
		// depends on linewrap=false
		fmt.Printf("\n%s%s%s", output, VIDOFF, CLEARLINE)
	}

	if lineno < br.dispRows {
		// save cursor when screen is not full
		fmt.Printf("\r%s", CURSAVE)
	}

	if br.hitEOF {
		printSEOF("EOF")
	}

	// scrollDown needs this
	br.shownEOF = br.hitEOF
}

func (br *browseObj) printPage(lineno int) {
	// print a page -- full screen if possible
	// lineno is the top line

	lineno = adjustLineNumber(lineno, br.dispRows, br.mapSiz)

	sop := lineno
	// +1 for EOF
	eop := minimum(sop+br.dispRows, br.mapSiz+1)

	if br.mapSiz > br.dispRows {
		if br.tryScroll(sop) {
			return
		}
	}

	fmt.Print(LINEWRAPOFF)
	// printLine starts with \n
	moveCursor(1, 1, false)

	for i := sop; i < eop; i++ {
		br.printLine(i)
	}

	moveCursor(2, 1, false)

	// reset these
	br.firstRow, br.lastRow = sop, eop
}

func adjustLineNumber(lineno, dispRows, mapSiz int) int {
	if lineno+dispRows > mapSiz {
		lineno = mapSiz - dispRows + 1
	}

	// Ensure lineno is within valid range
	if lineno < 0 {
		return 0
	} else if lineno > mapSiz {
		return mapSiz
	}

	return lineno
}

func (br *browseObj) tryScroll(sop int) bool {
	// attempt to scroll based on current position and target position

	if sop > br.firstRow && sop-br.firstRow <= br.dispRows>>2 {
		br.scrollDown(sop - br.firstRow)
		return true
	} else if br.firstRow > sop && br.firstRow-sop <= br.dispRows>>2 {
		br.scrollUp(br.firstRow - sop)
		return true
	}

	return false
}

func (br *browseObj) timedMessage(msg string, color string) {
	br.printMessage(msg, color)
	// sleep time is arbitrary
	time.Sleep(1500 * time.Millisecond)
}

func (br *browseObj) printMessage(msg string, color string) {
	// print a message on the bottom line of the display

	moveCursor(br.dispHeight, 1, true)
	fmt.Printf("%s %s %s", color, msg, VIDOFF)
	fmt.Print(LINEWRAPOFF)
	moveCursor(2, 1, false)

	// scrollDown needs this
	br.shownMsg = true
}

func (br *browseObj) restoreLast() {
	// restore the last (prompt) line

	if br.shownMsg {
		moveCursor(br.dispHeight, 1, true)

		// -2 for SOF, EOF
		if br.lastRow > (br.dispHeight - 2) {
			fmt.Print(CURUP)
			br.printLine(br.lastRow - 1)
		}

		fmt.Print(CURRESTORE)
		br.shownMsg = false
	}
}

// vim: set ts=4 sw=4 noet:
