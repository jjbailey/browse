// help.go
// the help screen
// ignore SIGWINCH here
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"strings"
)

func (br *browseObj) printHelp() {
	const (
		paddingTop  = 3
		paddingSide = 2
	)

	lines := []string{
		"                                                                     ",
		"  Command                       Function                             ",
		"  f [PAGE DOWN]  b [PAGE UP]    Page down/up                         ",
		"  ^F ^D z  ^B ^U Z              Scroll half page down/up             ",
		"  + [RIGHT] [ENTER]  - [LEFT]   Scroll one line down/up              ",
		"  d [DOWN]  u [UP]              Continuous scroll mode               ",
		"  > [TAB]  < [BACKSPACE] [DEL]  Scroll 4 characters right/left       ",
		"  ^ $                           Scroll to column 1, scroll to EOL    ",
		"  #                             Line numbers                         ",
		"  % ^G                          Page position                        ",
		"  j                             Jump to line number                  ",
		"  0 [HOME]                      Jump to SOF, column 1                ",
		"  G                             Jump to EOF                          ",
		"  m                             Mark a page with number 1-9          ",
		"  1-9                           Jump to mark                         ",
		"  / ?                           Regex search forward/reverse         ",
		"  n N                           Repeat search forward/reverse        ",
		"  i                             Case-sensitive search                ",
		"  F                             Run 'fmt -s' on the current file     ",
		"  &                             Run 'grep -nP' for pattern           ",
		"  p P                           Print/Clear search pattern           ",
		"  e [END]                       Follow mode                          ",
		"  t                             Tail mode                            ",
		"  !                             bash command                         ",
		"  B                             Browse file (expands %, ~, glob)     ",
		"  a                             Print filenames in the browse list   ",
		"  c C                           Print/Change working directory       ",
		"  q Q                           Quit, save/don't save browserc       ",
		"  x X                           Exit list, save/don't save browserc  ",
		"                                                                     ",
		"  Press any key to continue browsing...                              ",
		"                                                                     ",
	}

	helpHeight := len(lines)
	helpWidth := len(lines[0])

	// Verify screen space
	if br.dispHeight < (helpHeight+paddingTop+1) || br.dispWidth < (helpWidth+paddingSide) {
		br.timedMessage("Screen is too small... checking man page", MSG_ORANGE)
		br.manPage()
		return
	}

	// Centered horizontal offset
	col := (br.dispWidth - helpWidth) / 2
	row := paddingTop

	fmt.Print(VIDHELP)

	// Top border
	moveCursor(row, col, false)
	fmt.Printf("%s%s%s%s%s",
		ENTERGRAPHICS, UPPERLEFT,
		strings.Repeat(HORIZLINE, helpWidth),
		UPPERRIGHT, EXITGRAPHICS)

	// Title
	moveCursor(row, col+3, false)
	fmt.Print(" Browse ")
	moveCursor(row, col+12, false)
	fmt.Printf(" v%s ", BR_VERSION)

	// Main body
	for i, line := range lines {
		moveCursor(row+1+i, col, false)
		fmt.Printf("%s%s%s%s%s%s%s",
			ENTERGRAPHICS, VERTLINE, EXITGRAPHICS,
			line,
			ENTERGRAPHICS, VERTLINE, EXITGRAPHICS)
	}

	// Bottom border
	moveCursor(row+1+helpHeight, col, false)
	fmt.Printf("%s%s%s%s%s",
		ENTERGRAPHICS, LOWERLEFT,
		strings.Repeat(HORIZLINE, helpWidth),
		LOWERRIGHT, EXITGRAPHICS)

	fmt.Print(VIDOFF)

	br.userAnyKey("")
	br.resizeWindow()
	br.catchSignals()
}

// vim: set ts=4 sw=4 noet:
