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
		// should not happen
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

	if br.lastMatch == SEARCH_RESET {
		startOfPage = br.firstRow
		endOfPage = startOfPage + dispRows
	} else if next {
		startOfPage, endOfPage, wrapped = br.setNextPage(forward, br.firstRow)
	} else {
		// If not a new search and not continuing, use current page
		startOfPage = br.firstRow
		endOfPage = startOfPage + dispRows
	}

	for {
		firstMatch, lastMatch := br.pageIsMatch(startOfPage, endOfPage)

		if wrapped && warned {
			br.timedMessage("Pattern not found: "+br.pattern, MSG_ORANGE)
			br.pageCurrent()
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

	if br.noSearchPattern() || br.re == nil {
		return 0, string(br.readFromMap(lineno))
	}

	lineContent := string(br.readFromMap(lineno))
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

	dispRows := br.dispRows

	if forward {
		// Forward search
		newStart = startOfPage + dispRows
		if newStart >= br.mapSiz {
			// Wrap to start of file
			newStart, wrapped = 0, true
		}
	} else {
		// Reverse search
		switch {

		case startOfPage > dispRows:
			// Previous page
			newStart, wrapped = startOfPage-dispRows, false

		case br.lastMatch < dispRows || startOfPage < dispRows:
			// Either already searched top page, or already at top -- wrap to end
			newStart, wrapped = maximum(br.mapSiz-dispRows, 0), true

		default:
			// Top page but not wrapped → go to beginning
			newStart, wrapped = 0, false
		}
	}

	newEnd = newStart + dispRows
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

	if br.noSearchPattern() || br.re == nil {
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

func (br *browseObj) noSearchPattern() bool {
	return br.re == nil || br.re.String() == ""
}

func (br *browseObj) doSearch(oldDir, newDir bool) bool {
	// Search starts here

	moveCursor(br.dispHeight-1, 1, true)
	patbuf, cancelled := userSearchComp(newDir)

	if cancelled {
		// user backed out of prompt
		br.pageCurrent()
		return oldDir
	}

	// Save search to history
	history := loadHistory(searchHistory)
	history = append(history, patbuf)
	saveHistory(history, searchHistory)

	if br.pattern != "" {
		// substitute & with the current search pattern
		patbuf = subCommandChars(patbuf, "&", br.pattern)
	}

	if oldDir != newDir && (len(patbuf) > 0 || len(br.pattern) > 0) {
		// print direction
		if newDir {
			br.timedMessage("Searching forward", MSG_GREEN)
		} else {
			br.timedMessage("Searching reverse", MSG_GREEN)
		}
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
