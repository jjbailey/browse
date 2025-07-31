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

func (br *browseObj) searchFile(pattern string, forward, next bool) bool {
	// forward: true = forward, false = reverse
	// next: true = continue search, false = new search

	var err error
	var patternLen int

	// Reset search state if pattern changed
	if len(pattern) == 0 || pattern != br.pattern {
		br.lastMatch = SEARCH_RESET
		br.re = nil
		next = false

		patternLen, err = br.reCompile(pattern)
		if err != nil {
			br.printMessage(fmt.Sprintf("%v", err), MSG_ORANGE)
			return false
		}

		if patternLen == 0 {
			br.printMessage("No search pattern", MSG_ORANGE)
			return false
		}
	}

	// Determine start and end of page
	var startOfPage, endOfPage int
	var wrapped, warned bool

	if br.lastMatch == SEARCH_RESET {
		startOfPage = br.firstRow
		endOfPage = startOfPage + br.dispRows
	} else if next {
		startOfPage, endOfPage, wrapped = br.setNextPage(forward, br.firstRow)
	}

	for {
		firstMatch, lastMatch := br.pageIsMatch(startOfPage, endOfPage)

		if wrapped {
			if warned {
				br.printMessage("Pattern not found: "+br.pattern, MSG_ORANGE)
				return false
			}

			br.displayWrapMessage(forward)
			warned = true
		}

		if firstMatch == 0 || lastMatch == 0 {
			startOfPage, endOfPage, wrapped = br.setNextPage(forward, startOfPage)
			continue
		}

		// Display strategy: go to the page wherever the next match occurs
		if br.lastMatch == SEARCH_RESET {
			br.printPage(startOfPage)
			return true
		}

		// Display strategy: reposition the page to provide match context
		// 1/6 searching down, 5/6 searching up
		downOffset := br.dispRows / 6
		upOffset := downOffset * 5

		if forward {
			br.printPage(firstMatch - downOffset)
		} else {
			br.printPage(lastMatch - upOffset)
		}

		return true
	}
}

func (br *browseObj) displayWrapMessage(forward bool) {
	// displayWrapMessage prints a message when the search wraps around the file

	if forward {
		br.timedMessage("Resuming search from SOF", MSG_GREEN)
	} else {
		br.timedMessage("Resuming search from EOF", MSG_GREEN)
	}
}

func (br *browseObj) pageIsMatch(startOfPage, endOfPage int) (int, int) {
	// return the first and last regex match on the page

	var (
		firstMatchLine int
		lastMatchLine  int
		hasMatch       bool
	)

	for lineNum := startOfPage; lineNum < endOfPage; lineNum++ {
		matchCount, _ := br.lineIsMatch(lineNum)
		if matchCount == 0 {
			continue
		}

		if !hasMatch {
			firstMatchLine = lineNum
			hasMatch = true
		}

		lastMatchLine = lineNum
	}

	if !hasMatch {
		return 0, 0
	}

	return firstMatchLine, lastMatchLine
}

func (br *browseObj) lineIsMatch(lineno int) (int, string) {
	// check if this line has a regex match

	lineContent := string(br.readFromMap(lineno))

	if br.noSearchPattern() {
		// no regex
		return 0, lineContent
	}

	matchIndices := br.re.FindAllStringIndex(lineContent, -1)

	return len(matchIndices), lineContent
}

func (br *browseObj) setNextPage(forward bool, startOfPage int) (int, int, bool) {
	// figure out which page to search next

	var (
		newStart int
		newEnd   int
		wrapped  bool
	)

	if forward {
		// Forward search: move down by page size
		newStart = startOfPage + br.dispRows
		if newStart >= br.mapSiz {
			// Wrap to start of file
			newStart = 1
			wrapped = true
		}
	} else {
		// Reverse search: move up by page size
		if startOfPage > br.dispRows {
			// Page above
			newStart = startOfPage - br.dispRows
			wrapped = false
		} else if startOfPage > 0 {
			// Top page
			newStart = maximum(startOfPage-br.dispRows, 1)
			wrapped = false
		} else {
			// Wrap to end of file
			newStart = maximum(br.mapSiz-br.dispRows, 0)
			wrapped = true
		}
	}

	newEnd = newStart + br.dispRows
	return newStart, newEnd, wrapped
}

func (br *browseObj) replaceMatch(lineno int, input string) string {
	// Return the new line with or without line numbers, applying regex replacements as needed

	sol := br.shiftWidth

	// If the shifted start is past the end of the input, return blank/empty
	if sol >= len(input) {
		if br.modeNumbers {
			return fmt.Sprintf(LINENUMBERS, lineno, "")
		}

		return ""
	}

	// If no search pattern, just return the line (possibly with numbers)
	if br.noSearchPattern() {
		line := input[sol:]
		if br.modeNumbers {
			return fmt.Sprintf(LINENUMBERS, lineno, line)
		}

		return line
	}

	// There is a search pattern: do regex replacement and possibly highlight
	leftMatch, rightMatch := br.undisplayedMatches(input, sol)

	var replaced string
	if leftMatch || rightMatch {
		replaced = _VID_GREEN_FG + br.re.ReplaceAllString(input[sol:], br.replace+_VID_GREEN_FG)
	} else {
		replaced = br.re.ReplaceAllString(input[sol:], br.replace)
	}

	if br.modeNumbers {
		return fmt.Sprintf(LINENUMBERS, lineno, replaced)
	}

	return replaced
}

func (br *browseObj) noSearchPattern() bool {
	return br.re == nil || len(br.re.String()) == 0
}

func (br *browseObj) doSearch(oldDir, newDir bool) bool {
	prompt, message := "/", "Searching forward"
	if !newDir {
		prompt, message = "?", "Searching reverse"
	}

	patbuf, _, ignore := br.userInput(prompt)
	if ignore {
		// user backed out of prompt
		return oldDir
	}

	if br.pattern != "" {
		// substitute & with the current search pattern
		patbuf = subCommandChars(patbuf, "&", br.pattern)
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

	displayWidth := br.dispWidth
	if br.modeNumbers {
		displayWidth -= NUMCOLWIDTH
	}

	leftMatch, rightMatch := false, false

	for _, index := range matches {
		if !leftMatch && index[0] < sol {
			leftMatch = true
		}

		if !rightMatch && index[0]-br.shiftWidth+2 > displayWidth {
			// NB: off by two
			rightMatch = true
		}

		if leftMatch && rightMatch {
			break
		}
	}

	return leftMatch, rightMatch
}

// vim: set ts=4 sw=4 noet:
