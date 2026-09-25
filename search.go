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
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

// Search formatting and limits.
const (
	// Maximum regex pattern length to avoid accidental oversized searches.
	MAX_PATTERN_LENGTH = 1000

	// Bytes read per system call when scanning for a match.
	SEARCH_BLOCK_SIZE = 256 * 1024

	// Ctrl-C arrives as a byte while the terminal is in browse mode.
	SEARCH_CANCEL_KEY = 0x03
)

// errSearchCancelled reports that the user interrupted a search.
var errSearchCancelled = errors.New("search cancelled")

// searchFile scans for a regex pattern and updates the view accordingly.
// forward: true = forward, false = reverse
// next: true = continue search, false = new search
func (br *browseObj) searchFile(pattern string, forward, next bool) bool {
	if pattern == "" {
		br.printMessage("No search pattern", MSG_ORANGE)
		return false
	}

	// Reset search state after a new or missing regexp compiles successfully.
	if pattern != br.pattern || br.re == nil {
		if err := br.reCompile(pattern); err != nil {
			br.printMessage(fmt.Sprintf("Regex compilation error: %v", err), MSG_ORANGE)
			return false
		}

		br.lastMatch = SEARCH_RESET
		next = false
	}

	// A search that starts on the current page leaves the page in place when
	// the match is already visible.
	fromCurrentPage := !next || br.lastMatch == SEARCH_RESET

	matchLine, wrapped, err := br.findSearchMatch(forward, next)
	if err != nil {
		br.printMessage("Search cancelled", MSG_ORANGE)
		moveCursor(2, 1, false)
		return false
	}

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
	if fromCurrentPage && !wrapped && br.lineOnCurrentPage(matchLine) {
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

// displayWrapMessage informs the user when the search wraps.
func (br *browseObj) displayWrapMessage(forward bool) {
	if forward {
		br.timedMessage("Resuming search from SOF", MSG_GREEN)
	} else {
		br.timedMessage("Resuming search from EOF", MSG_GREEN)
	}
}

// findSearchMatch returns the next matching line without changing display
// state. The error is errSearchCancelled if the user interrupted the scan.
func (br *browseObj) findSearchMatch(forward, next bool) (int, bool, error) {
	mapSize := br.searchMapSize()
	if mapSize <= 0 {
		return -1, false, nil
	}

	startLine := br.searchStartLine(forward, next, mapSize)

	if forward {
		if matchLine, err := br.findForwardMatch(startLine, mapSize, mapSize); matchLine >= 0 || err != nil {
			return matchLine, false, err
		}

		if startLine > 0 {
			if matchLine, err := br.findForwardMatch(0, minimum(startLine, mapSize), mapSize); matchLine >= 0 || err != nil {
				return matchLine, true, err
			}
		}

		return -1, false, nil
	}

	if matchLine, err := br.findReverseMatch(startLine, 0, mapSize); matchLine >= 0 || err != nil {
		return matchLine, false, err
	}

	if startLine < mapSize-1 {
		if matchLine, err := br.findReverseMatch(mapSize-1, maximum(startLine+1, 0), mapSize); matchLine >= 0 || err != nil {
			return matchLine, true, err
		}
	}

	return -1, false, nil
}

// searchStartLine returns the first line to inspect for this search action.
func (br *browseObj) searchStartLine(forward, next bool, mapSize int) int {
	if !next || br.lastMatch == SEARCH_RESET {
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
		return maximum(br.firstRow, 1)
	}

	return minimum(br.firstRow+br.dispRows-1, mapSize-1)
}

// findForwardMatch scans from startLine up to endLine for the first match.
func (br *browseObj) findForwardMatch(startLine, endLine, mapSize int) (int, error) {
	// Line zero is the synthetic SOF marker, not file content.
	startLine = maximum(startLine, 1)
	endLine = minimum(endLine, mapSize)

	for first := startLine; first < endLine; {
		if br.searchCancelled() {
			return -1, errSearchCancelled
		}

		_, lines := br.loadSearchBlock(first, endLine, true)
		if len(lines) == 0 {
			break
		}

		for i, line := range lines {
			if br.searchLineMatches(line) {
				return first + i, nil
			}
		}

		first += len(lines)
	}

	return -1, nil
}

// findReverseMatch scans from startLine down to endLine for the first match.
func (br *browseObj) findReverseMatch(startLine, endLine, mapSize int) (int, error) {
	startLine = minimum(startLine, mapSize-1)
	// Line zero is the synthetic SOF marker, not file content.
	endLine = maximum(endLine, 1)

	for last := startLine; last >= endLine; {
		if br.searchCancelled() {
			return -1, errSearchCancelled
		}

		first, lines := br.loadSearchBlock(last, endLine, false)
		if len(lines) == 0 {
			break
		}

		for i := len(lines) - 1; i >= 0; i-- {
			if br.searchLineMatches(lines[i]) {
				return first + i, nil
			}
		}

		last = first - 1
	}

	return -1, nil
}

// loadSearchBlock reads a run of consecutive lines with one system call. A
// forward run starts at line and extends toward limit (exclusive); a reverse
// run ends at line and extends back toward limit (inclusive). It returns the
// run's first line number and each line's bytes in ascending order. A nil
// entry marks a line that could not be read, which never matches; a readable
// empty line is non-nil. The slices alias a per-session buffer that is only
// valid until the next call, so only the main goroutine may use it.
func (br *browseObj) loadSearchBlock(line, limit int, forward bool) (int, [][]byte) {
	br.mutex.Lock()
	defer br.mutex.Unlock()

	if line < 1 || line >= br.mapSiz || br.fp == nil {
		return line, nil
	}

	br.searchLines = br.searchLines[:0]

	// Reader metadata is capped at READBUFSIZ; reject corrupted metadata before
	// using it as a slice length or allocation size.
	validSize := func(n int) bool {
		return br.sizeMap[n] >= 0 && br.sizeMap[n] <= READBUFSIZ
	}
	if !validSize(line) {
		return line, append(br.searchLines, nil)
	}

	// Grow the run while it stays within one block. Lines ascend through the
	// file without overlap; stop at anything else rather than read garbage.
	first, last := line, line
	start := br.seekMap[line]
	end := start + br.sizeMap[line]

	if forward {
		for next := last + 1; next < limit && next < br.mapSiz && validSize(next); next++ {
			nextEnd := br.seekMap[next] + br.sizeMap[next]
			if br.seekMap[next] < end || nextEnd-start > SEARCH_BLOCK_SIZE {
				break
			}
			last, end = next, nextEnd
		}
	} else {
		for prev := first - 1; prev >= limit && prev >= 1 && validSize(prev); prev-- {
			prevEnd := br.seekMap[prev] + br.sizeMap[prev]
			if prevEnd > start || end-br.seekMap[prev] > SEARCH_BLOCK_SIZE {
				break
			}
			first, start = prev, br.seekMap[prev]
		}
	}

	// Allocate a full block up front so the buffer is never nil, which keeps
	// empty lines distinguishable from unreadable ones.
	if cap(br.searchBuf) < SEARCH_BLOCK_SIZE {
		br.searchBuf = make([]byte, SEARCH_BLOCK_SIZE)
	}
	buf := br.searchBuf[:end-start]

	n, err := br.fp.ReadAt(buf, start)
	if err != nil && err != io.EOF {
		// Skip the run; the scan continues with the lines after it.
		for range last - first + 1 {
			br.searchLines = append(br.searchLines, nil)
		}
		return first, br.searchLines
	}

	// A short read (file truncated underneath us) leaves later lines partial.
	for i := first; i <= last; i++ {
		lineStart := min(int(br.seekMap[i]-start), n)
		lineEnd := min(lineStart+int(br.sizeMap[i]), n)
		br.searchLines = append(br.searchLines, buf[lineStart:lineEnd])
	}

	return first, br.searchLines
}

// searchLineMatches reports whether a raw line matches the active pattern.
// Matching runs on the tab-expanded line, as displayed.
func (br *browseObj) searchLineMatches(line []byte) bool {
	if line == nil || br.re == nil {
		return false
	}

	// expandTabs returns line unchanged when there are no tabs, so the
	// common case stays allocation-free.
	expanded, scratch := expandTabs(line, br.matchTabScratch)
	br.matchTabScratch = scratch

	if br.matchLiteral == nil {
		return br.re.Match(expanded)
	}

	if br.matchFold {
		folded, scratch := foldASCII(expanded, br.matchFoldScratch)
		br.matchFoldScratch = scratch
		return bytes.Contains(folded, br.matchLiteral)
	}

	return bytes.Contains(expanded, br.matchLiteral)
}

// foldASCII lowercases ASCII letters for case-insensitive literal matching.
// The Kelvin sign and long s are the only non-ASCII runes that case-fold to
// ASCII letters, so they are mapped to k and s; with an ASCII pattern this
// gives the same result as the regexp (?i) flag. The output is for Contains
// only and does not preserve byte offsets. dst is reused as in expandTabs.
func foldASCII(data, dst []byte) ([]byte, []byte) {
	out := dst[:0]

	for i := 0; i < len(data); i++ {
		b := data[i]
		switch {

		case b >= 'A' && b <= 'Z':
			out = append(out, b+('a'-'A'))

		case b == 0xE2 && i+2 < len(data) && data[i+1] == 0x84 && data[i+2] == 0xAA:
			// U+212A KELVIN SIGN
			out = append(out, 'k')
			i += 2

		case b == 0xC5 && i+1 < len(data) && data[i+1] == 0xBF:
			// U+017F LATIN SMALL LETTER LONG S
			out = append(out, 's')
			i++

		default:
			out = append(out, b)
		}
	}

	return out, out[:0]
}

// searchCancelled reports whether the user pressed Ctrl-C during a search.
// It never blocks. Other keys typed during the search are kept for the
// command loop rather than dropped.
func (br *browseObj) searchCancelled() bool {
	if br.tty == nil {
		return false
	}

	pollFds := []unix.PollFd{{Fd: int32(br.tty.Fd()), Events: unix.POLLIN}}
	if ready, err := unix.Poll(pollFds, 0); err != nil || ready == 0 {
		return false
	}

	buf := make([]byte, 16)
	n, err := br.tty.Read(buf)
	if err != nil || n == 0 {
		return false
	}

	if bytes.IndexByte(buf[:n], SEARCH_CANCEL_KEY) >= 0 {
		return true
	}

	br.pendingInput = append(br.pendingInput, buf[:n]...)
	return false
}

// searchDisplayTop positions the match with directional context.
func (br *browseObj) searchDisplayTop(matchLine int, forward bool) int {
	if forward {
		return matchLine - br.dispRows/6
	}

	return matchLine - (br.dispRows*5)/6
}

// replaceMatch highlights matches in a line and formats it for display.
func (br *browseObj) replaceMatch(lineno int, input []byte) string {
	sol := max(br.shiftWidth, 0)

	// Slice safely
	var content []byte

	if sol < len(input) {
		content = input[sol:]
	}

	if br.re == nil {
		return br.formatLine(lineno, linkURLs(string(content)))
	}

	// Match against the whole line, not the shifted slice, so anchors and
	// word boundaries keep their meaning after a horizontal shift.
	matches := br.re.FindAllIndex(input, -1)
	leftMatch, rightMatch := undisplayedMatches(matches, sol, br.rightLimit(input, sol))

	if len(content) == 0 {
		if leftMatch {
			boldLeftArrow := _VID_BOLD + _VID_GREEN_FG + "\u2190" + VIDOFF
			return br.formatLine(lineno, boldLeftArrow)
		}

		return br.formatLine(lineno, "")
	}

	return br.formatLine(lineno, highlightLine(content, matches, sol, leftMatch || rightMatch))
}

// rightLimit returns the byte offset in input just past the last column
// visible when the line is shifted to sol.
func (br *browseObj) rightLimit(input []byte, sol int) int {
	displayWidth := br.dispWidth
	if br.modeNumbers {
		displayWidth -= NUMCOLWIDTH
	}

	// Two columns are reserved, matching the right-edge marker threshold.
	limit := sol
	for cols := 0; cols < displayWidth-2 && limit < len(input); cols++ {
		_, width := utf8.DecodeRune(input[limit:])
		limit += width
	}

	return limit
}

// textSpan is a half-open byte range [start, end).
type textSpan struct {
	start, end int
}

// highlightLine renders content with search matches highlighted and URLs
// wrapped as terminal hyperlinks. matches are index pairs from the full line,
// which begins sol bytes before content. When offscreen is set the line is
// tinted to flag matches outside the visible slice.
func highlightLine(content []byte, matches [][]int, sol int, offscreen bool) string {
	var hl []textSpan
	for _, m := range matches {
		// Empty matches have nothing to highlight.
		if m[1] <= sol || m[0] == m[1] {
			continue
		}

		hl = append(hl, textSpan{max(m[0], sol) - sol, m[1] - sol})
	}

	// Find URLs in the raw text; searching after highlighting would stop
	// each URL at the first highlight escape.
	var urls []textSpan
	if bytes.Contains(content, []byte("://")) {
		for _, u := range urlRe.FindAllIndex(content, -1) {
			urls = append(urls, textSpan{u[0], u[1]})
		}
	}

	base := ""
	if offscreen {
		base = _VID_GREEN_FG
	}

	var sb strings.Builder
	sb.Grow(len(content) + len(hl)*(len(MSG_GREEN)+len(VIDOFF)+len(base)) + 16)
	sb.WriteString(base)

	// Merge the two sorted span lists. SGR and OSC 8 state are independent,
	// so highlights and links may overlap freely.
	pos, h, u := 0, 0, 0
	inHL, inURL := false, false

	for {
		if inHL && hl[h].end == pos {
			sb.WriteString(VIDOFF + base)
			inHL = false
			h++
		}
		if inURL && urls[u].end == pos {
			sb.WriteString(oscClose)
			inURL = false
			u++
		}
		if !inURL && u < len(urls) && urls[u].start == pos {
			sb.WriteString(oscOpen(string(content[urls[u].start:urls[u].end])))
			inURL = true
		}
		if !inHL && h < len(hl) && hl[h].start == pos {
			sb.WriteString(MSG_GREEN)
			inHL = true
		}

		if pos == len(content) {
			break
		}

		next := len(content)
		if h < len(hl) {
			if inHL {
				next = min(next, hl[h].end)
			} else {
				next = min(next, hl[h].start)
			}
		}
		if u < len(urls) {
			if inURL {
				next = min(next, urls[u].end)
			} else {
				next = min(next, urls[u].start)
			}
		}

		sb.Write(content[pos:next])
		pos = next
	}

	if offscreen {
		sb.WriteString(VIDOFF)
	}

	return sb.String()
}

// formatLine formats a line with optional line numbers.
func (br *browseObj) formatLine(lineno int, content string) string {
	if br.modeNumbers {
		// dim attribute is optional in the ANSI spec
		num := strconv.Itoa(lineno)

		var sb strings.Builder
		sb.Grow(len(_VID_DIM) + len(_VID_OFF) + NUMCOLWIDTH + len(content))
		sb.WriteString(_VID_DIM)
		for i := len(num); i < NUMCOLWIDTH-1; i++ {
			sb.WriteByte(' ')
		}
		sb.WriteString(num)
		sb.WriteString(_VID_OFF)
		sb.WriteByte(' ')
		sb.WriteString(content)
		return sb.String()
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
	typed := pattern
	if pattern == "" {
		pattern = prevPattern

		if pattern != "" {
			moveCursor(br.dispRows, 2, true)
			fmt.Print(pattern + "\n")
		}
	}

	if pattern == "" {
		br.printMessage("No search pattern", MSG_ORANGE)
		return oldDir
	}

	// Expand '&' only in what the user typed; a reused pattern is already
	// expanded and may contain a literal '&'.
	if typed != "" {
		pattern = expandSearchSymbol(typed, prevPattern, br.searchFixed)
	}

	if oldDir != newDir {
		dir := "reverse"
		if newDir {
			dir = "forward"
		}
		br.timedMessage("Searching "+dir, MSG_GREEN)
	}

	continueSearch := (oldDir == newDir && pattern == prevPattern)
	searchSucceeded := br.searchFile(pattern, newDir, continueSearch)
	if searchSucceeded || pattern == br.pattern {
		updateHistory(pattern, searchHistory)
	}

	return newDir
}

// expandSearchSymbol replaces unescaped '&' in a typed pattern with the
// previous pattern. With no previous pattern '&' stays literal. In regex mode
// backslashes pass through untouched, so '\&' remains a regex-escaped
// literal and '\\&' is an escaped backslash followed by the expansion. In
// fixed-string mode backslashes are not regex syntax: a backslash run before
// '&' collapses by shell rules, so '\&' yields a literal '&'.
func expandSearchSymbol(input, prev string, fixed bool) string {
	var sb strings.Builder
	slashRun := 0

	for _, r := range input {
		switch {

		case r == '\\':
			slashRun++
			continue

		case r == '&':
			slashes := slashRun
			if fixed {
				slashes = slashRun / 2
			}
			sb.WriteString(strings.Repeat(`\`, slashes))

			if slashRun%2 == 0 && prev != "" {
				sb.WriteString(prev)
			} else {
				sb.WriteRune(r)
			}

		default:
			sb.WriteString(strings.Repeat(`\`, slashRun))
			sb.WriteRune(r)
		}

		slashRun = 0
	}

	sb.WriteString(strings.Repeat(`\`, slashRun))
	return sb.String()
}

// reCompile compiles the regex and updates search state. An empty pattern
// leaves the state unchanged.
func (br *browseObj) reCompile(pattern string) error {
	if pattern == "" {
		return nil
	}

	if len(pattern) > MAX_PATTERN_LENGTH {
		return fmt.Errorf("pattern too long (max %d characters)", MAX_PATTERN_LENGTH)
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
		return err
	}

	br.pattern = pattern
	br.re = re
	br.setMatchLiteral(pattern)

	return nil
}

// setMatchLiteral enables a faster byte search when the pattern has no regex
// syntax. Case-insensitive literals take the fast path only when ASCII, since
// foldASCII does not implement full Unicode case folding.
func (br *browseObj) setMatchLiteral(pattern string) {
	br.matchLiteral, br.matchFold = nil, false

	if !br.searchFixed && regexp.QuoteMeta(pattern) != pattern {
		return
	}

	if !br.ignoreCase {
		br.matchLiteral = []byte(pattern)
		return
	}

	for i := 0; i < len(pattern); i++ {
		if pattern[i] >= utf8.RuneSelf {
			return
		}
	}

	br.matchLiteral, _ = foldASCII([]byte(pattern), nil)
	br.matchFold = true
}

// toggleSearchOption flips a search option and recompiles the active pattern.
// If the pattern does not compile under the new setting, the option and the
// compiled regexp are left as they were.
func (br *browseObj) toggleSearchOption(option *bool) error {
	*option = !*option

	if err := br.reCompile(br.pattern); err != nil {
		*option = !*option
		return err
	}

	br.lastMatch = SEARCH_RESET
	return nil
}

// clearSearchRegex drops the compiled pattern.
func (br *browseObj) clearSearchRegex() {
	br.re = nil
	br.matchLiteral, br.matchFold = nil, false
}

// undisplayedMatches reports whether any non-empty match begins left of the
// visible slice (sol) or ends past its right edge (rightLimit). matches must
// be in left-to-right order, as FindAllIndex returns them.
func undisplayedMatches(matches [][]int, sol, rightLimit int) (bool, bool) {
	leftMatch := false

	for _, m := range matches {
		if m[0] == m[1] {
			continue
		}
		if m[0] < sol {
			leftMatch = true
		}
		// Later matches start further right, so the scan can stop here.
		if m[1] > rightLimit {
			return leftMatch, true
		}
	}

	return leftMatch, false
}

// vim: set ts=4 sw=4 noet:
