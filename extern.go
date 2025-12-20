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
	"sync"
)

// ─── Build Information ──────────────────────────────────────────────

const (
	BR_VERSION = "0.75"
)

// ─── Constants ──────────────────────────────────────────────────────

const (
	MAXMARKS     = 10
	READBUFSIZ   = 1024
	SEARCH_RESET = 0
	TABWIDTH     = 4
)

// ─── Terminal Control Sequences ─────────────────────────────────────

const (
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
	XTERMTITLE   = "\033]0;%s\007"
)

// ─── Graphic Line Drawing ───────────────────────────────────────────

const (
	ENTERGRAPHICS = "\033(0"
	EXITGRAPHICS  = "\033(B"

	LEFTTEE    = "\033)0u"
	RIGHTTEE   = "\033)0t"
	HORIZLINE  = "\033)0q"
	VERTLINE   = "\033)0x"
	UPPERLEFT  = "\033)0l"
	UPPERRIGHT = "\033)0k"
	LOWERLEFT  = "\033)0m"
	LOWERRIGHT = "\033)0j"
)

// ─── Color Modes ────────────────────────────────────────────────────

const (
	_VID_BLINK = "\033[5m"
	_VID_BOLD  = "\033[1m"
	_VID_REV   = "\033[7m"
	_VID_OFF   = "\033[0m"

	_VID_BLACK_FG  = "\033[38;5;16m"
	_VID_WHITE_FG  = "\033[38;5;15m"
	_VID_GREEN_FG  = "\033[38;5;46m"
	_VID_ORANGE_FG = "\033[38;5;208m"
	_VID_YELLOW_FG = "\033[38;5;226m"

	_VID_BLACK_BG  = "\033[48;5;16m"
	_VID_GREEN_BG  = "\033[48;5;46m"
	_VID_BLUE_BG   = "\033[48;5;21m"
	_VID_ORANGE_BG = "\033[48;5;208m"
	_VID_RED_BG    = "\033[48;5;160m"
)

// ─── Meaningful Attribute Groupings ─────────────────────────────────

const (
	VIDOFF     = _VID_OFF
	VIDBLINK   = _VID_BLINK
	VIDBOLDREV = _VID_BOLD + _VID_REV
	VIDHELP    = _VID_WHITE_FG + _VID_BLUE_BG

	MSG_GREEN         = _VID_BOLD + _VID_BLACK_FG + _VID_GREEN_BG
	MSG_ORANGE        = _VID_BOLD + _VID_BLACK_FG + _VID_ORANGE_BG
	MSG_RED           = _VID_BOLD + _VID_WHITE_FG + _VID_RED_BG
	MSG_NO_COMPLETION = _VID_BOLD + _VID_ORANGE_FG + _VID_BLACK_BG
)

// ─── Scrolling Modes ────────────────────────────────────────────────

const (
	MODE_SCROLL_NONE   = 0
	MODE_SCROLL_UP     = 1
	MODE_SCROLL_DN     = 2
	MODE_SCROLL_TAIL   = 3
	MODE_SCROLL_FOLLOW = 4
)

// ─── Default and Browse History Files ───────────────────────────────

const (
	RCDIRNAME      = ".browse"
	RCFILENAME     = "browserc"
	fileHistory    = "browse_files"
	commHistory    = "browse_shell"
	searchHistory  = "browse_search"
	dirHistory     = "browse_dirs"
	maxHistorySize = 500
)

// ─── browseObj Definition ───────────────────────────────────────────

type browseObj struct {
	// Terminal configuration
	tty        *os.File
	initTitle  string
	title      string
	dispWidth  int
	dispHeight int
	dispRows   int
	firstRow   int
	lastRow    int

	// File handling and structure
	fp          *os.File
	fileName    string
	absFileName string
	fromStdin   bool
	mapSiz      int
	seekMap     map[int]int64
	sizeMap     map[int]int64
	shiftWidth  int
	lastKey     byte

	// Search and match
	pattern    string
	re         *regexp.Regexp
	replace    string
	ignoreCase bool
	lastMatch  int

	// State flags
	hitEOF   bool
	shownEOF bool
	shownMsg bool
	saveRC   bool
	exit     bool

	// File size tracking
	newFileSiz int64
	savFileSiz int64
	newInode   uint64
	savInode   uint64

	// Marks (bookmarks within the file)
	marks [MAXMARKS]int

	// Display settings
	modeNumbers bool
	modeScroll  int

	// Synchronization
	mutex sync.Mutex
}

// vim: set ts=4 sw=4 noet:
