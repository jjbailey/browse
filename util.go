// util.go
// various uncategorized functions
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func expandTabs(data []byte) []byte {
	if !bytes.ContainsAny(data, "\t\r") {
		return data
	}

	tabCount := bytes.Count(data, []byte{'\t'})
	capacity := len(data) + tabCount*(TABWIDTH-1)

	buf := make([]byte, 0, capacity)

	for _, b := range data {
		switch b {

		case '\r':
			// silently map CR to space
			buf = append(buf, ' ')

		case '\t':
			tab := bytes.Repeat([]byte{' '}, TABWIDTH-len(buf)%TABWIDTH)
			buf = append(buf, tab...)

		default:
			buf = append(buf, b)
		}
	}

	return buf
}

func moveCursor(row, col int, clrflag bool) {
	fmt.Printf(CURPOS, row, col)

	if clrflag {
		fmt.Print(CLEARLINE)
	}
}

func printSEOF(what string) {
	if what == "EOF" {
		// save for modeScroll
		fmt.Printf("\r%s%s", CLEARSCREEN, CURSAVE)
	}

	fmt.Printf("\r %s%s%s\r", VIDBLINK, what, VIDOFF)
}

func windowAtEOF(lineno, mapsiz int) bool {
	return lineno == mapsiz
}

func maximum(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func setScrRegion(top, bot int) {
	fmt.Printf(SCROLLREGION, top, bot)
}

func resetScrRegion() {
	fmt.Print(RESETREGION)
}

func errorExit(err error) {
	if err != nil {
		fmt.Println(err)
		ttyRestore()
		os.Exit(1)
	}
}

func isValidMark(r rune) bool {
	return r >= '1' && r <= '9'
}

func getMark(buf string) int {
	if len(buf) == 0 {
		return 0
	}

	idx := strings.IndexFunc(buf, isValidMark)

	if idx == -1 {
		return 0
	}

	return int(buf[idx] - '0')
}

func isBinaryFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	const sampleSize = 4 * 1024
	buffer := make([]byte, sampleSize)

	bytesRead, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	for _, b := range buffer[:bytesRead] {
		if b == 0 {
			return true
		}
	}

	return false
}

func subCommandChars(input, char, repl string) string {
	// negative lookbehind not supported in golang RE2 engine
	// pattern := `(?<!\\)%`

	pattern := `(^|[^\\])` + regexp.QuoteMeta(char)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return input
	}

	return re.ReplaceAllString(input, `${1}`+repl)
}

func resolveSymlink(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "broken symlink"
	}

	return realPath
}

// vim: set ts=4 sw=4 noet:
