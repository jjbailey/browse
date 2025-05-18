// extern.go
// various constants, the structure of a browser object
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"os"
	"regexp"
)

const (
	BR_VERSION   = "0.53"
	READBUFSIZ   = 1024
	TABWIDTH     = 4
	MAXMARKS     = 10
	SEARCH_RESET = 0
)

const (
	// xterm escape sequences
	CURPOS       = "\033[%d;%dH"
	CURUP        = "\033[A"
	CURSAVE      = "\033\067"
	CURRESTORE   = "\033\070"
	CLEARSCREEN  = "\033[0J"
	CLEARLINE    = "\033[0K"
	SCROLLREGION = "\033[%d;%dr"
	SCROLLREV    = "\033[1L"
	RESETREGION  = "\033[r"
	LINEWRAPOFF  = "\033[?7l"
	LINEWRAPON   = "\033[?7h"

	ENTERGRAPHICS = "\033(0"
	EXITGRAPHICS  = "\033(B"
	LEFTTEE       = "\033)0u"
	RIGHTTEE      = "\033)0t"
	HORIZLINE     = "\033)0q"
	VERTLINE      = "\033)0x"
	LOWERRIGHT    = "\033)0j"
	UPPERRIGHT    = "\033)0k"
	UPPERLEFT     = "\033)0l"
	LOWERLEFT     = "\033)0m"
)

const (
	// colors
	_VID_BLINK = "\033[5m"
	_VID_BOLD  = "\033[1m"
	_VID_REV   = "\033[7m"
	_VID_OFF   = "\033[0m"

	_VID_BLACK_FG  = "\033[38;5;16m"
	_VID_BLUE_BG   = "\033[48;5;21m"
	_VID_GREEN_BG  = "\033[48;5;46m"
	_VID_GREEN_FG  = "\033[38;5;46m"
	_VID_ORANGE_BG = "\033[48;5;208m"
	_VID_RED_BG    = "\033[48;5;160m"
	_VID_WHITE_FG  = "\033[38;5;15m"

	VIDBLINK   = _VID_BLINK
	VIDBOLDREV = _VID_BOLD + _VID_REV
	VIDHELP    = _VID_WHITE_FG + _VID_BLUE_BG
	VIDOFF     = _VID_OFF

	MSG_GREEN  = _VID_BOLD + _VID_BLACK_FG + _VID_GREEN_BG
	MSG_ORANGE = _VID_BOLD + _VID_BLACK_FG + _VID_ORANGE_BG
	MSG_RED    = _VID_BOLD + _VID_WHITE_FG + _VID_RED_BG
)

const (
	MODE_SCROLL_NONE   = 0
	MODE_SCROLL_UP     = 1
	MODE_SCROLL_DN     = 2
	MODE_SCROLL_TAIL   = 3
	MODE_SCROLL_FOLLOW = 4
)

type browseObj struct {
	// Terminal display configuration
	tty        *os.File
	title      string
	dispWidth  int
	dispHeight int
	dispRows   int
	firstRow   int
	lastRow    int

	// File handling
	fp         *os.File
	fileName   string
	fromStdin  bool
	mapSiz     int
	seekMap    map[int]int64
	sizeMap    map[int]int64
	shiftWidth int

	// Search configuration
	pattern    string
	re         *regexp.Regexp
	replace    string
	ignoreCase bool
	lastMatch  int

	// State tracking
	hitEOF   bool
	shownEOF bool
	shownMsg bool
	marks    [MAXMARKS]int
	saveRC   bool
	exit     bool

	// File size tracking
	newFileSiz int64
	savFileSiz int64

	// Display modes
	modeNumbers bool
	modeScroll  int
}

// vim: set ts=4 sw=4 noet:
