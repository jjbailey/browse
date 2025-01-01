// rcfile.go
// write and read the .browserc session file
//
// Copyright (c) 2024 jjb
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

const RCFILENAME = ".browserc"

func writeRcFile(br *browseObj) bool {
	var data strings.Builder

	filePath := filepath.Join(os.Getenv("HOME"), RCFILENAME)

	// fileName
	absPath, _ := filepath.Abs(br.fileName)
	data.WriteString(absPath + "\n")

	// firstRow
	data.WriteString(strconv.Itoa(br.firstRow) + "\n")

	// pattern
	data.WriteString(br.pattern + "\n")

	// marks
	for mark := 1; mark <= 9; mark++ {
		data.WriteString(strconv.Itoa(br.marks[mark]) + " ")
	}
	data.WriteString("\n")

	// title
	data.WriteString(br.title + "\n")

	// save
	err := os.WriteFile(filePath, []byte(data.String()), 0644)

	return err == nil
}

func readRcFile(br *browseObj) bool {
	filePath := path.Join(os.Getenv("HOME"), RCFILENAME)
	filePath = os.ExpandEnv(filePath)

	fp, err := os.Open(filePath)

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

		line := strings.TrimSpace(scanner.Text())

		switch i {

		case 0:
			// fileName
			br.fileName = line

		case 1:
			// firstRow
			if firstRow, err := strconv.Atoi(line); err != nil {
				return false
			} else {
				br.firstRow = firstRow
			}

		case 2:
			// pattern
			br.pattern = line

		case 3:
			// marks
			markStrings := strings.Fields(line)

			if len(markStrings) != 9 {
				return false
			}

			for i, markString := range markStrings {
				if mark, err := strconv.Atoi(markString); err != nil {
					return false
				} else {
					br.marks[i+1] = mark
				}
			}

		case 4:
			// title
			br.title = line
		}
	}

	return true
}

// vim: set ts=4 sw=4 noet:
