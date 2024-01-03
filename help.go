// help.go
// the help screen
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
)

func (x *browseObj) printHelp() {
	lines := []string{
		"                                                               ",
		"   Browse                       Version 0.5                    ",
		"                                                               ",
		"   Command                      Function                       ",
		"   f b [PG UP] [PG DN]          Page forward/back (down/up)    ",
		"   + - [LEFT] [RIGHT] [ENTER]   Scroll one line                ",
		"   u d [UP] [DOWN]              Scroll continuous              ",
		"   N #                          Line numbers                   ",
		"   j                            Jump to line number            ",
		"   0 ^ [HOME]                   Jump to SOF                    ",
		"   G $ [END]                    Jump to EOF                    ",
		"   m                            Mark a page with number 1-9    ",
		"   /                            Regex pattern search           ",
		"   ?                            Regex pattern reverse search   ",
		"   n                            Repeat search                  ",
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

	col := int((x.dispWidth - helpWidth) / 2)

	// top line

	fmt.Printf("%s", WHITEBLUE)
	movecursor(3, col, false)
	fmt.Printf(ENTERGRAPHICS + UPPERLEFT)
	for j := 0; j < len(lines[0]); j++ {
		fmt.Printf("%s", HORIZLINE)
	}
	fmt.Printf(UPPERRIGHT + EXITGRAPHICS)

	// body

	i := 0

	for range lines {
		movecursor(i+4, col, false)
		fmt.Printf(ENTERGRAPHICS + VERTLINE + EXITGRAPHICS)
		fmt.Printf(lines[i])
		fmt.Printf(ENTERGRAPHICS + VERTLINE + EXITGRAPHICS)
		i++
	}

	// bottom line

	movecursor(i+4, col, false)
	fmt.Printf(ENTERGRAPHICS + LOWERLEFT)
	for j := 0; j < len(lines[0]); j++ {
		fmt.Printf("%s", HORIZLINE)
	}
	fmt.Printf(LOWERRIGHT + EXITGRAPHICS)
	fmt.Printf("%s", VIDOFF)

	// prompt is in the body of the help screen
	x.userAnyKey("")
	x.pageCurrent()
}

// vim: set ts=4 sw=4 noet:
