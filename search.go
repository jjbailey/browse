// search.go
// search the file for a given regex
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"regexp"
)

// Search formatting and limits.
const (
	// %6d + one space
	NUMCOLWIDTH = 7

	// Maximum regex pattern length to avoid accidental oversized searches.
	MAX_PATTERN_LENGTH = 1000
)

// searchFile scans for a regex pattern and updates the view accordingly.
// forward: true = forward, false = reverse
// next: true = continue search, false = new search
func (br *browseObj) searchFile(pattern string, forward, next bool) bool {
	var err error
	var patternLen int

	if pattern == "" {
		br.printMessage("No search pattern", MSG_ORANGE)
		return false
	}

	freshPatternSearch := pattern != br.pattern

	// Reset search state after a new or missing regexp compiles successfully.
	if freshPatternSearch || br.re == nil {
		patternLen, err = br.reCompile(pattern)
		if err != nil {
			br.printMessage(fmt.Sprintf("Regex compilation error: %v", err), MSG_ORANGE)
			return false
		}

		if patternLen == 0 {
			br.printMessage("Empty search pattern", MSG_ORANGE)
			return false
		}

		br.lastMatch = SEARCH_RESET
		next = false
	}

	matchLine, wrapped := br.findSearchMatch(forward, next)
	if matchLine < 0 {
		br.printMessage("Pattern not found", MSG_ORANGE)
		moveCursor(2, 1, false)
		return false
	}

	if wrapped {
		br.displayWrapMessage(forward)
	}

	br.lastMatch = matchLine
	displayTop := br.searchDisplayTop(matchLine, forward)
	if freshPatternSearch && !wrapped && br.lineOnCurrentPage(matchLine) {
		displayTop = br.firstRow
	}
	br.printPage(displayTop)

	return true
}

// lineOnCurrentPage reports whether a line is visible before search repositions.
func (br *browseObj) lineOnCurrentPage(lineNum int) bool {
	return lineNum >= br.firstRow && lineNum < br.firstRow+br.dispRows
}

// searchMapSize returns a stable snapshot of the currently mapped line count.
func (br *browseObj) searchMapSize() int {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	return br.mapSiz
}

// lineInMap reports whether a line is currently mapped.
func (br *browseObj) lineInMap(lineno int) bool {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	return lineno >= 0 && lineno < br.mapSiz
}

// displayWrapMessage informs the user when the search wraps.
func (br *browseObj) displayWrapMessage(forward bool) {
	if forward {
		br.timedMessage("Resuming search from SOF", MSG_GREEN)
	} else {
		br.timedMessage("Resuming search from EOF", MSG_GREEN)
	}
}

// findSearchMatch returns the next matching line without changing display state.
func (br *browseObj) findSearchMatch(forward, next bool) (int, bool) {
	mapSize := br.searchMapSize()
	if mapSize <= 0 {
		return -1, false
	}

	startLine := br.searchStartLine(forward, next, mapSize)

	if forward {
		if matchLine := br.findForwardMatch(startLine, mapSize, mapSize); matchLine >= 0 {
			return matchLine, false
		}

		if startLine > 0 {
			if matchLine := br.findForwardMatch(0, minimum(startLine, mapSize), mapSize); matchLine >= 0 {
				return matchLine, true
			}
		}

		return -1, false
	}

	if matchLine := br.findReverseMatch(startLine, 0, mapSize); matchLine >= 0 {
		return matchLine, false
	}

	if startLine < mapSize-1 {
		if matchLine := br.findReverseMatch(mapSize-1, maximum(startLine+1, 0), mapSize); matchLine >= 0 {
			return matchLine, true
		}
	}

	return -1, false
}

// searchStartLine returns the first line to inspect for this search action.
func (br *browseObj) searchStartLine(forward, next bool, mapSize int) int {
	if !next {
		return br.currentPageSearchStart(forward, mapSize)
	}

	if br.lastMatch == SEARCH_RESET && !br.currentPageHasMatch(mapSize) {
		return br.currentPageSearchStart(forward, mapSize)
	}

	if forward {
		return br.firstRow + br.dispRows
	}

	return br.firstRow - 1
}

// currentPageSearchStart returns the first current-page line to inspect.
func (br *browseObj) currentPageSearchStart(forward bool, mapSize int) int {
	if forward {
		return br.firstRow
	}

	return minimum(br.firstRow+br.dispRows-1, mapSize-1)
}

// currentPageHasMatch reports whether the visible page already shows a match.
func (br *browseObj) currentPageHasMatch(mapSize int) bool {
	pageEnd := minimum(br.firstRow+br.dispRows, mapSize)

	for lineNum := br.firstRow; lineNum < pageEnd; lineNum++ {
		if matchCount, _ := br.lineIsMatch(lineNum); matchCount > 0 {
			return true
		}
	}

	return false
}

// findForwardMatch scans page-sized chunks from startLine up to endLine.
func (br *browseObj) findForwardMatch(startLine, endLine, mapSize int) int {
	startLine = maximum(startLine, 0)
	endLine = minimum(endLine, mapSize)
	if startLine >= endLine {
		return -1
	}

	pageRows := maximum(br.dispRows, 1)

	for pageStart := startLine; pageStart < endLine; pageStart += pageRows {
		pageEnd := minimum(pageStart+pageRows, endLine)
		for lineNum := pageStart; lineNum < pageEnd; lineNum++ {
			if matchCount, _ := br.lineIsMatch(lineNum); matchCount > 0 {
				return lineNum
			}
		}
	}

	return -1
}

// findReverseMatch scans page-sized chunks from startLine down to endLine.
func (br *browseObj) findReverseMatch(startLine, endLine, mapSize int) int {
	startLine = minimum(startLine, mapSize-1)
	endLine = maximum(endLine, 0)
	if startLine < endLine {
		return -1
	}

	pageRows := maximum(br.dispRows, 1)

	for pageEnd := startLine + 1; pageEnd > endLine; {
		pageStart := maximum(pageEnd-pageRows, endLine)
		for lineNum := pageEnd - 1; lineNum >= pageStart; lineNum-- {
			if matchCount, _ := br.lineIsMatch(lineNum); matchCount > 0 {
				return lineNum
			}
		}
		pageEnd = pageStart
	}

	return -1
}

// searchDisplayTop positions the match with directional context.
func (br *browseObj) searchDisplayTop(matchLine int, forward bool) int {
	if forward {
		return matchLine - br.dispRows/6
	}

	return matchLine - (br.dispRows*5)/6
}

// lineIsMatch reports whether a line matches and returns its content.
func (br *browseObj) lineIsMatch(lineno int) (int, []byte) {
	if !br.lineInMap(lineno) {
		return 0, nil
	}

	lineContent := br.readFromMap(lineno)
	if br.re == nil {
		return 0, lineContent
	}

	if br.re.Match(lineContent) {
		return 1, lineContent
	}

	return 0, lineContent
}

// replaceMatch highlights matches in a line and formats it for display.
func (br *browseObj) replaceMatch(lineno int, input []byte) string {
	sol := max(br.shiftWidth, 0)

	// Slice safely
	var content []byte

	if sol < len(input) {
		content = input[sol:]
	} else {
		content = nil
	}

	if br.re == nil {
		return br.formatLine(lineno, string(content))
	}

	leftMatch, rightMatch := br.undisplayedMatches(input, sol)

	if len(content) == 0 {
		if leftMatch {
			boldLeftArrow := _VID_BOLD + _VID_GREEN_FG + "\u2190" + VIDOFF
			return br.formatLine(lineno, boldLeftArrow)
		}

		return br.formatLine(lineno, "")
	}

	var replaced []byte

	if leftMatch || rightMatch {
		replaced = br.re.ReplaceAll(content, []byte(br.replace+_VID_GREEN_FG))
		replaced = append([]byte(_VID_GREEN_FG), replaced...)
		replaced = append(replaced, []byte(VIDOFF)...)
	} else {
		replaced = br.re.ReplaceAll(content, []byte(br.replace))
	}

	return br.formatLine(lineno, string(replaced))
}

// formatLine formats a line with optional line numbers.
func (br *browseObj) formatLine(lineno int, content string) string {
	content = linkURLs(content)

	if br.modeNumbers {
		// dim attribute is optional in the ANSI spec
		return fmt.Sprintf("%s%6d%s %s", _VID_DIM, lineno, _VID_OFF, content)
	}

	return content
}

// doSearch prompts for a pattern and performs a search in the given direction.
func (br *browseObj) doSearch(oldDir, newDir bool) bool {
	moveCursor(br.dispRows, 1, true)

	pattern, cancelled := userSearchComp(newDir)
	br.shownMsg = true

	if cancelled {
		br.restoreLast()
		return oldDir
	}

	prevPattern := br.pattern
	if pattern == "" {
		pattern = prevPattern

		if pattern != "" {
			moveCursor(br.dispRows, 2, true)
			fmt.Print(pattern)
		}
	}

	if pattern == "" {
		br.printMessage("No search pattern", MSG_ORANGE)
		return oldDir
	}

	// Substitute '&' with previous pattern for continued searches
	if prevPattern != "" {
		pattern = subCommandChars(pattern, "&", prevPattern)
	}

	if oldDir != newDir {
		dir := "reverse"
		if newDir {
			dir = "forward"
		}
		br.timedMessage("Searching "+dir, MSG_GREEN)
		br.lastMatch = SEARCH_RESET
	}

	continueSearch := (oldDir == newDir && pattern == prevPattern)
	searchSucceeded := br.searchFile(pattern, newDir, continueSearch)
	if searchSucceeded || pattern == br.pattern {
		updateHistory(pattern, searchHistory)
	}

	return newDir
}

// reCompile compiles the regex and updates search state.
func (br *browseObj) reCompile(pattern string) (int, error) {
	if pattern == "" {
		if br.pattern == "" {
			return 0, nil
		}

		pattern = br.pattern
	}

	if len(pattern) > MAX_PATTERN_LENGTH {
		return 0, fmt.Errorf("pattern too long (max %d characters)", MAX_PATTERN_LENGTH)
	}

	if pattern == "" {
		return 0, nil
	}

	compilePattern := pattern

	if br.searchFixed {
		compilePattern = regexp.QuoteMeta(compilePattern)
	}

	if br.ignoreCase {
		compilePattern = "(?i)" + compilePattern
	}

	re, err := regexp.Compile(compilePattern)
	if err != nil {
		return 0, err
	}

	br.pattern = pattern
	br.re = re
	br.replace = fmt.Sprintf("%s%s%s", MSG_GREEN, "$0", VIDOFF)

	return len(pattern), nil
}

// undisplayedMatches reports whether matches exist outside the visible slice.
func (br *browseObj) undisplayedMatches(input []byte, sol int) (bool, bool) {
	if br.re == nil {
		return false, false
	}

	// Bounds check for sol parameter
	if sol < 0 {
		sol = 0
	}

	// Use FindAllIndex for efficiency
	matches := br.re.FindAllIndex(input, -1)
	if len(matches) == 0 {
		return false, false
	}

	displayWidth := br.dispWidth
	if br.modeNumbers {
		displayWidth -= NUMCOLWIDTH
	}

	// Additional safety: ensure displayWidth is positive
	if displayWidth <= 0 {
		return false, false
	}

	leftMatch, rightMatch := false, false

	for _, index := range matches {
		// Ensure index has at least 2 elements (start and end positions)
		if len(index) < 2 {
			continue
		}

		// Validate index bounds
		if index[0] < 0 || index[0] >= len(input) {
			continue
		}

		if !leftMatch && index[0] < sol {
			leftMatch = true
		}

		// Calculate right boundary with safety checks
		rightBoundary := index[1] - sol + 2
		if !rightMatch && rightBoundary > displayWidth {
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
