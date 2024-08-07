// scroll.go
// scrolling functions
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

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

	if x.lastRow > x.mapSiz || x.hitEOF {
		// nothing more to show
		return
	}

	for i := 0; i < count && !x.hitEOF; i++ {
		// printLine finds EOF, sets hitEOF
		// add line -- +1 for header
		moveCursor(x.lastRow+1, 1, false)

		if x.shownEOF {
			// print previous line before printing the current line
			fmt.Printf("%s%s", CURRESTORE, CURUP)
			x.printLine(x.lastRow - 1)
		}

		x.printLine(x.lastRow)

		if x.lastRow >= x.dispRows {
			x.firstRow++
		}

		x.lastRow++
	}

	if x.inMotion() {
		// in one of the follow modes
		fmt.Print(CURRESTORE)
	} else {
		// idle
		moveCursor(2, 1, false)
	}
}

func (x *browseObj) scrollUp(count int) {
	if x.firstRow <= 0 {
		x.modeScroll = MODE_SCROLL_NONE
		return
	}

	rowsToScroll := minimum(count, x.firstRow)

	if x.shownEOF {
		// cursor is on the bottom line
		moveCursor(2, 1, false)
	}

	for i := 0; i < rowsToScroll; i++ {
		x.firstRow--
		x.lastRow--

		// add line
		fmt.Print(SCROLLREV)
		moveCursor(1, 1, false)
		x.printLine(x.firstRow)
	}

	if !x.inMotion() {
		moveCursor(2, 1, false)
	}
}

func (x *browseObj) toggleMode(mode int) {
	// arrows and function keys toggle modes, with some
	// exceptions required for modes to work as users expect

	needsScrollCancel := false

	switch mode {

	case MODE_SCROLL_UP:
		// toggle
		needsScrollCancel = x.modeScroll == mode

	case MODE_SCROLL_DN:
		// cancel down or either follow mode, else start this mode
		needsScrollCancel = x.inFollow()

	case MODE_SCROLL_TAIL, MODE_SCROLL_FOLLOW:
		// cancel either follow mode at EOF, else start this mode
		needsScrollCancel = x.inFollow() && x.shownEOF
	}

	if needsScrollCancel {
		x.modeScroll = MODE_SCROLL_NONE
	} else {
		x.modeScroll = mode
	}
}

func (x *browseObj) inMotion() bool {
	return x.modeScroll != MODE_SCROLL_NONE
}

func (x *browseObj) inFollow() bool {
	return (x.modeScroll == MODE_SCROLL_DN ||
		x.modeScroll == MODE_SCROLL_TAIL ||
		x.modeScroll == MODE_SCROLL_FOLLOW)
}

// vim: set ts=4 sw=4 noet:
