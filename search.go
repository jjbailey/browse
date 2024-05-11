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
	"strings"
)

// %6d + one space
const NUMCOLWIDTH = 7

func (x *browseObj) searchFile(pattern string, searchDir, next bool) {
	var sop, eop int
	var wrapped, warned bool
	var firstMatch, lastMatch int

	// to suppress S1002
	searchFwd := searchDir

	if pattern != x.pattern {
		// reset on first search
		x.lastMatch = SEARCH_RESET
		next = false
	}

	patternLen, err := x.reCompile(pattern)

	if err != nil {
		x.warnMessage(fmt.Sprintf("%v", err))
		return
	}

	if patternLen == 0 {
		x.warnMessage("No search pattern")
		return
	}

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
				x.warnMessage("Pattern not found: " + x.pattern)
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

	if lastMatch < firstMatch {
		lastMatch = firstMatch
	}

	return firstMatch, lastMatch
}

func (x *browseObj) lineIsMatch(lineno int) (int, string) {
	// check if this line has a regex match

	input := string(x.readFromMap(lineno))

	if x.noSearchPattern() {
		// no regex
		return 0, input
	}

	matches := x.re.FindAllStringIndex(input, -1)
	return len(matches), input
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

	// sop may be a negative number
	eop = sop + x.dispRows
	return sop, eop, wrapped
}

func (x *browseObj) replaceMatch(lineno int, input string) string {
	// make the regex replacements and return the new line

	var line string
	var leftMatch, rightMatch bool

	sol := minimum(x.shiftWidth, len(input))
	// NB: -2 from the end
	eol := minimum((sol+x.dispWidth), len(input)) - 2

	if x.modeNumbers {
		eol -= NUMCOLWIDTH
	}

	if x.noSearchPattern() {
		// no regex
		line = input[sol:]
	} else {
		// regex
		leftMatch, rightMatch = x.undisplayedMatches(input, sol, eol)

		if leftMatch || rightMatch {
			line = _VID_BOLD + x.re.ReplaceAllString(input[sol:], x.replstr)
		} else {
			line = x.re.ReplaceAllString(input[sol:], x.replstr)
		}
	}

	if x.modeNumbers {
		// line numbers -- uses NUMCOLWIDTH columns
		return fmt.Sprintf("%6d %s", lineno, line)
	}

	// no line numbers
	return line
}

func (x *browseObj) noSearchPattern() bool {
	return x.re == nil || len(x.re.String()) == 0
}

func (x *browseObj) doSearch(oldDir, newDir bool) bool {
	prompt, message := "/", "Searching forward"

	if !newDir {
		prompt, message = "?", "Searching reverse"
	}

	patbuf, cancel := x.userInput(prompt)

	if cancel {
		x.restoreLast()
		moveCursor(2, 1, false)
		return oldDir
	}

	if oldDir != newDir && (len(patbuf) > 0 || len(x.pattern) > 0) {
		// print direction
		x.timedMessage(message)
	}

	if len(patbuf) == 0 {
		// null -- change direction
		x.searchFile(x.pattern, newDir, true)
	} else {
		// search this page
		x.lastMatch = SEARCH_RESET
		x.searchFile(patbuf, newDir, false)
	}

	return newDir
}

func (x *browseObj) reCompile(pattern string) (int, error) {
	var cp string

	if len(pattern) == 0 {
		if len(x.pattern) == 0 {
			return 0, nil
		}

		pattern = x.pattern
	}

	if strings.HasPrefix(pattern, "(?i)") {
		x.ignoreCase = true
		pattern = strings.TrimPrefix(pattern, "(?i)")
	}

	if x.ignoreCase {
		cp = "(?i)" + pattern
	} else {
		cp = pattern
	}

	re, err := regexp.Compile(cp)

	if err != nil {
		return 0, err
	}

	x.pattern = pattern
	x.re = re
	x.replstr = fmt.Sprintf("%s%s%s", VIDPATTERN, "$0", VIDOFF)

	return len(pattern), nil
}

func (x *browseObj) undisplayedMatches(input string, sol, eol int) (bool, bool) {
	// check for matches to the left and right of the line as displayed

	var leftMatch, rightMatch bool

	for _, index := range x.re.FindAllStringSubmatchIndex(input, -1) {
		if index[0] < sol {
			leftMatch = true
		} else if index[0] > eol {
			rightMatch = true
		}

		if leftMatch && rightMatch {
			break
		}
	}

	return leftMatch, rightMatch
}

// vim: set ts=4 sw=4 noet:
