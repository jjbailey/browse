// help.go
// the help screen
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
	"strings"
)

func (x *browseObj) printHelp() {
	var i int

	lines := []string{
		"                                                               ",
		"   Browse                       Version 0.14                   ",
		"                                                               ",
		"   Command                      Function                       ",
		"   f b [PAGE UP] [PAGE DOWN]    Page forward/back (down/up)    ",
		"   + - [LEFT] [RIGHT] [ENTER]   Scroll one line                ",
		"   u d [UP] [DOWN]              Continuous scroll mode         ",
		"   < >                          Horizontal scroll left/right   ",
		"   N #                          Line numbers                   ",
		"   j                            Jump to line number            ",
		"   0 ^ [HOME]                   Jump to SOF                    ",
		"   G $ [END]                    Jump to EOF                    ",
		"   m                            Mark a page with number 1-9    ",
		"   1-9                          Jump to mark                   ",
		"   /                            Regex pattern search           ",
		"   ?                            Regex pattern reverse search   ",
		"   n                            Repeat search                  ",
		"   C                            Clear search                   ",
		"   t                            Tail mode                      ",
		"   !                            bash command                   ",
		"   q                            Quit                           ",
		"   Q                            Quit, don't save ~/.browserc   ",
		"                                                               ",
		"   Press any key to continue browsing...                       ",
		"                                                               ",
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
	fmt.Printf("%s", WHITEBLUE)
	movecursor(3, col, false)
	fmt.Printf(ENTERGRAPHICS + UPPERLEFT)
	fmt.Printf("%s", strings.Repeat(HORIZLINE, len(lines[0])))
	fmt.Printf(UPPERRIGHT + EXITGRAPHICS)

	// body

	for i = 0; i < len(lines); i++ {
		movecursor(i+4, col, false)

		fmt.Printf("%s%s%s%s%s%s%s",
			ENTERGRAPHICS, VERTLINE, EXITGRAPHICS,
			lines[i],
			ENTERGRAPHICS, VERTLINE, EXITGRAPHICS)
	}

	// bottom line

	movecursor(i+4, col, false)
	fmt.Printf(ENTERGRAPHICS + LOWERLEFT)
	fmt.Printf("%s", strings.Repeat(HORIZLINE, len(lines[0])))
	fmt.Printf(LOWERRIGHT + EXITGRAPHICS)
	fmt.Printf("%s", VIDOFF)

	// prompt is in the body of the help screen
	x.userAnyKey("")
	x.pageCurrent()
}

// vim: set ts=4 sw=4 noet:
