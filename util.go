// util.go
// various uncategorized functions
//
// Copyright (c) 2024-2026 jjb
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
	"slices"
	"strings"
	"sync"
)

var tabBufPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

// expandTabs replaces tabs and carriage returns with spaces.
func expandTabs(data []byte) []byte {
	if !bytes.ContainsAny(data, "\t\r") {
		return data
	}

	buf := tabBufPool.Get().(*bytes.Buffer)
	buf.Reset()

	tabCount := bytes.Count(data, []byte{'\t'})
	buf.Grow(len(data) + tabCount*(TABWIDTH-1))

	for _, b := range data {
		switch b {

		case '\r':
			buf.WriteByte(' ')

		case '\t':
			spaces := TABWIDTH - (buf.Len() % TABWIDTH)
			for i := 0; i < spaces; i++ {
				buf.WriteByte(' ')
			}

		default:
			buf.WriteByte(b)
		}
	}

	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	tabBufPool.Put(buf)
	return result
}

// moveCursor positions the cursor and optionally clears the line.
func moveCursor(row, col int, clrflag bool) {
	if clrflag {
		fmt.Printf(CURPOS+CLEARLINE, row, col)
		return
	}

	fmt.Printf(CURPOS, row, col)
}

// printSEOF prints SOF/EOF markers on the display.
func printSEOF(what string) {
	if what == "EOF" {
		// save for modeScroll
		fmt.Printf("\r%s%s %s%s%s\r", CLEARSCREEN, CURSAVE, VIDBLINK, what, VIDOFF)
		return
	}

	fmt.Printf("\r %s%s%s\r", VIDBLINK, what, VIDOFF)
}

// windowAtEOF reports whether a line index is at EOF.
func windowAtEOF(lineno, mapsiz int) bool {
	return lineno == mapsiz
}

// maximum returns the larger of two integers.
func maximum(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// minimum returns the smaller of two integers.
func minimum(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// resetScrRegion restores the terminal scroll region.
func resetScrRegion() {
	fmt.Print(CURSAVE + RESETREGION + CURRESTORE)
}

// errorExit prints an error and exits after restoring the terminal.
func errorExit(err error) {
	if err != nil {
		fmt.Println(err)
		ttyRestore()
		os.Exit(1)
	}
}

// isValidMark reports whether a rune is a valid mark digit.
func isValidMark(r rune) bool {
	return r >= '1' && r <= '9'
}

// getMark extracts a mark digit from a string.
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

// isBinaryFile reports whether a file appears to be binary.
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

	return slices.Contains(buffer[:bytesRead], 0)
}

// subCommandChars replaces unescaped occurrences of a character.
// negative lookbehind not supported in golang RE2 engine
// pattern := `(?<!\\)%`
func subCommandChars(input, char, repl string) string {
	pattern := `(^|[^\\])` + regexp.QuoteMeta(char)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return input
	}

	return re.ReplaceAllString(input, `${1}`+repl)
}

// resolveSymlink resolves symlinks and returns a clean path.
func resolveSymlink(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return path, err
	}

	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "broken symlink", err
	}

	return filepath.Clean(realPath), nil
}

// lastNChars returns the tail of a string limited by display width.
func lastNChars(s string, dispWidth int) string {
	const padding = 45

	usable := maximum(dispWidth-padding, dispWidth>>1)
	if usable < 0 {
		usable = 0
	}
	runes := []rune(s)
	runeLen := len(runes)
	size := minimum(usable, runeLen)

	return string(runes[len(runes)-size:])
}

// shellEscapeSingle safely single-quotes a string for the shell.
func shellEscapeSingle(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// prevDirectory returns the previous directory from history.
func prevDirectory() string {
	history := loadHistory(dirHistory)
	if len(history) < 2 {
		return ""
	}

	return history[len(history)-2]
}

// unQuote removes surrounding single quotes from a string.
func unQuote(s string) string {
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		s = s[1 : len(s)-1]
	}

	return s
}

// vim: set ts=4 sw=4 noet:
