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

func (x *browseObj) searchFile(pattern string, searchFWD bool) {
	var sop, eop int
	var wrapped, warned bool

	if len(pattern) == 0 {
		if len(x.pattern) == 0 {
			x.printMessage("No search pattern")
			return
		}

		pattern = x.pattern
	}

	re, err := regexp.Compile(pattern)
	// save in case regexp.Compile fails
	x.pattern = pattern

	if err != nil {
		x.printMessage(fmt.Sprintf("%v", err))
		return
	}

	// save regexp.Compile source
	x.pattern = re.String()

	// initial settings
	replstr := fmt.Sprintf("%s%s%s", VIDBOLDGREEN, "$0", VIDOFF)
	sop = x.firstRow
	eop = sop + x.dispRows
	wrapped = false
	warned = false

	for {
		matchLine := x.pageIsMatch(re, sop, eop)
		searchThisPage := (x.lastMatch == RESETSRCH || wrapped) && (matchLine > 0)

		if !searchThisPage {
			sop, eop, wrapped = x.setNextPage(searchFWD, sop)
			matchLine = x.pageIsMatch(re, sop, eop)
		}

		if wrapped {
			if warned {
				x.printMessage("Pattern not found: " + x.pattern)
				return
			}

			time.Sleep(750 * time.Millisecond)

			if searchFWD {
				x.printMessage("Resuming search from SOF")
			} else {
				x.printMessage("Resuming search from EOF")
			}

			time.Sleep(1500 * time.Millisecond)
			warned = true
		}

		if matchLine > 0 {
			x.printPage(sop) // sets firstRow, lastRow
			// reset
			sop = x.firstRow
			eop = x.lastRow
			curRow := 1
			foundMatch := false

			for i := sop; i < eop; i++ {
				matches, input := x.lineIsMatch(re, i)
				curRow++

				if matches == 0 {
					continue
				}

				output := x.replaceMatch(re, i, input, replstr)
				n := (len(VIDBOLDGREEN)+len(VIDOFF))*matches + x.dispWidth
				movecursor(curRow, 1, false)
				fmt.Printf("%.*s%s%s", n, output, VIDOFF, CLEARLINE)
				x.lastMatch = i
				foundMatch = true
			}

			if foundMatch {
				movecursor(2, 1, false)
				return
			}
		}
	}
}

func (x *browseObj) pageIsMatch(re *regexp.Regexp, sop, eop int) int {
	// check if this page has a regex match

	for i := sop; i < eop; i++ {
		matches, _ := x.lineIsMatch(re, i)

		if matches > 0 {
			return i
		}
	}

	return 0
}

func (x *browseObj) lineIsMatch(re *regexp.Regexp, lineno int) (int, string) {
	// check if this line has a regex match

	var n int

	data, nbytes := x.readFromMap(lineno)

	// match on what's visible

	if !x.modeNumbers {
		n = minimum(nbytes, x.dispWidth)
	} else {
		// line numbers -- uses 7 columns
		n = minimum(nbytes, x.dispWidth-7)
	}

	input := string(data[0:n])
	return len(re.FindAllString(input, -1)), input
}

func (x *browseObj) setNextPage(searchFWD bool, sop int) (int, int, bool) {
	// figure out which page to search next

	var eop int
	var wrapped bool

	if searchFWD {
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

func (x *browseObj) replaceMatch(re *regexp.Regexp, lineno int, input, replstr string) string {
	// make the regex replacements, return the new line

	var output string

	line := re.ReplaceAllString(input, replstr)

	if !x.modeNumbers {
		// no line numbers
		output = line
	} else {
		// 7 columns for line numbers
		output = fmt.Sprintf("%6d %s", lineno, line)
	}

	return output
}

// vim: set ts=4 sw=4 noet:
