// search.go
// search the file for a given regex
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
	"regexp"
	"time"
)

// %6d + one space
const NUMCOLWIDTH = 7

func (x *browseObj) searchFile(pattern string, searchDir, next bool) {
	var sop, eop int
	var wrapped, warned bool
	var err error

	// to suppress S1002
	var searchFwd = searchDir

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
	// save in case regexp.Compile fails
	x.pattern = pattern

	if err != nil {
		x.printMessage(fmt.Sprintf("%v", err))
		return
	}

	// save regexp.Compile source and replstr
	x.pattern = x.re.String()
	x.replstr = fmt.Sprintf("%s%s%s", VIDBOLDGREEN, "$0", VIDOFF)

	// where to start search

	if x.lastMatch == SEARCH_RESET {
		// new search
		sop = x.firstRow
		eop = sop + x.dispRows
		wrapped = false
	} else if next {
		sop, eop, wrapped = x.setNextPage(searchDir, x.firstRow)
	}

	warned = false

	for {
		matchLine := x.pageIsMatch(sop, eop)

		if wrapped {
			if warned {
				x.printMessage("Pattern not found: " + x.pattern)
				return
			}

			if searchFwd {
				x.printMessage("Resuming search from SOF")
			} else {
				x.printMessage("Resuming search from EOF")
			}

			time.Sleep(1500 * time.Millisecond)
			warned = true
		}

		if matchLine > 0 {
			// sets firstRow, lastRow
			x.printPage(sop)
			return
		}

		if matchLine == 0 {
			sop, eop, wrapped = x.setNextPage(searchDir, sop)
		}
	}
}

func (x *browseObj) pageIsMatch(sop, eop int) int {
	// check if this page has a regex match

	for lineno := sop; lineno < eop; lineno++ {
		matches, _ := x.lineIsMatch(lineno)

		if matches > 0 {
			return lineno
		}
	}

	return 0
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
	var searchFwd = searchDir

	if searchFwd {
		sop += x.dispRows

		if sop >= x.mapSiz {
			wrapped = true
			sop = 0
		}
	} else {
		sop -= x.dispRows

		if sop < 0 {
			sop = x.mapSiz - x.dispRows
			wrapped = true
			// +1 for EOF
			sop++

			if sop < 0 {
				sop = 0
			}
		}
	}

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

	if !x.modeNumbers || windowAtEOF(lineno, x.mapSiz) {
		// no line numbers
		output = line
	} else {
		// line numbers -- uses NUMCOLWIDTH columns
		output = fmt.Sprintf("%6d %s", lineno, line)
	}

	return output
}

func (x *browseObj) noSearchPattern() bool {
	return x.re == nil || len(x.re.String()) == 0
}

// vim: set ts=4 sw=4 noet:
