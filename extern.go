// extern.go
// verious constants, the structure of a browser object
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"os"
	"regexp"
)

const (
	READBUFSIZ   = 512
	TABWIDTH     = 8
	MAXMARKS     = 10
	SEARCH_RESET = 0
	PAGE_SEARCH  = false
)

const (
	// xterm escape sequences
	CURPOS        = "\033[%d;%dH"
	CURUP         = "\033[A"
	CURSAVE       = "\033\067"
	CURRESTORE    = "\033\070"
	CLEARSCREEN   = "\033[0J"
	CLEARLINE     = "\033[0K"
	SCROLLREGION  = "\033[%d;%dr"
	SCROLLREV     = "\033[1L"
	RESETREGION   = "\033[r"
	VIDBLINK      = "\033[5m"
	VIDBOLDREV    = "\033[1m\033[7m"
	VIDBOLDGREEN  = "\033[1m\033[32m"
	VIDMESSAGE    = "\033[1;7m\033[32m"
	VIDOFF        = "\033[0m"
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
	WHITEBLUE     = "\033[48;5;21m"
	LINEWRAPOFF   = "\033[?7l"
	LINEWRAPON    = "\033[?7h"
)

type browseObj struct {
	// screen vars
	tty        *os.File
	title      string
	dispWidth  int
	dispHeight int
	dispRows   int
	firstRow   int
	lastRow    int

	// file vars
	fp         *os.File
	fileName   string
	fromStdin  bool
	mapSiz     int
	seekMap    map[int]int64
	sizeMap    map[int]int64
	shiftWidth int64
	pattern    string
	re         *regexp.Regexp
	replstr    string
	lastMatch  int
	hitEOF     bool
	shownEOF   bool
	shownMsg   bool
	marks      [MAXMARKS]int
	saveRC     bool

	// modes
	modeNumbers    bool
	modeScrollUp   bool
	modeScrollDown bool
	modeTail       bool
}

// vim: set ts=4 sw=4 noet:
