// rcfile.go
// write and read the .browserc session file
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"bufio"
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const RCFILENAME = ".browserc"

func writeRcFile(br *browseObj) bool {
	filePath := filepath.Join(os.Getenv("HOME"), RCFILENAME)

	// fileName
	absPath, _ := filepath.Abs(br.fileName)
	fileName := absPath + "\n"

	// data buffer to accumulate all the data
	var data bytes.Buffer

	data.WriteString(fileName)

	// firstRow
	data.WriteString(strconv.Itoa(br.firstRow) + "\n")

	// pattern
	data.WriteString(br.pattern + "\n")

	// marks
	for i := 1; i <= 9; i++ {
		data.WriteString(strconv.Itoa(br.marks[i]) + " ")
	}
	data.WriteString("\n")

	// write the buffer data to the file
	err := os.WriteFile(filePath, data.Bytes(), 0644)

	return err == nil
}

func readRcFile(br *browseObj) bool {
	filePath := path.Join(os.Getenv("HOME"), RCFILENAME)
	filePath = os.ExpandEnv(filePath)

	fp, err := os.Open(filePath)

	if err != nil {
		return false
	}

	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	if scanner.Scan() {
		br.fileName = strings.TrimSpace(scanner.Text())
	} else {
		return false
	}

	if scanner.Scan() {
		firstRow, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))

		if err != nil {
			return false
		}

		br.firstRow = firstRow
	} else {
		return false
	}

	if scanner.Scan() {
		br.pattern = strings.TrimSpace(scanner.Text())
	} else {
		return false
	}

	if scanner.Scan() {
		markStrings := strings.Fields(strings.TrimSpace(scanner.Text()))

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
	} else {
		return false
	}

	return true
}

// vim: set ts=4 sw=4 noet:
