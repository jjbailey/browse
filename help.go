// help.go
// the help screen
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"strings"
)

func (x *browseObj) printHelp() {
	var i int

	lines := []string{
		"                                                                ",
		"   Browse                       Version " + BR_VERSION + "                    ",
		"                                                                ",
		"   Command                      Function                        ",
		"   f b [PAGE UP] [PAGE DOWN]    Page down/up                    ",
		"   + - [LEFT] [RIGHT] [ENTER]   Scroll one line                 ",
		"   u d [UP] [DOWN]              Continuous scroll mode          ",
		"   < >                          Horizontal scroll left/right    ",
		"   #                            Line numbers                    ",
		"   j                            Jump to line number             ",
		"   0 ^ [HOME]                   Jump to SOF                     ",
		"   G $ [END]                    Jump to EOF                     ",
		"   z                            Center page on top line         ",
		"   m                            Mark a page with number 1-9     ",
		"   1-9                          Jump to mark                    ",
		"   / ?                          Regex search forward/reverse    ",
		"   n N                          Repeat search forward/reverse   ",
		"   i                            Case-sensitive search           ",
		"   &                            Pipe search to grep -nP         ",
		"   C                            Clear search                    ",
		"   t                            Tail mode                       ",
		"   !                            bash command                    ",
		"   q                            Quit                            ",
		"   Q                            Quit, don't save ~/.browserc    ",
		"                                                                ",
		"   Press any key to continue browsing...                        ",
		"                                                                ",
	}

	// the help screen needs room to print

	helpHeight := len(lines)
	helpWidth := len(lines[0])

	if x.dispHeight < (helpHeight+4) || x.dispWidth < (helpWidth+2) {
		x.printMessage("Screen is too small")
		return
	}

	// top line

	col := int((x.dispWidth - helpWidth) / 2)
	fmt.Print(VIDHELP)
	moveCursor(3, col, false)
	fmt.Printf(ENTERGRAPHICS + UPPERLEFT)
	fmt.Print(strings.Repeat(HORIZLINE, len(lines[0])))
	fmt.Printf(UPPERRIGHT + EXITGRAPHICS)

	// body

	for i = 0; i < len(lines); i++ {
		moveCursor(i+4, col, false)

		fmt.Printf("%s%s%s%s%s%s%s",
			ENTERGRAPHICS, VERTLINE, EXITGRAPHICS,
			lines[i],
			ENTERGRAPHICS, VERTLINE, EXITGRAPHICS)
	}

	// bottom line

	moveCursor(i+4, col, false)
	fmt.Printf(ENTERGRAPHICS + LOWERLEFT)
	fmt.Print(strings.Repeat(HORIZLINE, len(lines[0])))
	fmt.Printf(LOWERRIGHT + EXITGRAPHICS)
	fmt.Print(VIDOFF)

	// prompt is in the body of the help screen
	x.userAnyKey("")
	x.pageCurrent()
}

// vim: set ts=4 sw=4 noet:
