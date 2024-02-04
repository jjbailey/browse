// scroll.go
// scrolling functions
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
)

func (x *browseObj) scrollDown(count int) {
	// scroll down, toward EOF, stop at EOF
	// there's more hand-waving here than meets the eye

	if x.shownMsg {
		// the last line contains a message
		x.restoreLast()
	}

	if x.lastRow > x.mapSiz {
		// nothing more to show
		return
	}

	for i := 0; i < count; i++ {
		// printLine finds EOF, sets hitEOF

		if x.hitEOF {
			break
		}

		// add line -- +1 for header
		movecursor(x.lastRow+1, 1, false)

		if x.shownEOF {
			// print previous line before printing the current line
			fmt.Printf("%s%s", CURRESTORE, CURUP)
			x.printLine(x.lastRow - 1)
		}

		x.printLine(x.lastRow)

		x.firstRow++
		x.lastRow++
	}

	if x.modeScrollDown || x.modeTail {
		// in one of the follow modes
		fmt.Print(CURRESTORE)
	} else {
		// idle
		movecursor(2, 1, false)
	}
}

func (x *browseObj) scrollUp(count int) {
	// scroll up, toward SOF, stop at SOF

	if x.firstRow <= 0 {
		return
	}

	for i := 0; i < count && x.firstRow > 0; i++ {
		x.firstRow--
		x.lastRow--

		// scroll
		movecursor(2, 1, false)
		fmt.Print(SCROLLREV)

		// add line
		movecursor(1, 1, false)
		x.printLine(x.firstRow)
	}
}

// vim: set ts=4 sw=4 noet:
