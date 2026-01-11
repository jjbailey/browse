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

	br.restoreLast()

	br.mutex.Lock()
	mapSize := br.mapSiz
	br.mutex.Unlock()

	if br.lastRow > mapSize || br.hitEOF {
		// nothing more to show
		return
	}

	for i := 0; i < count && !br.hitEOF; i++ {
		// printLine finds EOF, sets hitEOF
		// add line -- +1 for header
		row := br.lastRow + 1
		if row > br.dispHeight {
			row = br.dispHeight
		}
		moveCursor(row, 1, false)

		if br.shownEOF {
			// print previous line before printing the current line
			fmt.Print(CURRESTORE + CURUP)
			br.printLine(br.lastRow - 1)
			fmt.Print(CURSAVE)
		}

		br.printLine(br.lastRow)

		if br.lastRow >= br.dispRows {
			br.firstRow++
		}

		br.lastRow++
	}

	if br.inMotion() {
		fmt.Print(CURRESTORE)
	} else {
		moveCursor(2, 1, false)
	}
}

func (br *browseObj) scrollUp(count int) {
	// scroll up, toward SOF, stop at SOF

	br.restoreLast()

	if br.firstRow <= 0 {
		br.modeScroll = MODE_SCROLL_NONE
		return
	}

	rowsToScroll := minimum(count, br.firstRow)

	for range rowsToScroll {
		br.firstRow--
		br.lastRow--

		// add line
		fmt.Printf(CURPOS+SCROLLREV, 2, 1)

		// printLine starts with \n
		moveCursor(1, 1, false)
		br.printLine(br.firstRow)
	}

	if !br.inMotion() {
		moveCursor(2, 1, false)
	}
}

func (br *browseObj) tryScroll(sop int) bool {
	// attempt to scroll based on current position and target position

	if sop > br.firstRow {
		if diff := sop - br.firstRow; diff <= br.dispRows>>2 {
			br.scrollDown(diff)
			return true
		}
	} else if br.firstRow > sop {
		if diff := br.firstRow - sop; diff <= br.dispRows>>2 {
			br.scrollUp(diff)
			return true
		}
	}

	return false
}

func (br *browseObj) toggleMode(mode int) {
	// arrows and function keys toggle modes, with some
	// exceptions required for modes to work as users expect

	needsScrollCancel := false

	switch mode {

	case MODE_SCROLL_UP:
		needsScrollCancel = br.modeScroll == mode

	case MODE_SCROLL_DN, MODE_SCROLL_TAIL, MODE_SCROLL_FOLLOW:
		needsScrollCancel = br.inFollow() && (mode == MODE_SCROLL_DN || br.shownEOF)
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

func (br *browseObj) restoreLast() {
	// the search prompt uses the last n lines

	const promptLines = 2

	if !br.shownMsg {
		return
	}

	if br.lastRow < br.dispRows {
		// partial display
		fmt.Printf(CURPOS+CLEARSCREEN, br.dispRows, 1)
	}

	if br.lastRow >= (br.dispHeight - promptLines) {
		// full display
		moveCursor((br.dispHeight - promptLines), 1, false)

		for i := promptLines; i > 0; i-- {
			lineNum := br.lastRow - i
			if lineNum > 0 {
				br.printLine(lineNum)
			}
		}
	}

	if br.inMotion() {
		fmt.Print(CURRESTORE)
	} else {
		moveCursor(2, 1, false)
	}

	br.shownMsg = false
}

// vim: set ts=4 sw=4 noet:
