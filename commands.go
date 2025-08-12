// commands.go
// the command processor
// all user activity starts here
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// ─── Command Groups ─────────────────────────────────────────────────

const (
	// Navigation commands
	CMD_PAGE_DN        = 'f'
	CMD_PAGE_DN_1      = ' '
	CMD_PAGE_UP        = 'b'
	CMD_HALF_PAGE_DN   = '\006'
	CMD_HALF_PAGE_DN_1 = '\004'
	CMD_HALF_PAGE_DN_2 = 'z'
	CMD_HALF_PAGE_UP   = '\002'
	CMD_HALF_PAGE_UP_1 = '\025'
	CMD_HALF_PAGE_UP_2 = 'Z'
	CMD_SCROLL_DN      = '+'
	CMD_SCROLL_DN_1    = '\r'
	CMD_SCROLL_UP      = '-'
	CMD_MODE_DN        = 'd'
	CMD_MODE_UP        = 'u'
	CMD_MODE_TAIL      = 't'
	CMD_MODE_FOLLOW    = 'e'
	CMD_SOF            = '0'
	CMD_EOF            = 'G'

	// Search commands
	CMD_SEARCH_FWD      = '/'
	CMD_SEARCH_REV      = '?'
	CMD_SEARCH_NEXT     = 'n'
	CMD_SEARCH_NEXT_REV = 'N'
	CMD_SEARCH_IGN_CASE = 'i'
	CMD_GREP            = '&'
	CMD_SEARCH_PRINT    = 'p'
	CMD_SEARCH_CLEAR    = 'P'

	// Horizontal scrolling commands
	CMD_SHIFT_LEFT    = '<'
	CMD_SHIFT_LEFT_1  = '\b'
	CMD_SHIFT_LEFT_2  = '\177'
	CMD_SHIFT_RIGHT   = '>'
	CMD_SHIFT_RIGHT_1 = '\011'
	CMD_SHIFT_ZERO    = '^'
	CMD_SHIFT_LONGEST = '$'

	// File operations
	CMD_NEWFILE      = 'B'
	CMD_QUIT         = 'q'
	CMD_QUIT_NO_SAVE = 'Q'
	CMD_EXIT         = 'x'
	CMD_EXIT_NO_SAVE = 'X'

	// Other commands
	CMD_BASH      = '!'
	CMD_HELP      = 'h'
	CMD_JUMP      = 'j'
	CMD_MARK      = 'm'
	CMD_NUMBERS   = '#'
	CMD_PERCENT   = '%'
	CMD_PERCENT_1 = '\007'
)

// ─── Virtual Key Mappings ───────────────────────────────────────────

const (
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

// ─── Search and Scroll Constants ────────────────────────────────────

const (
	SEARCH_FWD  = true
	SEARCH_REV  = false
	SCROLL_TAIL = 256
	SCROLL_CONT = 2
)

// searchDir controls the direction of search operations
var searchDir bool = SEARCH_FWD

func commands(br *browseObj) {
	// seed the saved search pattern
	br.reCompile(br.pattern)

	// wait for a full page
	waitForInput(br)

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

	// handle panic
	defer handlePanic(br)

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
				br.toggleMode(MODE_SCROLL_UP)

			case CMD_MODE_DN:
				// toggle scroll down mode
				br.toggleMode(MODE_SCROLL_DN)

			case CMD_MODE_TAIL:
				// toggle tail mode
				br.toggleMode(MODE_SCROLL_TAIL)

			case CMD_MODE_FOLLOW:
				// toggle follow mode
				br.toggleMode(MODE_SCROLL_FOLLOW)

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
			moveCursor(2, 1, false)

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

		case CMD_SHIFT_LEFT, CMD_SHIFT_LEFT_1, CMD_SHIFT_LEFT_2:
			// horizontal scroll left
			if br.shiftWidth >= TABWIDTH {
				br.shiftWidth -= TABWIDTH
				br.pageCurrent()
			}
			moveCursor(2, 1, false)

		case CMD_SHIFT_RIGHT, CMD_SHIFT_RIGHT_1:
			// horizontal scroll right
			if br.shiftWidth < (READBUFSIZ - (TABWIDTH * 2)) {
				br.shiftWidth += TABWIDTH
				br.pageCurrent()
			}
			moveCursor(2, 1, false)

		case CMD_SHIFT_ZERO:
			// horizontal scroll left to column 1
			if br.shiftWidth > 0 {
				br.shiftWidth = 0
				br.pageCurrent()
			}
			moveCursor(2, 1, false)

		case CMD_SHIFT_LONGEST:
			// horizontal scroll longest
			br.shiftWidth = shiftLongest(br)
			br.pageCurrent()
			moveCursor(2, 1, false)

		case CMD_SOF:
			// beginning of file at column 1
			br.shiftWidth = 0
			br.printPage(0)

		case CMD_EOF:
			// end of file
			br.pageLast()

		case CMD_NUMBERS:
			// show line numbers
			br.modeNumbers = !br.modeNumbers
			br.pageCurrent()

		case CMD_MODE_TAIL, CMD_MODE_FOLLOW:
			// tail or follow
			br.pageLast()
			if br.inMotion() {
				fmt.Print(CURRESTORE)
			}

		case CMD_JUMP:
			// jump to line
			lbuf, cancelled, _ := br.userInput("Jump: ")
			if !cancelled && len(lbuf) > 0 {
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
				br.printMessage("Search ignores case", MSG_GREEN)
			} else {
				br.printMessage("Search considers case", MSG_GREEN)
			}

		case CMD_GREP:
			// grep -nP pattern
			br.runGrep()

		case CMD_SEARCH_PRINT:
			// print the search pattern
			if len(br.pattern) == 0 {
				br.printMessage("No search pattern", MSG_ORANGE)
			} else {
				br.printMessage(br.pattern, MSG_GREEN)
			}

		case CMD_SEARCH_CLEAR:
			// clear the search pattern
			br.re = nil
			br.pattern = ""
			br.printMessage("Search pattern cleared", MSG_GREEN)

		case CMD_MARK:
			// mark page
			lbuf, cancelled, _ := br.userInput("Mark: ")
			if !cancelled && len(lbuf) > 0 {
				if m := getMark(lbuf); m != 0 {
					br.marks[m] = br.firstRow
					br.printMessage(fmt.Sprintf("Mark %d at line %d", m, br.marks[m]), MSG_GREEN)
				}
			}

		case CMD_BASH:
			br.bashCommand()

		case CMD_HALF_PAGE_DN, CMD_HALF_PAGE_DN_1, CMD_HALF_PAGE_DN_2:
			// scroll half page forward/down
			br.scrollDown(br.dispRows >> 1)
			moveCursor(2, 1, false)

		case CMD_HALF_PAGE_UP, CMD_HALF_PAGE_UP_1, CMD_HALF_PAGE_UP_2:
			// scroll half page backward/up
			br.scrollUp(br.dispRows >> 1)

		case CMD_PERCENT, CMD_PERCENT_1:
			// page position
			// -1 for SOF
			t := float32(br.firstRow) / float32(br.mapSiz-1) * 100.0
			br.printMessage(fmt.Sprintf("\"%s\" %d lines --%1.1f%%--",
				filepath.Base(br.fileName), br.mapSiz-1, t), MSG_GREEN)

		case CMD_NEWFILE:
			if fileCommand(br) {
				return
			}

		case CMD_QUIT:
			br.saveRC = true
			br.exit = false
			return

		case CMD_QUIT_NO_SAVE:
			br.saveRC = false
			br.exit = false
			return

		case CMD_EXIT:
			br.saveRC = true
			br.exit = true
			return

		case CMD_EXIT_NO_SAVE:
			br.saveRC = false
			br.exit = true
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

func fileCommand(br *browseObj) bool {
	// Browse a new file

	moveCursor(br.dispHeight, 1, true)
	lbuf, cancelled := userFileComp()
	file := strings.TrimSpace(lbuf)

	// Remove single and double quotes from the input
	file = strings.ReplaceAll(file, "'", "")
	file = strings.ReplaceAll(file, "\"", "")

	if !cancelled && len(file) > 0 {
		sbuf := subCommandChars(file, "%", br.fileName)
		if browseFile(br, sbuf, setTitle(sbuf, sbuf), false, true) {
			return true
		}
	}

	br.pageCurrent()
	return false
}

func waitForInput(br *browseObj) {
	// attempt to read an entire page
	// strike a balance between waiting for a page and small files

	const (
		maxAttempts      = 20
		stableThreshold  = 10
		waitInterval     = 100 * time.Millisecond
		modTimeThreshold = 5 * time.Second
	)

	targetSize := br.firstRow + br.dispHeight
	lastMapSize := br.mapSiz
	stableCount := 0

	for attempt := 0; attempt < maxAttempts; attempt++ {
		info, err := os.Stat(br.fileName)
		if err == nil {
			if time.Since(info.ModTime()) > modTimeThreshold {
				// don't wait for files unchanged in the last 5 seconds
				break
			}
		}

		if br.mapSiz >= targetSize {
			break
		}

		if br.mapSiz == lastMapSize {
			stableCount++
		} else {
			stableCount = 0
			lastMapSize = br.mapSiz
		}

		if stableCount > stableThreshold {
			break
		}

		time.Sleep(waitInterval)
	}

	if br.mapSiz < targetSize {
		// did not get a full read -- reset
		// +2 for header, EOF
		br.firstRow = maximum(0, br.mapSiz-br.dispHeight+2)
	}
}

func shiftLongest(br *browseObj) int {
	// shift to the end of the longest line on the page

	if TABWIDTH == 0 {
		return 0
	}

	longest := 0
	lastRow := minimum(br.firstRow+br.dispRows, br.mapSiz)

	for i := br.firstRow; i < lastRow; i++ {
		line := br.readFromMap(i)
		if line == nil {
			continue
		}
		lineLength := len(line)

		if lineLength > longest {
			longest = lineLength
		}
	}

	if br.modeNumbers {
		longest += NUMCOLWIDTH
	}

	if longest <= br.dispWidth {
		return 0
	}

	return ((longest - br.dispWidth + TABWIDTH) / TABWIDTH) * TABWIDTH
}

func handlePanic(br *browseObj) {
	// graceful exit

	if r := recover(); r != nil {
		moveCursor(br.dispRows, 1, true)
		fmt.Printf("%s%s panic: %v %s\n", CLEARSCREEN, MSG_RED, r, VIDOFF)
		br.saneExit()
	}
}

// vim: set ts=4 sw=4 noet:
