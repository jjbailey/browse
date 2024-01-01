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
	// scroll down, toward EOF
	// there's more hand-waving here than meets the eye

	if x.lastRow > x.mapSiz {
		return
	}

	if x.shownMsg {
		// the last line contains a message
		x.restoreLast()
	}

	for i := 0; i < count; i++ {
		if x.hitEOF {
			break
		}

		// add line
		// printLine finds EOF, sets hitEOF
		// printLine starts with \r\n
		// +1 for header
		movecursor(x.lastRow+1, 1, false)

		if x.shownEOF {
			fmt.Printf("%s", CURRESTORE)
			fmt.Printf("%s", CURUP)
			x.printLine(x.lastRow - 1)
		}

		x.printLine(x.lastRow)

		x.firstRow++
		x.lastRow++

		if x.lastRow == x.mapSiz {
			break
		}
	}

	if (x.modeScrollDown && x.hitEOF) || x.modeTail {
		// in tail mode
		fmt.Printf("%s", CURRESTORE)
	} else {
		// idle
		movecursor(2, 1, false)
	}
}

func (x *browseObj) scrollUp(count int) {
	// scroll up, toward SOF

	if x.firstRow <= 0 {
		return
	}

	for i := 0; i < count; i++ {
		x.firstRow--
		x.lastRow--

		// scroll
		movecursor(2, 1, false)
		fmt.Printf("%s", SCROLLREV)

		// add line
		// printLine starts with \r\n
		movecursor(1, 1, false)
		x.printLine(x.firstRow)

		if x.firstRow == 0 {
			break
		}
	}
}

// vim: set ts=4 sw=4 noet:
