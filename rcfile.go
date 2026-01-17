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

	rcFileName := filepath.Join(os.Getenv("HOME"), (RCDIRNAME + "/" + RCFILENAME))

	// abs fileName
	data.WriteString(br.absFileName + "\n")

	// firstRow
	data.WriteString(strconv.Itoa(br.firstRow) + "\n")

	// pattern
	if br.ignoreCase {
		data.WriteString("(?i)" + br.pattern + "\n")
	} else {
		data.WriteString(br.pattern + "\n")
	}

	// marks
	for mark := 1; mark <= 9; mark++ {
		data.WriteString(strconv.Itoa(br.marks[mark]) + " ")
	}
	data.WriteString("\n")

	// title
	data.WriteString(br.title + "\n")

	// save
	err := os.WriteFile(rcFileName, []byte(data.String()), 0644)

	return err == nil
}

func (br *browseObj) readRcFile() bool {
	rcFileName := path.Join(os.Getenv("HOME"), (RCDIRNAME + "/" + RCFILENAME))
	rcFileName = os.ExpandEnv(rcFileName)

	fp, err := os.Open(rcFileName)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	for i := 0; i < 5; i++ {
		if !scanner.Scan() {
			// partial read ok
			return true
		}

		line := strings.TrimRight(scanner.Text(), "\r\n")

		if !br.handleRcFileLine(i, line) {
			return false
		}
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
