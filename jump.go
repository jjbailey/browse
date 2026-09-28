// jump.go
// Jump to a location in a file
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"strconv"
	"strings"
)

// Jump to line
func jumpLine(br *browseObj) {
	lbuf, cancelled := br.userInput("Jump: ")
	if !cancelled && len(lbuf) > 0 {
		n, err := strconv.Atoi(strings.TrimSpace(lbuf))
		if err != nil {
			br.printMessage("Invalid line number", MSG_ORANGE)
		} else if n < 0 {
			br.printMessage("Line number must be positive", MSG_ORANGE)
		} else {
			br.printPage(n)
		}
	}
}

// lineForOffset returns the line containing offset. seekMap includes the EOF
// sentinel, so the last usable line is len(seekMap)-2.
func lineForOffset(offset int64, seekMap []int64) int {
	lastLine := len(seekMap) - 2
	if lastLine < 1 {
		return 1
	}

	// find the line containing the offset
	lo, hi := 0, len(seekMap)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if seekMap[mid] <= offset {
			lo = mid + 1
		} else {
			hi = mid
		}
	}

	line := lo - 1
	if line < 1 {
		return 1
	}
	if line > lastLine {
		return lastLine
	}

	return line
}

// Jump to character position
func jumpPosition(br *browseObj) {
	lbuf, cancelled := br.userInput("Jump to position: ")
	if !cancelled && len(lbuf) > 0 {
		n, err := strconv.Atoi(strings.TrimSpace(lbuf))
		if err != nil {
			br.printMessage("Invalid position", MSG_ORANGE)
		} else if n < 0 {
			br.printMessage("Position must be positive", MSG_ORANGE)
		} else {
			// look in the map for the line number containing character n
			br.mutex.Lock()
			if br.fp == nil || br.mapSiz == 0 {
				br.mutex.Unlock()
				br.printMessage("File not loaded", MSG_ORANGE)
				return
			}

			// seekMap holds the starting offset of each line in ascending
			// order; find the last line whose start offset is <= n.
			lineno := lineForOffset(int64(n), br.seekMap[:br.mapSiz])
			br.mutex.Unlock()

			br.printPage(lineno)
		}
	}
}

// vim: set ts=4 sw=4 noet:
