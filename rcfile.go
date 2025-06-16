// rcfile.go
// write and read the .browserc session file
//
// Copyright (c) 2024-2025 jjb
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

func (br *browseObj) writeRcFile() bool {
	var absPath string
	var data strings.Builder
	var err error

	rcFileName := filepath.Join(os.Getenv("HOME"), RCFILENAME)
	absPath = resolveSymlink(br.fileName)

	// fileName
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
	err = os.WriteFile(rcFileName, []byte(data.String()), 0644)

	return err == nil
}

func (br *browseObj) readRcFile() bool {
	rcFileName := path.Join(os.Getenv("HOME"), RCFILENAME)
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

		case 4:
			// title
			br.title = line
		}
	}

	return true
}

func resolveSymlink(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil || len(absPath) == 0 {
		return path
	}

	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil || len(realPath) == 0 {
		return absPath
	}

	return realPath
}

// vim: set ts=4 sw=4 noet:
