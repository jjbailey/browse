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

	// Do not proceed if we're beyond known lines
	if lineno > br.mapSiz {
		fmt.Print(CLEARLINE)
		return
	}

	// Get matches and content from map
	matches, input := br.lineIsMatch(lineno)
	if matches > 0 {
		br.lastMatch = lineno
	}

	output := br.replaceMatch(lineno, input)

	// Use a Builder for line output, reducing allocations and Print calls
	lineLen := len(output)
	const extra = 32 // Extra slack for control codes
	var sb strings.Builder
	sb.Grow(lineLen + extra)
	sb.WriteString(LINEWRAPOFF)
	sb.WriteByte('\n')
	sb.WriteString(output)
	sb.WriteString(VIDOFF)
	sb.WriteString(CLEARLINE)
	fmt.Print(sb.String())

	if br.hitEOF {
		printSEOF("EOF")
	}

	// scrollDown needs this
	br.shownEOF = br.hitEOF
}

func (br *browseObj) printPage(lineno int) {
	// print/refresh a page -- full screen if possible
	// lineno is the top line

	lineno = adjustLineNumber(lineno, br.dispRows, br.mapSiz)
	sop := lineno
	// +1 for EOF
	eop := minimum(sop+br.dispRows, br.mapSiz+1)

	if br.mapSiz > br.dispRows && br.tryScroll(sop) {
		return
	}

	// Only one cursor move here for all lines
	// printLine starts with \n
	moveCursor(1, 1, false)
	for i := sop; i < eop; i++ {
		br.printLine(i)
	}

	// reset these
	br.firstRow, br.lastRow = sop, eop
	moveCursor(2, 1, false)
}

func (br *browseObj) printCurrentList() {
	// Print the arg list

	line := ""

	// Leave room for ellipsis (3 chars)
	maxLen := br.dispWidth - 8

	for i, name := range CurrentList {
		var part string

		if i == 0 {
			part = "[" + name + "]"
		} else {
			part = " " + name
		}

		if len(line)+len(part) > maxLen {
			line += " " + "..."
			break
		}

		line += part
	}

	br.printMessage(line, MSG_GREEN)
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
	var sb strings.Builder
	sb.Grow(len(msg) + len(color) + len(VIDOFF) + 8)
	sb.WriteString(LINEWRAPOFF)
	sb.WriteString(color)
	sb.WriteByte(' ')
	sb.WriteString(msg)
	sb.WriteByte(' ')
	sb.WriteString(VIDOFF)
	fmt.Print(sb.String())
	time.Sleep(1500 * time.Millisecond)
	// scrollDown needs this
	br.shownMsg = true
}

func (br *browseObj) printMessage(msg string, color string) {
	// print a message on the bottom line of the display

	moveCursor(br.dispHeight, 1, true)
	var sb strings.Builder
	sb.Grow(len(msg) + len(color) + len(VIDOFF) + 8)
	sb.WriteString(LINEWRAPOFF)
	sb.WriteString(color)
	sb.WriteByte(' ')
	sb.WriteString(msg)
	sb.WriteByte(' ')
	sb.WriteString(VIDOFF)
	fmt.Print(sb.String())
	moveCursor(2, 1, false)
	// scrollDown needs this
	br.shownMsg = true
}

func (br *browseObj) debugPrintf(format string, args ...any) {
	// for debugging

	moveCursor(br.dispHeight, 1, true)
	msg := fmt.Sprintf(format, args...)
	var sb strings.Builder
	sb.Grow(len(msg) + len(_VID_YELLOW_FG) + len(VIDOFF) + 8)
	sb.WriteString(LINEWRAPOFF)
	sb.WriteString(_VID_YELLOW_FG)
	sb.WriteByte(' ')
	sb.WriteString(msg)
	sb.WriteByte(' ')
	sb.WriteString(VIDOFF)
	fmt.Print(sb.String())
	time.Sleep(3 * time.Second)
	// scrollDown needs this
	br.shownMsg = true
}

// vim: set ts=4 sw=4 noet:
