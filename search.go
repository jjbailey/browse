// search.go
// search the file for a given regex
//
// Copyright (c) 2024-2025 jjb
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

const (
	// %6d + one space
	LINENUMBERS = "%6d %s"
	NUMCOLWIDTH = 7
)

func (br *browseObj) searchFile(pattern string, searchDir, next bool) bool {
	var sop, eop int
	var wrapped, warned bool
	var firstMatch, lastMatch int

	// to suppress S1002
	searchFwd := searchDir

	if pattern != br.pattern {
		// reset on first search
		br.lastMatch = SEARCH_RESET
		next = false
	}

	patternLen, err := br.reCompile(pattern)

	if err != nil {
		br.printMessage(fmt.Sprintf("%v", err), MSG_ORANGE)
		return false
	}

	if patternLen == 0 {
		br.printMessage("No search pattern", MSG_ORANGE)
		return false
	}

	// where to start search

	if br.lastMatch == SEARCH_RESET {
		// new search
		sop = br.firstRow
		eop = sop + br.dispRows
	} else if next {
		sop, eop, wrapped = br.setNextPage(searchDir, br.firstRow)
	}

	warned = false

	for {
		firstMatch, lastMatch = br.pageIsMatch(sop, eop)

		if wrapped {
			if warned {
				br.printMessage("Pattern not found: "+br.pattern, MSG_ORANGE)
				return false
			}

			if searchFwd {
				br.timedMessage("Resuming search from SOF", MSG_GREEN)
			} else {
				br.timedMessage("Resuming search from EOF", MSG_GREEN)
			}

			warned = true
		}

		if firstMatch == 0 || lastMatch == 0 {
			sop, eop, wrapped = br.setNextPage(searchDir, sop)
			continue
		}

		// display strategy: go to the page wherever the next match occurs

		if br.lastMatch == SEARCH_RESET {
			br.printPage(sop)
			return true
		}

		// display strategy: reposition the page to provide match context
		// 1/8 forward, 7/8 reverse

		if searchFwd {
			br.printPage(firstMatch - (br.dispRows >> 3))
		} else {
			br.printPage(lastMatch - (br.dispRows>>3)*7)
		}

		return true
	}
}

func (br *browseObj) pageIsMatch(sop, eop int) (int, int) {
	// return the first and last regex match on the page

	firstMatch := 0
	lastMatch := 0
	foundMatch := false

	for lineno := sop; lineno < eop; lineno++ {
		if matches, _ := br.lineIsMatch(lineno); matches > 0 {
			if !foundMatch {
				firstMatch = lineno
				foundMatch = true
			}

			lastMatch = lineno
		}
	}

	if lastMatch < firstMatch {
		lastMatch = firstMatch
	}

	return firstMatch, lastMatch
}

func (br *browseObj) lineIsMatch(lineno int) (int, string) {
	// check if this line has a regex match

	input := string(br.readFromMap(lineno))

	if br.noSearchPattern() {
		// no regex
		return 0, input
	}

	matches := br.re.FindAllStringIndex(input, -1)
	return len(matches), input
}

func (br *browseObj) setNextPage(searchDir bool, sop int) (int, int, bool) {
	// figure out which page to search next

	var eop int
	var wrapped bool

	// to suppress S1002
	searchFwd := searchDir

	if searchFwd {
		sop += br.dispRows
		if sop >= br.mapSiz {
			sop = 0
			wrapped = true
		}
	} else {
		sop -= br.dispRows
		if (sop + br.dispRows) < 0 {
			sop = maximum(br.mapSiz-br.dispRows, 0)
			wrapped = true
		}
	}

	// sop may be a negative number
	eop = sop + br.dispRows
	return sop, eop, wrapped
}

func (br *browseObj) replaceMatch(lineno int, input string) string {
	// make the regex replacements
	// return the new line with or without line numbers as necessary

	var line string
	sol := br.shiftWidth

	if sol >= len(input) {
		if br.modeNumbers {
			return fmt.Sprintf(LINENUMBERS, lineno, "")
		}

		return ""
	}

	if br.noSearchPattern() {
		if br.modeNumbers {
			return fmt.Sprintf(LINENUMBERS, lineno, input[sol:])
		}

		return input[sol:]
	}

	// regex
	leftMatch, rightMatch := br.undisplayedMatches(input, sol)

	if leftMatch || rightMatch {
		line = _VID_GREEN_FG + br.re.ReplaceAllString(input[sol:], br.replace+_VID_GREEN_FG)
	} else {
		line = br.re.ReplaceAllString(input[sol:], br.replace)
	}

	if br.modeNumbers {
		return fmt.Sprintf(LINENUMBERS, lineno, line)
	}

	return line
}

func (br *browseObj) noSearchPattern() bool {
	return br.re == nil || len(br.re.String()) == 0
}

func (br *browseObj) doSearch(oldDir, newDir bool) bool {
	prompt, message := "/", "Searching forward"

	if !newDir {
		prompt, message = "?", "Searching reverse"
	}

	patbuf, cancel := br.userInput(prompt)

	if cancel {
		br.restoreLast()
		moveCursor(2, 1, false)
		return oldDir
	}

	if oldDir != newDir && (len(patbuf) > 0 || len(br.pattern) > 0) {
		// print direction
		br.timedMessage(message, MSG_GREEN)
	}

	if len(patbuf) == 0 {
		// null -- change direction
		br.searchFile(br.pattern, newDir, true)
	} else {
		// search this page
		br.lastMatch = SEARCH_RESET
		br.searchFile(patbuf, newDir, false)
	}

	return newDir
}

func (br *browseObj) reCompile(pattern string) (int, error) {
	var cp string

	if len(pattern) == 0 {
		if len(br.pattern) == 0 {
			return 0, nil
		}

		pattern = br.pattern
	}

	if strings.HasPrefix(pattern, "(?i)") {
		br.ignoreCase = true
		pattern = strings.TrimPrefix(pattern, "(?i)")
	}

	if br.ignoreCase {
		cp = "(?i)" + pattern
	} else {
		cp = pattern
	}

	re, err := regexp.Compile(cp)

	if err != nil {
		return 0, err
	}

	br.pattern = pattern
	br.re = re
	br.replace = fmt.Sprintf("%s%s%s", MSG_GREEN, "$0", VIDOFF)

	return len(pattern), nil
}

func (br *browseObj) undisplayedMatches(input string, sol int) (bool, bool) {
	matches := br.re.FindAllStringSubmatchIndex(input, -1)

	if len(matches) == 0 {
		return false, false
	}

	leftMatch := false
	rightMatch := false
	displayWidth := br.dispWidth

	if br.modeNumbers {
		displayWidth -= NUMCOLWIDTH
	}

	for _, index := range matches {
		if index[0] < sol {
			leftMatch = true

			if rightMatch {
				break
			}

			continue
		}

		if index[0]-br.shiftWidth+2 > displayWidth {
			// NB: off by two
			rightMatch = true

			if leftMatch {
				break
			}
		}
	}

	return leftMatch, rightMatch
}

// vim: set ts=4 sw=4 noet:
