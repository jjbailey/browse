// help.go
// the help screen
// ignore SIGWINCH here
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"os/signal"
	"strings"
	"syscall"
)

func (x *browseObj) printHelp() {
	var i int

	lines := []string{
		"                                                                  ",
		"   Command                        Function                        ",
		"   f [PAGE DOWN]  b [PAGE UP]     Page down/up                    ",
		"   ^F ^D z  ^B ^U Z               Scroll half page down/up        ",
		"   + [RIGHT] [ENTER]  - [LEFT]    Scroll one line down/up         ",
		"   d [DOWN]  u [UP]               Continuous scroll mode          ",
		"   > [TAB]  < [BACKSPACE] [DEL]   Horizontal scroll right/left    ",
		"   #                              Line numbers                    ",
		"   % ^G                           Page position                   ",
		"   j                              Jump to line number             ",
		"   0 ^ [HOME]  G $                Jump to SOF/EOF                 ",
		"   m                              Mark a page with number 1-9     ",
		"   1-9                            Jump to mark                    ",
		"   / ?                            Regex search forward/reverse    ",
		"   n N                            Repeat search forward/reverse   ",
		"   i                              Case-sensitive search           ",
		"   &                              Run 'grep -nP' for pattern      ",
		"   C                              Clear search pattern            ",
		"   e [END]                        Follow mode                     ",
		"   t                              Tail mode                       ",
		"   !                              bash command                    ",
		"   q                              Quit                            ",
		"   Q                              Quit, don't save ~/.browserc    ",
		"                                                                  ",
		"   Press any key to continue browsing...                          ",
		"                                                                  ",
	}

	// the help screen needs room to print

	helpHeight := len(lines)
	helpWidth := len(lines[0])

	if x.dispHeight < (helpHeight+4) || x.dispWidth < (helpWidth+2) {
		x.printMessage("Screen is too small", MSG_ORANGE)
		return
	}

	// top line
	col := int((x.dispWidth - helpWidth) / 2)
	fmt.Print(VIDHELP)
	moveCursor(3, col, false)
	fmt.Printf(ENTERGRAPHICS + UPPERLEFT)
	fmt.Print(strings.Repeat(HORIZLINE, helpWidth))
	fmt.Printf(UPPERRIGHT + EXITGRAPHICS)

	// title
	moveCursor(3, col+3, false)
	fmt.Printf(" Browse ")
	moveCursor(3, col+12, false)
	fmt.Printf(" v" + BR_VERSION + " ")

	// body
	for i = 0; i < helpHeight; i++ {
		moveCursor(i+4, col, false)

		fmt.Printf("%s%s%s%s%s%s%s",
			ENTERGRAPHICS, VERTLINE, EXITGRAPHICS,
			lines[i],
			ENTERGRAPHICS, VERTLINE, EXITGRAPHICS)
	}

	// bottom line
	moveCursor(i+4, col, false)
	fmt.Printf(ENTERGRAPHICS + LOWERLEFT)
	fmt.Print(strings.Repeat(HORIZLINE, helpWidth))
	fmt.Printf(LOWERRIGHT + EXITGRAPHICS)
	fmt.Print(VIDOFF)

	// signals
	signal.Ignore(syscall.SIGWINCH)

	// prompt is in the body of the help screen
	x.userAnyKey("")

	// restore and reset window size
	x.resizeWindow()

	// reset signals
	x.catchSignals()
}

// vim: set ts=4 sw=4 noet:
