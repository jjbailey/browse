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
)

func commands(br *browseObj) {
	const (
		CMD_BASH         = '!'
		CMD_EOF          = '$'
		CMD_EOF_1        = 'G'
		CMD_HELP         = 'h'
		CMD_JUMP         = 'j'
		CMD_MARK         = 'm'
		CMD_NUMBERS      = '#'
		CMD_NUMBERS_1    = 'N'
		CMD_PAGE_DN      = 'f'
		CMD_PAGE_UP      = 'b'
		CMD_QUIT         = 'q'
		CMD_QUIT_NO_SAVE = 'Q'
		CMD_SCROLL_DN    = '+'
		CMD_SCROLL_DN_1  = '\r'
		CMD_SCROLL_UP    = '-'
		CMD_SOF          = '^'
		CMD_SOF_1        = '0'
		CMD_SRCH_FWD     = '/'
		CMD_SRCH_REV     = '?'
		CMD_SRCH_NEXT    = 'n'
		CMD_SRCH_CLEAR   = 'C'

		MODE_UP   = 'u'
		MODE_DN   = 'd'
		MODE_TAIL = 't'

		VK_UP    = "\033[A\000"
		VK_DOWN  = "\033[B\000"
		VK_LEFT  = "\033[D\000"
		VK_RIGHT = "\033[C\000"

		VK_HOME  = "\033[1~"
		VK_END   = "\033[4~"
		VK_PRIOR = "\033[5~"
		VK_NEXT  = "\033[6~"
	)

	var patbuf string
	var searchDir bool = SEARCHFWD

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

	ttyBrowser()
	br.pageHeader()
	br.pageCurrent()

	for {
		// scan for input -- need to compare 4 characters

		b := make([]byte, 4)
		_, err := br.tty.Read(b)
		bbuf := string(b)

		// continuous modes

		if err != nil {
			if br.modeTail {
				// in tail mode
				br.scrollDown(READBUFSIZ)
			}

			if br.modeScrollUp {
				// in continuous scroll-up mode
				br.scrollUp(2)
			}

			if br.modeScrollDown {
				// in continuous scroll-down mode
				br.scrollDown(2)
			}

			continue
		}

		// convert arrow and page keys to commands

		switch {

		case bbuf == VK_UP:
			// up arrow -- lines move down
			b[0] = MODE_UP

		case bbuf == VK_DOWN:
			// down arrow -- lines move up
			b[0] = MODE_DN

		case bbuf == VK_RIGHT:
			// right arrow -- scroll up one
			b[0] = CMD_SCROLL_DN

		case bbuf == VK_LEFT:
			// left arrow -- scroll down one
			b[0] = CMD_SCROLL_UP

		case bbuf == VK_HOME:
			// home/SOF
			b[0] = CMD_SOF

		case bbuf == VK_END:
			// end/EOF
			b[0] = CMD_EOF

		case bbuf == VK_PRIOR:
			// PG UP
			b[0] = CMD_PAGE_UP

		case bbuf == VK_NEXT:
			// PG DN
			b[0] = CMD_PAGE_DN
		}

		// mode cancellations

		if bbuf != "" {
			if b[0] != MODE_TAIL {
				br.modeTail = false
			}

			if b[0] != MODE_UP {
				br.modeScrollUp = false
			}

			if b[0] != MODE_DN {
				br.modeScrollDown = false
			}
		}

		// commands

		switch {

		case b[0] == CMD_PAGE_DN:
			// page forward/down
			if !br.hitEOF {
				br.pageDown()
			}
			movecursor(2, 1, false)
			continue

		case b[0] == CMD_SCROLL_DN, b[0] == CMD_SCROLL_DN_1:
			// scroll forward/down
			br.scrollDown(1)
			movecursor(2, 1, false)
			continue

		case b[0] == MODE_DN:
			// toggle continuous scroll-down mode
			if br.modeScrollDown {
				br.modeScrollDown = false
				movecursor(2, 1, false)
			} else {
				br.modeScrollDown = true
				fmt.Printf("%s", CURRESTORE)
			}
			// modeTail is a faster version of modeScrollDown
			br.modeTail = false
			continue

		case b[0] == CMD_PAGE_UP:
			// page backward/up
			if br.firstRow > 0 {
				br.pageUp()
			}
			movecursor(2, 1, false)
			continue

		case b[0] == CMD_SCROLL_UP:
			// scroll backward/up
			br.scrollUp(1)
			movecursor(2, 1, false)
			continue

		case b[0] == MODE_UP:
			// toggle continuous scroll-up mode
			br.modeScrollUp = !br.modeScrollUp
			br.modeScrollDown = false
			continue

		case b[0] == CMD_SOF, b[0] == CMD_SOF_1:
			// beginning of file
			br.printPage(0)
			movecursor(2, 1, false)
			continue

		case b[0] == CMD_EOF, b[0] == CMD_EOF_1:
			// end of file
			br.pageLast()
			movecursor(2, 1, false)
			continue

		case b[0] == CMD_NUMBERS, b[0] == CMD_NUMBERS_1:
			// show line numbers
			br.modeNumbers = !br.modeNumbers
			br.pageCurrent()
			movecursor(2, 1, false)
			continue

		case b[0] == MODE_TAIL:
			// tail file
			if br.modeTail {
				br.modeTail = false
				movecursor(2, 1, false)
			} else {
				br.modeTail = true
				if !br.hitEOF {
					br.pageLast()
				}
				fmt.Printf("%s", CURRESTORE)
			}
			// modeScrollDown is a slower version of modeTail
			br.modeScrollDown = false
			continue

		case b[0] == CMD_JUMP:
			// jump to line
			lbuf := br.userInput("Junp: ")

			if lbuf == "" {
				br.pageCurrent()
			} else {
				var n int
				fmt.Sscanf(lbuf, "%d", &n)
				br.printPage(n)
			}
			movecursor(2, 1, false)
			continue

		case b[0] == CMD_SRCH_FWD:
			// search forward/down
			patbuf = br.userInput("/")
			searchDir = SEARCHFWD
			// null -- just changing direction -- don't reset
			if patbuf != "" {
				br.lastMatch = RESETSRCH
			}
			br.searchFile(patbuf, searchDir, false)
			continue

		case b[0] == CMD_SRCH_REV:
			// search backward/up
			patbuf = br.userInput("?")
			searchDir = SEARCHREV
			// null -- just changing direction -- don't reset
			if patbuf != "" {
				br.lastMatch = RESETSRCH
			}
			br.searchFile(patbuf, searchDir, false)
			continue

		case b[0] == CMD_SRCH_NEXT:
			br.searchFile(br.pattern, searchDir, true)
			continue

		case b[0] == CMD_SRCH_CLEAR:
			br.re = nil
			br.pattern = ""
			br.printMessage("OK")
			continue

		case b[0] == CMD_MARK:
			// mark page
			lbuf := br.userInput("Mark: ")
			if m := getMark(lbuf); m != 0 {
				br.marks[m] = br.firstRow
				br.printMessage("OK")
			}
			movecursor(2, 1, false)
			continue

		case b[0] == CMD_BASH:
			// bash command
			br.bashCommand()
			br.pageHeader()
			br.pageCurrent()
			continue

		case b[0] == CMD_QUIT:
			// quit -- this is the only way to save an rc file
			br.saveRC = true
			return

		case b[0] == CMD_QUIT_NO_SAVE:
			// Quit -- do not save an rc file
			br.saveRC = false
			return

		case b[0] == CMD_HELP:
			// help
			br.printHelp()
			continue

		default:
			// if digit, go to marked page
			if m := getMark(bbuf); m != 0 {
				br.pageMarked(m)
			}
			movecursor(2, 1, false) // no modes active
			continue
		}
	}
}

// vim: set ts=4 sw=4 noet:
