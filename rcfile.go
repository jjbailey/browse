// rcfile.go
// write and read the browserc session file
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func (br *browseObj) writeRcFile() bool {
	var data strings.Builder

	rcFileName := filepath.Join(os.Getenv("HOME"), RCDIRNAME, RCFILENAME)

	// abs fileName
	data.WriteString(br.absFileName)
	data.WriteByte('\n')

	// firstRow
	data.WriteString(strconv.Itoa(br.firstRow))
	data.WriteByte('\n')

	// pattern
	data.WriteString(br.pattern)
	data.WriteByte('\n')

	// marks
	for mark := 1; mark <= 9; mark++ {
		data.WriteString(strconv.Itoa(br.marks[mark]))
		data.WriteByte(' ')
	}
	data.WriteByte('\n')

	// title
	data.WriteString(br.title)
	data.WriteByte('\n')

	// ignoreCase
	data.WriteString(strconv.FormatBool(br.ignoreCase))
	data.WriteByte('\n')

	// save
	err := os.WriteFile(rcFileName, []byte(data.String()), 0644)

	return err == nil
}

func (br *browseObj) readRcFile() bool {
	rcFileName := path.Join(os.Getenv("HOME"), RCDIRNAME, RCFILENAME)
	rcFileName = os.ExpandEnv(rcFileName)

	fp, err := os.Open(rcFileName)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	linesRead := 0
	for i := range 6 {
		if !scanner.Scan() {
			break
		}

		linesRead++
		line := strings.TrimRight(scanner.Text(), "\r\n")

		if !br.handleRcFileLine(i, line) {
			return false
		}
	}

	if err := scanner.Err(); err != nil {
		return false
	}

	if linesRead < 6 && strings.HasPrefix(br.pattern, "(?i)") {
		br.pattern = strings.TrimPrefix(br.pattern, "(?i)")
		br.ignoreCase = true
	}

	return true
}

func (br *browseObj) handleRcFileLine(i int, line string) bool {
	switch i {

	case 0:
		// fileName
		br.fileName = line

	case 1:
		// firstRow
		firstRow, err := strconv.Atoi(line)
		if err != nil {
			return false
		}
		br.firstRow = firstRow

	case 2:
		// pattern
		br.pattern = line

	case 3:
		// marks
		return br.parseMarks(line)

	case 4:
		// title
		br.title = line

	case 5:
		// ignoreCase
		ignoreCase, err := strconv.ParseBool(line)
		if err != nil {
			return false
		}
		br.ignoreCase = ignoreCase
	}

	return true
}

func (br *browseObj) parseMarks(line string) bool {
	markStrings := strings.Fields(line)
	if len(markStrings) != 9 {
		return false
	}

	for i, markString := range markStrings {
		mark, err := strconv.Atoi(markString)
		if err != nil {
			return false
		}
		br.marks[i+1] = mark
	}

	return true
}

// vim: set ts=4 sw=4 noet:
