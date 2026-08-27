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
	"bytes"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// lineBufPool holds output-assembly buffers. bytes.Buffer rather than
// strings.Builder so the assembled bytes can be handed to Write without
// the copy a Builder's String would make.
var lineBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// writeCursorPos appends a CURPOS escape without going through fmt.
func writeCursorPos(buf *bytes.Buffer, row, col int) {
	buf.WriteString("\033[")
	buf.WriteString(strconv.Itoa(row))
	buf.WriteByte(';')
	buf.WriteString(strconv.Itoa(col))
	buf.WriteByte('H')
}

func (br *browseObj) currentMapSize() int {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	return br.mapSiz
}

// currentFileName returns a stable snapshot of br.fileName. The reader
// goroutine mutates fileName under the mutex during file rotation/rescue,
// so command handlers on the main goroutine must read it under the lock.
func (br *browseObj) currentFileName() string {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	return br.fileName
}

func (br *browseObj) setEOFState(hitEOF, shownEOF bool) {
	br.mutex.Lock()
	br.hitEOF = hitEOF
	br.shownEOF = shownEOF
	br.mutex.Unlock()
}

func (br *browseObj) hitEOFState() bool {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	return br.hitEOF
}

func (br *browseObj) shownEOFState() bool {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	return br.shownEOF
}

// printLine renders a single line by number, handling SOF/EOF markers.
func (br *browseObj) printLine(lineno int) {
	br.printLineWithMapSize(lineno, br.currentMapSize())
}

func (br *browseObj) printLineWithMapSize(lineno, mapSize int) {
	lineBuf := lineBufPool.Get().(*bytes.Buffer)
	lineBuf.Reset()
	br.appendLine(lineBuf, lineno, mapSize)
	os.Stdout.Write(lineBuf.Bytes())
	lineBufPool.Put(lineBuf)
}

// appendLine renders one line into buf instead of writing it directly, so
// callers that emit many lines at once (printPage, scrollUp) can issue a
// single write for the whole batch.
func (br *browseObj) appendLine(buf *bytes.Buffer, lineno, mapSize int) {
	isEOF := windowAtEOF(lineno, mapSize)
	br.setEOFState(isEOF, isEOF)

	// Handle SOF marker
	if lineno == 0 {
		writeCursorPos(buf, 2, 1)
		buf.WriteString(CLEARLINE)
		appendSEOF(buf, "SOF")
		return
	}

	// Do not proceed if we're beyond known lines
	if lineno > mapSize {
		buf.WriteString(CLEARLINE)
		return
	}

	// Get content from map. Search owns br.lastMatch; rendering must not move it.
	// Read directly: replaceMatch re-runs the regex itself, so a match here
	// would be thrown away.
	input := br.readFromMap(lineno)

	output := br.replaceMatch(lineno, input)

	buf.Grow(len(output) + 32)
	buf.WriteByte('\n')
	buf.WriteString(output)
	buf.WriteString(VIDOFF)
	buf.WriteString(CLEARLINE)

	if isEOF {
		appendSEOF(buf, "EOF")
	}
}

// postMessage queues a status-line message for the main goroutine to
// display. Only the main goroutine draws; other goroutines (the file
// reader, the decompressor waiter) post display work under the mutex
// and the commands loop applies it on its next tick.
func (br *browseObj) postMessage(msg, color string) {
	br.mutex.Lock()
	br.pendingMsg = msg
	br.pendingMsgColor = color
	br.mutex.Unlock()
}

// drainDisplayEvents applies display work posted by other goroutines.
// Call only from the main goroutine.
func (br *browseObj) drainDisplayEvents() {
	br.mutex.Lock()
	msg, color := br.pendingMsg, br.pendingMsgColor
	transient := br.pendingMsgTransient
	refresh := br.refreshPending
	cancelScroll := br.scrollCancelPending
	resize := br.resizePending
	br.pendingMsg = ""
	br.pendingMsgTransient = false
	br.refreshPending = false
	br.scrollCancelPending = false
	br.resizePending = false
	br.mutex.Unlock()

	if cancelScroll {
		br.modeScroll = MODE_SCROLL_NONE
	}
	if resize {
		br.resizeWindow()
	}

	// Transient notices (e.g. "Re-reading file") print first and get a
	// moment on screen, then the refresh below naturally overwrites the
	// same status row with the real last line of the page.
	if msg != "" && transient {
		br.printMessage(msg, color)
		if refresh {
			time.Sleep(1500 * time.Millisecond)
		}
	}

	if refresh {
		br.pageCurrent()
	}

	// non-transient messages print last so they survive the refresh
	if msg != "" && !transient {
		br.printMessage(msg, color)
	}
}

// printPage renders a page starting at the provided top line.
func (br *browseObj) printPage(lineno int) {
	mapSize := br.currentMapSize()

	lineno = adjustLineNumber(lineno, br.dispRows, mapSize)
	sop := lineno
	// +1 for EOF
	eop := minimum(sop+br.dispRows, mapSize+1)

	if mapSize > br.dispRows && br.tryScroll(sop) {
		return
	}

	// Assemble the whole page in one buffer and emit it with a single
	// write; the terminal sees the identical byte stream either way.
	pageBuf := lineBufPool.Get().(*bytes.Buffer)
	pageBuf.Reset()

	// Only one cursor move here for all lines
	// appendLine starts with \n
	writeCursorPos(pageBuf, 1, 1)
	for i := sop; i < eop; i++ {
		br.appendLine(pageBuf, i, mapSize)
	}

	// reset
	pageBuf.WriteString(SGR0)
	writeCursorPos(pageBuf, 2, 1)
	os.Stdout.Write(pageBuf.Bytes())
	lineBufPool.Put(pageBuf)

	// reset these
	br.firstRow, br.lastRow = sop, eop
}

// printCurrentList shows the current browsing file list.
func (br *browseObj) printCurrentList() {
	var sb strings.Builder

	// Leave room for ellipsis (3 chars)
	maxLen := br.dispWidth - 8

	for i, name := range br.currentList {
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

// printBrowseStack shows the current file followed by suspended parent files.
func (br *browseObj) printBrowseStack() {
	br.printMessage(br.browseStackText(), MSG_GREEN)
}

// browseStackText returns the current file followed by suspended, resumable
// parent files, with the immediately resumable parent shown first. Its
// formatting matches the current-list display used by the a command.
func (br *browseObj) browseStackText() string {
	var sb strings.Builder

	// Match printCurrentList: reserve room for an ellipsis when needed.
	maxLen := br.dispWidth - 8
	sb.WriteString("[")
	sb.WriteString(br.currentFileName())
	sb.WriteString("]")

	for i := len(br.browseStack) - 1; i >= 0; i-- {
		if br.browseStack[i].fromStdin {
			continue
		}
		name := " " + br.browseStack[i].fileName
		if sb.Len()+len(name) > maxLen {
			sb.WriteString(" ...")
			break
		}
		sb.WriteString(name)
	}

	return sb.String()
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
