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
	"os"
	"strings"
	"time"
)

func (br *browseObj) printLine(lineno int) {
	// Check for EOF first for earliest exit opportunity
	br.mutex.Lock()
	mapSize := br.mapSiz
	br.mutex.Unlock()

	isEOF := windowAtEOF(lineno, mapSize)
	br.hitEOF = isEOF
	br.shownEOF = isEOF

	// Handle SOF marker
	if lineno == 0 {
		moveCursor(2, 1, true)
		printSEOF("SOF")
		return
	}

	// Do not proceed if we're beyond known lines
	if lineno > mapSize {
		os.Stdout.WriteString(CLEARLINE)
		return
	}

	// Get matches and content from map
	matches, input := br.lineIsMatch(lineno)
	if matches > 0 {
		br.lastMatch = lineno
	}

	output := br.replaceMatch(lineno, input)

	// Use a Builder for line output, reducing allocations and Print calls
	var sb strings.Builder
	sb.Grow(len(output) + 32)
	sb.WriteByte('\n')
	sb.WriteString(output)
	sb.WriteString(VIDOFF)
	sb.WriteString(CLEARLINE)
	os.Stdout.WriteString(sb.String())

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

	var sb strings.Builder

	// Leave room for ellipsis (3 chars)
	maxLen := br.dispWidth - 8

	for i, name := range CurrentList {
		// Add brackets to the current file in the list
		if i == 0 {
			name = "[" + name + "]"
		} else {
			name = " " + name
		}

		if sb.Len()+len(name) > maxLen {
			sb.WriteString(" ...")
			break
		}

		sb.WriteString(name)
	}

	br.printMessage(sb.String(), MSG_GREEN)
}

func adjustLineNumber(lineno, dispRows, mapSiz int) int {
	if lineno < 0 {
		return 0
	}

	if mapSiz < dispRows {
		return 0
	}

	maxTopLine := mapSiz - dispRows + 1
	if lineno > maxTopLine {
		return maxTopLine
	}

	return lineno
}

func (br *browseObj) timedMessage(msg, color string) {
	// display a short-lived message on the bottom line of the display

	moveCursor(br.dispHeight, 1, true)
	var sb strings.Builder
	sb.Grow(len(msg) + len(color) + len(VIDOFF) + 8)
	sb.WriteString(color)
	sb.WriteByte(' ')
	sb.WriteString(msg)
	sb.WriteByte(' ')
	sb.WriteString(VIDOFF)
	os.Stdout.WriteString(sb.String())
	time.Sleep(1500 * time.Millisecond)
	// scrollDown needs this
	br.shownMsg = true
}

func (br *browseObj) printMessage(msg string, color string) {
	// print a message on the bottom line of the display

	moveCursor(br.dispHeight, 1, true)
	var sb strings.Builder
	sb.Grow(len(msg) + len(color) + len(VIDOFF) + 8)
	sb.WriteString(color)
	sb.WriteByte(' ')
	sb.WriteString(msg)
	sb.WriteByte(' ')
	sb.WriteString(VIDOFF)
	os.Stdout.WriteString(sb.String())
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
	sb.WriteString(_VID_YELLOW_FG)
	sb.WriteByte(' ')
	sb.WriteString(msg)
	sb.WriteByte(' ')
	sb.WriteString(VIDOFF)
	os.Stdout.WriteString(sb.String())
	time.Sleep(3 * time.Second)
	// scrollDown needs this
	br.shownMsg = true
}

// vim: set ts=4 sw=4 noet:
