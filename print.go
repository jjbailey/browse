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
	"strings"
	"time"
)

func (br *browseObj) printLine(lineno int) {
	// Check for EOF first for earliest exit opportunity
	isEOF := windowAtEOF(lineno, br.mapSiz)
	br.hitEOF = isEOF
	br.shownEOF = isEOF

	// Handle SOF marker
	if lineno == 0 {
		moveCursor(2, 1, true)
		printSEOF("SOF")
		return
	}

	// Get matches and content from map
	matches, input := br.lineIsMatch(lineno)
	if matches > 0 {
		br.lastMatch = lineno
	}

	// Do not proceed if we're beyond known lines
	if lineno > br.mapSiz {
		return
	}

	// Formatting
	output := br.replaceMatch(lineno, input)

	// Build output using a Builder to reduce stdout calls
	var lineOut strings.Builder
	// rough guess with margin
	lineOut.Grow(len(output) + 16)

	lineOut.WriteString(LINEWRAPOFF)
	lineOut.WriteByte('\n')
	lineOut.WriteString(output)
	lineOut.WriteString(VIDOFF)
	lineOut.WriteString(CLEARLINE)

	fmt.Print(lineOut.String())

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
	maxTopLine := mapSiz - dispRows + 1

	if maxTopLine < 0 {
		maxTopLine = 0
	}

	if lineno > maxTopLine {
		return maxTopLine
	}

	if lineno < 0 {
		return 0
	}

	return lineno
}

func (br *browseObj) timedMessage(msg, color string) {
	// display a short-lived message on the bottom line of the display

	moveCursor(br.dispHeight, 1, true)
	fmt.Print(LINEWRAPOFF)
	fmt.Printf("%s %s %s", color, msg, VIDOFF)
	time.Sleep(1400 * time.Millisecond)
	// scrollDown needs this
	br.shownMsg = true
}

func (br *browseObj) printMessage(msg string, color string) {
	// print a message on the bottom line of the display

	moveCursor(br.dispHeight, 1, true)
	fmt.Print(LINEWRAPOFF)
	fmt.Printf("%s %s %s", color, msg, VIDOFF)
	moveCursor(2, 1, false)

	// scrollDown needs this
	br.shownMsg = true
}

func (br *browseObj) debugPrintf(format string, args ...interface{}) {
	// for debugging

	msg := fmt.Sprintf(format, args...)
	moveCursor(br.dispHeight, 1, true)
	fmt.Print(LINEWRAPOFF)
	fmt.Printf("%s %s %s", _VID_YELLOW_FG, msg, VIDOFF)
	time.Sleep(3 * time.Second)
	// scrollDown needs this
	br.shownMsg = true
}

// vim: set ts=4 sw=4 noet:
