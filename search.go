// search.go
// search the file for a given regex
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"regexp"
)

// %6d + one space
const NUMCOLWIDTH = 7

func (x *browseObj) searchFile(pattern string, searchDir, next bool) {
	var sop, eop int
	var err error
	var wrapped, warned bool
	var firstMatch, lastMatch int

	// to suppress S1002
	searchFwd := searchDir

	if len(pattern) == 0 {
		if len(x.pattern) == 0 {
			x.printMessage("No search pattern")
			return
		}

		pattern = x.pattern
		// reset on first search
		x.lastMatch = SEARCH_RESET
		next = false
	}

	x.re, err = regexp.Compile(pattern)
	// save in case regexp.Compile failed
	x.pattern = pattern

	if err != nil {
		x.printMessage(fmt.Sprintf("%v", err))
		return
	}

	// save regexp.Compile source and replstr
	x.pattern = x.re.String()
	x.replstr = fmt.Sprintf("%s%s%s", VIDPATTERN, "$0", VIDOFF)

	// where to start search

	if x.lastMatch == SEARCH_RESET {
		// new search
		sop = x.firstRow
		eop = sop + x.dispRows
	} else if next {
		sop, eop, wrapped = x.setNextPage(searchDir, x.firstRow)
	}

	warned = false

	for {
		firstMatch, lastMatch = x.pageIsMatch(sop, eop)

		if wrapped {
			if warned {
				x.printMessage("Pattern not found: " + x.pattern)
				return
			}

			if searchFwd {
				x.timedMessage("Resuming search from SOF")
			} else {
				x.timedMessage("Resuming search from EOF")
			}

			warned = true
		}

		if firstMatch == 0 || lastMatch == 0 {
			sop, eop, wrapped = x.setNextPage(searchDir, sop)
			continue
		}

		// display strategy: go to the page wherever the next match occurs

		if PAGE_SEARCH || x.lastMatch == SEARCH_RESET {
			x.printPage(sop)
			return
		}

		// display strategy: reposition the page to provide match context

		if searchFwd {
			x.printPage(firstMatch - (x.dispRows >> 3))
		} else {
			x.printPage(lastMatch - (x.dispRows - (x.dispRows >> 3)))
		}

		return
	}
}

func (x *browseObj) pageIsMatch(sop, eop int) (int, int) {
	// return the first and last regex match on the page

	firstMatch := 0
	lastMatch := 0

	for lineno := sop; lineno < eop; lineno++ {
		if matches, _ := x.lineIsMatch(lineno); matches > 0 {
			if firstMatch == 0 {
				firstMatch = lineno
			}

			lastMatch = lineno
		}
	}

	if lastMatch == 0 {
		lastMatch = firstMatch
	}

	return firstMatch, lastMatch
}

func (x *browseObj) lineIsMatch(lineno int) (int, string) {
	// check if this line has a regex match

	var n int

	data, nbytes := x.readFromMap(lineno)

	// match on what's visible

	if !x.modeNumbers {
		n = minimum(nbytes, x.dispWidth)
	} else {
		// line numbers -- uses NUMCOLWIDTH columns
		n = minimum(nbytes, x.dispWidth-NUMCOLWIDTH)
	}

	input := string(data[0:n])

	if x.noSearchPattern() {
		// no regex
		return 0, input
	}

	return len(x.re.FindAllString(input, -1)), input
}

func (x *browseObj) setNextPage(searchDir bool, sop int) (int, int, bool) {
	// figure out which page to search next

	var eop int
	var wrapped bool

	// to suppress S1002
	searchFwd := searchDir

	if searchFwd {
		sop += x.dispRows
		if sop >= x.mapSiz {
			sop = 0
			wrapped = true
		}
	} else {
		sop -= x.dispRows
		if (sop + x.dispRows) < 0 {
			sop = maximum(x.mapSiz-x.dispRows, 0)
			wrapped = true
		}
	}

	// sop map be a negative number
	eop = sop + x.dispRows
	return sop, eop, wrapped
}

func (x *browseObj) replaceMatch(lineno int, input string) string {
	// make the regex replacements, return the new line

	var line string
	var output string

	if x.noSearchPattern() {
		// no regex
		line = input
	} else {
		line = x.re.ReplaceAllString(input, x.replstr)
	}

	if x.modeNumbers && !windowAtEOF(lineno, x.mapSiz) {
		// line numbers -- uses NUMCOLWIDTH columns
		output = fmt.Sprintf("%6d %s", lineno, line)
	} else {
		// no line numbers
		output = line
	}

	return output
}

func (x *browseObj) noSearchPattern() bool {
	return x.re == nil || len(x.re.String()) == 0
}

func (x *browseObj) doSearch(oldDir, newDir bool) bool {
	var prompt, message string

	// search direction
	if newDir {
		prompt, message = "/", "Searching forward"
	} else {
		prompt, message = "?", "Searching reverse"
	}

	patbuf, cancel := x.userInput(prompt)

	if cancel {
		x.restoreLast()
		movecursor(2, 1, false)
		return oldDir
	}

	if len(patbuf) == 0 {
		// null -- change direction
		if oldDir != newDir {
			x.timedMessage(message)
		}
		x.searchFile(x.pattern, newDir, true)
	} else {
		// search this page
		x.lastMatch = SEARCH_RESET
		x.searchFile(patbuf, newDir, false)
	}

	return newDir
}

// vim: set ts=4 sw=4 noet:
