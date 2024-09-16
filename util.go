// util.go
// various uncategorized functions
//
// Copyright (c) 2024 jjb
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
	"strings"
)

func expandTabs(data []byte) []byte {
	var output []byte

	if !strings.ContainsAny(string(data), "\t\r") {
		return data
	}

	for _, char := range data {
		switch char {

		case '\r':
			// silently map CR to space
			output = append(output, ' ')

		case '\t':
			spaceCount := TABWIDTH - (len(output) % TABWIDTH)
			output = append(output, bytes.Repeat([]byte{' '}, spaceCount)...)

		default:
			output = append(output, char)
		}

		if len(output) == cap(output) {
			newOutput := make([]byte, len(output), 2*len(output))
			copy(newOutput, output)
			output = newOutput
		}
	}

	return output
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

	var d int

	if _, err := fmt.Sscanf(buf, "%d", &d); err != nil {
		return 0
	}

	return int(minimum(maximum(d, 1), 9))
}

// vim: set ts=4 sw=4 noet:
