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

	if br.modeScrollDown {
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

	if br.modeScrollDown {
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
			if br.modeTail {
				// in tail mode
				br.scrollDown(SCROLL_TAIL)
			}

			if br.modeScrollUp {
				// in continuous scroll-up mode
				br.scrollUp(SCROLL_CONT)
			}

			if br.modeScrollDown {
				// in continuous scroll-down mode
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
			b[0] = CMD_EOF

		case VK_PRIOR:
			// PG UP
			b[0] = CMD_PAGE_UP

		case VK_NEXT:
			// PG DN
			b[0] = CMD_PAGE_DN
		}

		// mode cancellations

		inMotion := (br.modeTail || br.modeScrollUp || br.modeScrollDown)

		if string(b) != "" {
			if b[0] != CMD_MODE_TAIL {
				br.modeTail = false
			}

			if b[0] != CMD_MODE_UP {
				br.modeScrollUp = false
			}

			if b[0] != CMD_MODE_DN {
				br.modeScrollDown = false
			}

			if inMotion && b[0] == CMD_PAGE_DN_1 {
				// CMD_PAGE_DN_1 doubles as mode cancel
				moveCursor(2, 1, false)
				continue
			}
		}

		// commands

		switch b[0] {

		case CMD_PAGE_DN, CMD_PAGE_DN_1:
			// page forward/down
			if !br.hitEOF {
				br.pageDown()
			} else {
				br.restoreLast()
				moveCursor(2, 1, false)
			}

		case CMD_SCROLL_DN, CMD_SCROLL_DN_1:
			// scroll forward/down
			br.scrollDown(1)
			moveCursor(2, 1, false)

		case CMD_MODE_DN:
			// follow mode -- follow file leisurely
			// toggle in either follow case
			br.modeScrollDown = !br.modeScrollDown
			if inMotion {
				moveCursor(2, 1, false)
			} else {
				fmt.Print(CURRESTORE)
			}
			br.modeTail = false

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

		case CMD_MODE_UP:
			// toggle continuous scroll-up mode
			br.modeScrollUp = !br.modeScrollUp

		case CMD_SHIFT_LEFT, CMD_SHIFT_LEFT_1:
			// horizontal scroll left
			if br.shiftWidth > 0 {
				br.shiftWidth -= TABWIDTH
				br.pageCurrent()
			} else {
				moveCursor(2, 1, false)
			}

		case CMD_SHIFT_RIGHT, CMD_SHIFT_RIGHT_1:
			// horizontal scroll right
			if br.shiftWidth < (READBUFSIZ - (TABWIDTH * 2)) {
				br.shiftWidth += TABWIDTH
				br.pageCurrent()
			} else {
				moveCursor(2, 1, false)
			}

		case CMD_SOF, CMD_SOF_1:
			// beginning of file at column 1
			br.shiftWidth = 0
			br.printPage(0)

		case CMD_EOF, CMD_EOF_1:
			// follow mode -- follow file leisurely
			if inMotion {
				br.modeScrollDown = false
				moveCursor(2, 1, false)
			} else {
				br.modeScrollDown = true
				if !br.hitEOF {
					br.pageLast()
				}
				fmt.Print(CURRESTORE)
			}
			br.modeTail = false

		case CMD_NUMBERS:
			// show line numbers
			br.modeNumbers = !br.modeNumbers
			br.pageCurrent()

		case CMD_MODE_TAIL:
			// tail mode -- follow file rapidly
			if inMotion {
				br.modeTail = false
				moveCursor(2, 1, false)
			} else {
				br.modeTail = true
				if !br.hitEOF {
					br.pageLast()
				}
				fmt.Print(CURRESTORE)
			}
			br.modeScrollDown = false

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
			moveCursor(2, 1, false)

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
			moveCursor(2, 1, false)

		case CMD_BASH:
			cancel := br.bashCommand()
			if cancel {
				br.restoreLast()
				moveCursor(2, 1, false)
			} else {
				br.pageHeader()
				br.pageCurrent()
			}

		case CMD_HALF_PAGE_DN, CMD_HALF_PAGE_DN_1, CMD_HALF_PAGE_DN_2:
			// half page forward/down
			br.printPage(br.firstRow + (br.dispRows >> 1))

		case CMD_HALF_PAGE_UP, CMD_HALF_PAGE_UP_1, CMD_HALF_PAGE_UP_2:
			// half page backward/up
			br.printPage(br.firstRow - (br.dispRows >> 1))

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
