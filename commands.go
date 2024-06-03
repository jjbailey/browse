// commands.go
// the command processor
// all user activity starts here
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"time"
	"unicode"
)

func commands(br *browseObj) {
	const (
		CMD_BASH            = '!'
		CMD_EOF             = '$'
		CMD_EOF_1           = 'G'
		CMD_HELP            = 'h'
		CMD_JUMP            = 'j'
		CMD_MARK            = 'm'
		CMD_NUMBERS         = '#'
		CMD_PAGE_DN         = 'f'
		CMD_PAGE_DN_1       = ' '
		CMD_PAGE_UP         = 'b'
		CMD_HALF_PAGE_DN    = '\006'
		CMD_HALF_PAGE_DN_1  = '\004'
		CMD_HALF_PAGE_DN_2  = 'z'
		CMD_HALF_PAGE_UP    = '\002'
		CMD_HALF_PAGE_UP_1  = '\025'
		CMD_HALF_PAGE_UP_2  = 'Z'
		CMD_SHIFT_LEFT      = '<'
		CMD_SHIFT_LEFT_1    = '\010'
		CMD_SHIFT_RIGHT     = '>'
		CMD_SHIFT_RIGHT_1   = '\011'
		CMD_QUIT            = 'q'
		CMD_QUIT_NO_SAVE    = 'Q'
		CMD_SCROLL_DN       = '+'
		CMD_SCROLL_DN_1     = '\r'
		CMD_SCROLL_UP       = '-'
		CMD_MODE_DN         = 'd'
		CMD_MODE_UP         = 'u'
		CMD_MODE_TAIL       = 't'
		CMD_MODE_FOLLOW     = 'e'
		CMD_SOF             = '^'
		CMD_SOF_1           = '0'
		CMD_SEARCH_FWD      = '/'
		CMD_SEARCH_REV      = '?'
		CMD_SEARCH_NEXT     = 'n'
		CMD_SEARCH_NEXT_REV = 'N'
		CMD_SEARCH_IGN_CASE = 'i'
		CMD_GREP            = '&'
		CMD_PERCENT         = '%'
		CMD_SEARCH_CLEAR    = 'C'

		VK_UP    = "\033[A\000"
		VK_DOWN  = "\033[B\000"
		VK_LEFT  = "\033[D\000"
		VK_RIGHT = "\033[C\000"

		VK_HOME   = "\033[1~"
		VK_HOME_1 = "\033[H\000"
		VK_END    = "\033[4~"
		VK_END_1  = "\033[F\000"
		VK_PRIOR  = "\033[5~"
		VK_NEXT   = "\033[6~"
	)

	const (
		SEARCH_FWD  = true
		SEARCH_REV  = false
		SCROLL_TAIL = 256
		SCROLL_CONT = 2
	)

	var searchDir bool = SEARCH_FWD

	// seed the saved search pattern
	br.reCompile(br.pattern)

	// reasons for a delayed start

	if br.fromStdin {
		// wait for some input
		time.Sleep(500 * time.Millisecond)
	}

	if br.inMotion() {
		// another attempt to read more
		time.Sleep(500 * time.Millisecond)
	}

	if br.firstRow > br.mapSiz {
		// one last attempt for big files
		time.Sleep(500 * time.Millisecond)
	}

	ttyBrowser()
	br.pageHeader()

	// gratuitous save cursor
	moveCursor(2, 1, false)
	fmt.Print(CURSAVE)

	if br.inMotion() {
		br.pageLast()
		fmt.Print(CURRESTORE)
	} else {
		br.pageCurrent()
	}

	for {
		// scan for input -- compare 4 characters

		b := make([]byte, 4)
		_, err := br.tty.Read(b)

		// continuous modes

		if err != nil {
			switch br.modeScroll {

			case MODE_SCROLL_UP:
				// in continuous scroll-up mode
				br.scrollUp(SCROLL_CONT)

			case MODE_SCROLL_DN:
				// in continuous scroll-down mode
				br.scrollDown(SCROLL_CONT)

			case MODE_SCROLL_TAIL:
				// in tail mode
				br.scrollDown(SCROLL_TAIL)

			case MODE_SCROLL_FOLLOW:
				// in follow mode
				br.scrollDown(SCROLL_CONT)
			}

			continue
		}

		// convert arrow and page keys to commands

		switch string(b) {

		case VK_UP:
			// up arrow -- lines move down
			b[0] = CMD_MODE_UP

		case VK_DOWN:
			// down arrow -- lines move up
			b[0] = CMD_MODE_DN

		case VK_RIGHT:
			// right arrow -- scroll up one
			b[0] = CMD_SCROLL_DN

		case VK_LEFT:
			// left arrow -- scroll down one
			b[0] = CMD_SCROLL_UP

		case VK_HOME, VK_HOME_1:
			// home/SOF
			b[0] = CMD_SOF

		case VK_END, VK_END_1:
			// end/EOF
			b[0] = CMD_MODE_FOLLOW

		case VK_PRIOR:
			// PG UP
			b[0] = CMD_PAGE_UP

		case VK_NEXT:
			// PG DN
			b[0] = CMD_PAGE_DN
		}

		// mode cancellations

		prevMotion := br.inMotion()

		if string(b) != "" {
			switch b[0] {

			case CMD_MODE_UP:
				// toggle scroll up mode
				br.toggleScroll(MODE_SCROLL_UP)

			case CMD_MODE_DN:
				// toggle scroll down mode
				br.toggleScroll(MODE_SCROLL_DN)

			case CMD_MODE_TAIL:
				// toggle tail mode
				br.toggleScroll(MODE_SCROLL_TAIL)

			case CMD_MODE_FOLLOW:
				// toggle follow mode
				br.toggleScroll(MODE_SCROLL_FOLLOW)

			default:
				br.modeScroll = MODE_SCROLL_NONE

				if prevMotion && b[0] == CMD_PAGE_DN_1 {
					// CMD_PAGE_DN_1 doubles as mode cancel
					moveCursor(2, 1, false)
					continue
				}
			}
		}

		// commands

		switch b[0] {

		case CMD_PAGE_DN, CMD_PAGE_DN_1:
			// page forward/down
			br.restoreLast()
			br.pageDown()

		case CMD_SCROLL_DN, CMD_SCROLL_DN_1:
			// scroll forward/down
			br.scrollDown(1)
			moveCursor(2, 1, false)

		case CMD_MODE_DN:
			// follow mode -- follow file leisurely
			if br.inMotion() {
				fmt.Print(CURRESTORE)
			} else {
				moveCursor(2, 1, false)
			}

		case CMD_MODE_UP:
			// continuous scroll-up mode
			if br.inMotion() {
				fmt.Print(CURRESTORE)
			} else {
				moveCursor(2, 1, false)
			}

		case CMD_PAGE_UP:
			// page backward/up
			if br.firstRow > 0 {
				br.pageUp()
			} else {
				moveCursor(2, 1, false)
			}

		case CMD_SCROLL_UP:
			// scroll backward/up
			br.scrollUp(1)
			moveCursor(2, 1, false)

		case CMD_SHIFT_LEFT, CMD_SHIFT_LEFT_1:
			// horizontal scroll left
			if br.shiftWidth > 0 {
				br.shiftWidth -= TABWIDTH
				br.pageCurrent()
			}

		case CMD_SHIFT_RIGHT, CMD_SHIFT_RIGHT_1:
			// horizontal scroll right
			if br.shiftWidth < (READBUFSIZ - (TABWIDTH * 2)) {
				br.shiftWidth += TABWIDTH
				br.pageCurrent()
			}

		case CMD_SOF, CMD_SOF_1:
			// beginning of file at column 1
			br.shiftWidth = 0
			br.printPage(0)

		case CMD_EOF, CMD_EOF_1:
			// end of file
			br.pageLast()

		case CMD_NUMBERS:
			// show line numbers
			br.modeNumbers = !br.modeNumbers
			br.pageCurrent()

		case CMD_MODE_TAIL:
			// tail mode -- follow file rapidly
			br.pageLast()
			if br.inMotion() {
				fmt.Print(CURRESTORE)
			}

		case CMD_MODE_FOLLOW:
			// follow mode -- follow file leisurely
			br.pageLast()
			if br.inMotion() {
				fmt.Print(CURRESTORE)
			}

		case CMD_JUMP:
			// jump to line
			lbuf, cancel := br.userInput("Junp: ")
			if cancel || len(lbuf) == 0 {
				br.restoreLast()
			} else {
				var n int
				fmt.Sscanf(lbuf, "%d", &n)
				br.printPage(n)
			}

		case CMD_SEARCH_FWD:
			// search forward/down
			searchDir = br.doSearch(searchDir, SEARCH_FWD)

		case CMD_SEARCH_REV:
			// search backward/up
			searchDir = br.doSearch(searchDir, SEARCH_REV)

		case CMD_SEARCH_NEXT:
			br.searchFile(br.pattern, searchDir, true)

		case CMD_SEARCH_NEXT_REV:
			// vim compat
			br.searchFile(br.pattern, !searchDir, true)

		case CMD_SEARCH_IGN_CASE:
			br.ignoreCase = !br.ignoreCase
			br.reCompile(br.pattern)
			if br.ignoreCase {
				br.printMessage("Search ignores case")
			} else {
				br.printMessage("Search considers case")
			}

		case CMD_GREP:
			// grep -nP pattern
			br.grep()

		case CMD_SEARCH_CLEAR:
			// clear the search pattern
			br.re = nil
			br.pattern = ""
			br.printMessage("Search pattern cleared")

		case CMD_MARK:
			// mark page
			lbuf, cancel := br.userInput("Mark: ")
			if cancel {
				br.restoreLast()
			} else if m := getMark(lbuf); m != 0 {
				br.marks[m] = br.firstRow
				br.printMessage(fmt.Sprintf("Mark %d at line %d", m, br.marks[m]))
			}

		case CMD_BASH:
			br.bashCommand()

		case CMD_HALF_PAGE_DN, CMD_HALF_PAGE_DN_1, CMD_HALF_PAGE_DN_2:
			// scroll half page forward/down
			br.scrollDown(br.dispRows >> 1)

		case CMD_HALF_PAGE_UP, CMD_HALF_PAGE_UP_1, CMD_HALF_PAGE_UP_2:
			// scroll half page backward/up
			br.scrollUp(br.dispRows >> 1)

		case CMD_PERCENT:
			// this page position in percentages
			// +1 for EOF
			t := float32(br.firstRow) / float32(br.mapSiz+1) * 100.0
			b := float32(br.lastRow) / float32(br.mapSiz+1) * 100.0
			br.printMessage(fmt.Sprintf("Position is %1.2f%% - %1.2f%%", t, b))

		case CMD_QUIT:
			// quit -- this is the only way to save an rc file
			br.saveRC = true
			return

		case CMD_QUIT_NO_SAVE:
			// Quit -- do not save an rc file
			br.saveRC = false
			return

		case CMD_HELP:
			// help
			br.printHelp()

		default:
			// if digit, go to marked page
			if unicode.IsDigit(rune(b[0])) {
				br.pageMarked(getMark(string(b)))
			}
			// no modes active
			moveCursor(2, 1, false)
		}
	}
}

// vim: set ts=4 sw=4 noet:
