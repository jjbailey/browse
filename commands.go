// commands.go
// the command processor
// all user activity starts here
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
	"regexp"
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
		CMD_SHIFT_LEFT      = '<'
		CMD_SHIFT_RIGHT     = '>'
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
		CMD_GREP            = '&'
		CMD_SEARCH_CLEAR    = 'C'

		VK_UP    = "\033[A\000"
		VK_DOWN  = "\033[B\000"
		VK_LEFT  = "\033[D\000"
		VK_RIGHT = "\033[C\000"

		VK_HOME  = "\033[1~"
		VK_END   = "\033[4~"
		VK_PRIOR = "\033[5~"
		VK_NEXT  = "\033[6~"
	)

	const (
		SEARCH_FWD  = true
		SEARCH_REV  = false
		SCROLL_TAIL = 256
		SCROLL_CONT = 2
	)

	var searchDir bool = SEARCH_FWD

	// seed the saved search pattern

	if len(br.pattern) > 0 {
		var err error

		br.re, err = regexp.Compile(br.pattern)

		if err != nil {
			// silently throw away bad pattern
			br.re = nil
			br.pattern = ""
		} else {
			// save regexp.Compile source and replstr
			br.pattern = br.re.String()
			br.replstr = fmt.Sprintf("%s%s%s", VIDBOLDGREEN, "$0", VIDOFF)
		}
	}

	// reasons for a delayed start

	if br.fromStdin {
		// wait a sec for more input
		time.Sleep(1 * time.Second)
	}

	if br.modeScrollDown {
		// another attempt to read more
		time.Sleep(1 * time.Second)
	}

	if br.firstRow > br.mapSiz {
		// one last attempt for big files
		time.Sleep(1 * time.Second)
	}

	ttyBrowser()
	br.pageHeader()

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

		case VK_HOME:
			// home/SOF
			b[0] = CMD_SOF

		case VK_END:
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
		}

		// commands

		switch b[0] {

		case CMD_PAGE_DN, CMD_PAGE_DN_1:
			// page forward/down
			if !br.hitEOF {
				br.pageDown()
			} else {
				br.restoreLast()
				movecursor(2, 1, false)
			}

		case CMD_SCROLL_DN, CMD_SCROLL_DN_1:
			// scroll forward/down
			br.scrollDown(1)
			movecursor(2, 1, false)

		case CMD_MODE_DN:
			// toggle continuous scroll-down mode
			if br.modeScrollDown {
				br.modeScrollDown = false
				movecursor(2, 1, false)
			} else {
				br.modeScrollDown = true
				fmt.Print(CURRESTORE)
			}
			// modeTail is a faster version of modeScrollDown
			br.modeTail = false

		case CMD_PAGE_UP:
			// page backward/up
			if br.firstRow > 0 {
				br.pageUp()
			} else {
				movecursor(2, 1, false)
			}

		case CMD_SCROLL_UP:
			// scroll backward/up
			br.scrollUp(1)
			movecursor(2, 1, false)

		case CMD_MODE_UP:
			// toggle continuous scroll-up mode
			br.modeScrollUp = !br.modeScrollUp

		case CMD_SHIFT_LEFT:
			// horizontal scroll left
			if br.shiftWidth > 0 {
				br.shiftWidth--
				br.pageCurrent()
			} else {
				movecursor(2, 1, false)
			}

		case CMD_SHIFT_RIGHT:
			// horizontal scroll right
			if br.shiftWidth < READBUFSIZ {
				br.shiftWidth++
				br.pageCurrent()
			} else {
				movecursor(2, 1, false)
			}

		case CMD_SOF, CMD_SOF_1:
			// beginning of file
			br.printPage(0)

		case CMD_EOF, CMD_EOF_1:
			// end of file
			br.pageLast()

		case CMD_NUMBERS:
			// show line numbers
			br.modeNumbers = !br.modeNumbers
			br.pageCurrent()

		case CMD_MODE_TAIL:
			// tail file
			if br.modeTail {
				br.modeTail = false
				movecursor(2, 1, false)
			} else {
				br.modeTail = true
				if !br.hitEOF {
					br.pageLast()
				}
				fmt.Print(CURRESTORE)
			}
			// modeScrollDown is a slower version of modeTail
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
			movecursor(2, 1, false)

		case CMD_SEARCH_FWD:
			// search forward/down
			patbuf, cancel := br.userInput("/")
			searchDir = SEARCH_FWD
			if cancel {
				br.restoreLast()
				movecursor(2, 1, false)
			} else if len(patbuf) == 0 {
				// null -- change direction
				br.timedMessage("Searching forward")
				// next
				br.searchFile(br.pattern, searchDir, true)
			} else {
				// search this page
				br.lastMatch = SEARCH_RESET
				br.searchFile(patbuf, searchDir, false)
			}

		case CMD_SEARCH_REV:
			// search backward/up
			patbuf, cancel := br.userInput("?")
			searchDir = SEARCH_REV
			if cancel {
				br.restoreLast()
				movecursor(2, 1, false)
			} else if len(patbuf) == 0 {
				// null -- change direction
				br.timedMessage("Searching reverse")
				// next
				br.searchFile(br.pattern, searchDir, true)
			} else {
				// search this page
				br.lastMatch = SEARCH_RESET
				br.searchFile(patbuf, searchDir, false)
			}

		case CMD_SEARCH_NEXT:
			br.searchFile(br.pattern, searchDir, true)

		case CMD_SEARCH_NEXT_REV:
			// vim compat
			br.searchFile(br.pattern, !searchDir, true)

		case CMD_GREP:
			// grep -nP pattern
			br.grep()

		case CMD_SEARCH_CLEAR:
			// clear the search pattern
			br.re = nil
			br.pattern = ""
			br.printMessage("Search cleared")

		case CMD_MARK:
			// mark page
			lbuf, cancel := br.userInput("Mark: ")
			if cancel {
				br.restoreLast()
			} else if m := getMark(lbuf); m != 0 {
				br.marks[m] = br.firstRow
				br.printMessage(fmt.Sprintf("Mark %d at line %d", m, br.marks[m]))
			}
			movecursor(2, 1, false)

		case CMD_BASH:
			cancel := br.bashCommand()
			if cancel {
				br.restoreLast()
				movecursor(2, 1, false)
			} else {
				br.pageHeader()
				br.pageCurrent()
			}

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
			movecursor(2, 1, false)
		}
	}
}

// vim: set ts=4 sw=4 noet:
