// scroll.go
// scrolling functions
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
)

func (br *browseObj) scrollDown(count int) {
	// scroll down, toward EOF, stop at EOF
	// there's more hand-waving here than meets the eye

	if br.shownMsg {
		// the last line contains a message
		br.restoreLast()
	}

	if br.lastRow > br.mapSiz || br.hitEOF {
		// nothing more to show
		return
	}

	for i := 0; i < count && !br.hitEOF; i++ {
		// printLine finds EOF, sets hitEOF
		// add line -- +1 for header
		moveCursor(br.lastRow+1, 1, false)

		if br.shownEOF {
			// print previous line before printing the current line
			fmt.Printf("%s%s", CURRESTORE, CURUP)
			br.printLine(br.lastRow - 1)
		}

		br.printLine(br.lastRow)

		if br.lastRow >= br.dispRows {
			br.firstRow++
		}

		br.lastRow++
	}

	if br.inMotion() {
		// in one of the follow modes
		fmt.Print(CURRESTORE)
	} else {
		// idle
		moveCursor(2, 1, false)
	}
}

func (br *browseObj) scrollUp(count int) {
	if br.firstRow <= 0 {
		br.modeScroll = MODE_SCROLL_NONE
		return
	}

	rowsToScroll := minimum(count, br.firstRow)

	if br.shownEOF {
		// cursor is on the bottom line
		moveCursor(2, 1, false)
	}

	for i := 0; i < rowsToScroll; i++ {
		br.firstRow--
		br.lastRow--

		// add line
		fmt.Print(SCROLLREV)
		moveCursor(1, 1, false)
		br.printLine(br.firstRow)
	}

	if !br.inMotion() {
		moveCursor(2, 1, false)
	}
}

func (br *browseObj) toggleMode(mode int) {
	// arrows and function keys toggle modes, with some
	// exceptions required for modes to work as users expect

	needsScrollCancel := false

	switch mode {

	case MODE_SCROLL_UP:
		// toggle
		needsScrollCancel = br.modeScroll == mode

	case MODE_SCROLL_DN:
		// cancel down or either follow mode, else start this mode
		needsScrollCancel = br.inFollow()

	case MODE_SCROLL_TAIL, MODE_SCROLL_FOLLOW:
		// cancel either follow mode at EOF, else start this mode
		needsScrollCancel = br.inFollow() && br.shownEOF
	}

	if needsScrollCancel {
		br.modeScroll = MODE_SCROLL_NONE
	} else {
		br.modeScroll = mode
	}
}

func (br *browseObj) inMotion() bool {
	return br.modeScroll != MODE_SCROLL_NONE
}

func (br *browseObj) inFollow() bool {
	return (br.modeScroll == MODE_SCROLL_DN ||
		br.modeScroll == MODE_SCROLL_TAIL ||
		br.modeScroll == MODE_SCROLL_FOLLOW)
}

// vim: set ts=4 sw=4 noet:
