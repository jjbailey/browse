// util.go
// various uncategorized functions
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
	"math"
	"os"
)

func expandTabs(data []byte) ([]byte, int) {
	// replace tabs with the appropriate amount of spaces
	// assume the standard 8-character tab stops

	var newdata = make([]byte, READBUFSIZ*2)
	var j int = 0

	for i := 0; i < len(data); i++ {
		switch data[i] {

		case '\r':
			// silently map CR to space
			newdata[j] = ' '

		case '\t':
			k := TABWIDTH - (j % TABWIDTH)

			if k < 0 {
				k += TABWIDTH
			}

			for k > 0 {
				newdata[j] = ' '
				j++
				k--
			}

			continue

		default:
			newdata[j] = data[i]
		}

		if j++; j == READBUFSIZ*2 {
			break
		}
	}

	return newdata, j
}

func movecursor(row int, col int, clrflag bool) {
	fmt.Printf(CURPOS, row, col)

	if clrflag {
		fmt.Printf("%s", CLEARLINE)
	}
}

func printSEOF(what string) {
	if what == "EOF" {
		// save for modeTail
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
	fmt.Printf("%s", RESETREGION)
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
