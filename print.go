// print.go
// printing and some support functions
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"os"
	"strings"
	"sync"
	"time"
)

var lineBufPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

// printLine renders a single line by number, handling SOF/EOF markers.
func (br *browseObj) printLine(lineno int) {
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

	// Use a pooled Builder for line output, reducing allocations and Print calls
	lineBuf := lineBufPool.Get().(*strings.Builder)
	lineBuf.Reset()
	lineBuf.Grow(len(output) + 32)
	lineBuf.WriteByte('\n')
	lineBuf.WriteString(output)
	lineBuf.WriteString(VIDOFF)
	lineBuf.WriteString(CLEARLINE)
	os.Stdout.WriteString(lineBuf.String())
	lineBufPool.Put(lineBuf)

	if br.hitEOF {
		printSEOF("EOF")
	}

	// scrollDown needs this
	br.shownEOF = br.hitEOF
}

// printPage renders a page starting at the provided top line.
func (br *browseObj) printPage(lineno int) {
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

// printCurrentList shows the current browsing file list.
func (br *browseObj) printCurrentList() {
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

// adjustLineNumber clamps a requested top line into a valid range.
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

// timedMessage displays a temporary message on the status line.
func (br *browseObj) timedMessage(msg, color string) {
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

// printMessage displays a message on the status line.
func (br *browseObj) printMessage(msg string, color string) {
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

// vim: set ts=4 sw=4 noet:
