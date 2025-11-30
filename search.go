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

	if pattern == "" {
		br.printMessage("No search pattern", MSG_ORANGE)
		return false
	}

	// Reset search state if pattern changed
	if pattern != br.pattern {
		br.lastMatch = SEARCH_RESET
		br.re = nil
		next = false

		patternLen, err = br.reCompile(pattern)
		if err != nil {
			br.printMessage(fmt.Sprintf("Regex compilation error: %v", err), MSG_ORANGE)
			return false
		}

		if patternLen == 0 {
			br.printMessage("Empty search pattern", MSG_ORANGE)
			return false
		}
	}

	dispRows := br.dispRows

	// Determine start and end of page
	var startOfPage, endOfPage int
	var wrapped, warned bool

	if br.lastMatch == SEARCH_RESET || !next {
		// New search or not continuing - use current page
		startOfPage = br.firstRow
		endOfPage = startOfPage + dispRows
	} else {
		// Continuing search - get next page
		startOfPage, endOfPage, wrapped = br.setNextPage(forward, br.firstRow)
	}

	for {
		firstMatch, lastMatch := br.pageIsMatch(startOfPage, endOfPage)

		if wrapped && warned {
			br.printMessage("Pattern not found: "+br.pattern, MSG_ORANGE)
			moveCursor(2, 1, false)
			return false
		}

		if wrapped && !warned {
			br.displayWrapMessage(forward)
			warned = true
		}

		if firstMatch < 0 {
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
		downOffset := dispRows / 6
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
		firstMatchLine int = -1
		lastMatchLine  int = -1
	)

	for lineNum := startOfPage; lineNum < endOfPage; lineNum++ {
		matchCount, _ := br.lineIsMatch(lineNum)
		if matchCount == 0 {
			continue
		}

		if firstMatchLine == -1 {
			firstMatchLine = lineNum
		}

		lastMatchLine = lineNum
	}

	return firstMatchLine, lastMatchLine
}

func (br *browseObj) lineIsMatch(lineno int) (int, string) {
	// check if this line has a regex match

	if lineno < 0 || lineno >= br.mapSiz {
		return 0, ""
	}

	if br.re == nil {
		return 0, string(br.readFromMap(lineno))
	}

	lineContent := string(br.readFromMap(lineno))
	matchIndices := br.re.FindAllStringIndex(lineContent, -1)
	return len(matchIndices), lineContent
}

func (br *browseObj) setNextPage(forward bool, startOfPage int) (int, int, bool) {
	dispRows := br.dispRows
	totalRows := br.mapSiz
	wrapped := false

	if totalRows <= 0 {
		return 0, 0, false
	}

	var newStart, newEnd int

	if forward {
		newStart = startOfPage + dispRows
		if newStart >= totalRows {
			newStart = 0
			wrapped = true
		}
		newEnd = minimum(newStart+dispRows, totalRows)
	} else {
		if startOfPage == 0 {
			// At SOF, wrap to last page at EOF
			newStart = maximum(totalRows-dispRows, 0)
			newEnd = totalRows
			wrapped = true
		} else {
			// Go up one page
			newEnd = startOfPage
			newStart = 0
			if newEnd > dispRows {
				newStart = newEnd - dispRows
			}
		}
	}

	return newStart, newEnd, wrapped
}

func (br *browseObj) replaceMatch(lineno int, input string) string {
	sol := br.shiftWidth

	// Slice safely
	var content string

	if sol < len(input) {
		content = input[sol:]
	} else {
		content = ""
	}

	if br.re == nil {
		return br.formatLine(lineno, content)
	}

	leftMatch, rightMatch := br.undisplayedMatches(input, sol)

	if content == "" {
		if leftMatch {
			boldLeftArrow := _VID_BOLD + _VID_GREEN_FG + "\u2190" + VIDOFF
			return br.formatLine(lineno, boldLeftArrow)
		}

		return br.formatLine(lineno, "")
	}

	var replaced string

	if leftMatch || rightMatch {
		replaced = _VID_GREEN_FG + br.re.ReplaceAllString(content, br.replace+_VID_GREEN_FG)
	} else {
		replaced = br.re.ReplaceAllString(content, br.replace)
	}

	return br.formatLine(lineno, replaced)
}

func (br *browseObj) formatLine(lineno int, content string) string {
	if br.modeNumbers {
		return fmt.Sprintf(LINENUMBERS, lineno, content)
	}

	return content
}

func (br *browseObj) doSearch(oldDir, newDir bool) bool {
	moveCursor(br.dispRows, 1, true)

	pattern, cancelled := userSearchComp(newDir)
	br.shownMsg = true

	if cancelled {
		br.restoreLast()
		return oldDir
	}

	prevPattern := br.pattern
	if strings.TrimSpace(pattern) == "" {
		pattern = prevPattern
	}

	if pattern == "" {
		br.printMessage("No search pattern", MSG_ORANGE)
		return false
	}

	// Substitute '&' with previous pattern for continued searches
	if strings.Contains(pattern, "&") {
		pattern = subCommandChars(pattern, "&", prevPattern)
	}

	updateSearchHistory(pattern)

	if oldDir != newDir {
		if newDir {
			br.timedMessage("Searching forward", MSG_GREEN)
		} else {
			br.timedMessage("Searching reverse", MSG_GREEN)
		}
		br.lastMatch = SEARCH_RESET
	}

	continueSearch := (oldDir == newDir && pattern == prevPattern)
	br.searchFile(pattern, newDir, continueSearch)
	return newDir
}

func (br *browseObj) reCompile(pattern string) (int, error) {
	// Compile regex

	if pattern == "" {
		if br.pattern == "" {
			return 0, nil
		}

		pattern = br.pattern
	}

	if strings.HasPrefix(pattern, "(?i)") {
		br.ignoreCase = true
		pattern = strings.TrimPrefix(pattern, "(?i)")
	}

	var cp string

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
	// Safety check: ensure regex is compiled

	if br.re == nil {
		return false, false
	}

	// Use FindAllStringIndex (not Submatch) for efficiency
	matches := br.re.FindAllStringIndex(input, -1)
	if len(matches) == 0 {
		return false, false
	}

	displayWidth := br.dispWidth
	if br.modeNumbers {
		displayWidth -= NUMCOLWIDTH
	}

	leftMatch, rightMatch := false, false

	for _, index := range matches {
		// Ensure index has at least 2 elements (start and end positions)
		if len(index) < 2 {
			continue
		}

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
