// util.go
// various uncategorized functions
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
)

func expandTabs(data []byte) []byte {
	var buf []byte
	var tab []byte

	if !bytes.ContainsAny(data, "\t\r") {
		return data
	}

	for _, b := range data {
		switch b {

		case '\r':
			// Silently map CR to space
			buf = append(buf, ' ')

		case '\t':
			tab = bytes.Repeat([]byte{' '}, TABWIDTH-len(buf)%TABWIDTH)
			buf = append(buf, tab...)

		default:
			buf = append(buf, b)
		}
	}

	return buf
}

func moveCursor(row int, col int, clrflag bool) {
	fmt.Printf(CURPOS, row, col)

	if clrflag {
		fmt.Print(CLEARLINE)
	}
}

func printSEOF(what string) {
	if what == "EOF" {
		// save for modeScroll
		fmt.Printf("\r%s%s", CLEARSCREEN, CURSAVE)
	}

	fmt.Printf("\r %s%s%s\r", VIDBLINK, what, VIDOFF)
}

func windowAtEOF(lineno int, mapsiz int) bool {
	return lineno == mapsiz
}

func maximum(a int, b int) int {
	return int(math.Max(float64(a), float64(b)))
}

func minimum(a int, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

func setScrRegion(top int, bot int) {
	fmt.Printf(SCROLLREGION, top, bot)
}

func resetScrRegion() {
	fmt.Print(RESETREGION)
}

func errorExit(err error) {
	if err != nil {
		fmt.Println(err)
		ttyRestore()
		os.Exit(1)
	}
}

func getMark(buf string) int {
	// scan a mark digit from the buffer
	// valid marks are 1 - 9

	var mark int

	if _, err := fmt.Sscanf(buf, "%d", &mark); err != nil {
		return 0
	}

	mark = int(math.Max(float64(mark), 1))
	mark = int(math.Min(float64(mark), 9))

	return mark
}

// vim: set ts=4 sw=4 noet:
