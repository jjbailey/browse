// rcfile.go
// write and read the .browserc session file
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

const RCFILENAME = ".browserc"

func writeRcFile(br *browseObj) bool {
	filePath := fmt.Sprintf("%s/%s", os.Getenv("HOME"), RCFILENAME)
	fp, err := os.Create(filePath)

	if err != nil {
		return false
	}

	// fileName
	absPath, _ := filepath.Abs(br.fileName)
	fileName := fmt.Sprintf("%s\n", absPath)
	fp.WriteString(fileName)

	// firstRow
	fp.WriteString(fmt.Sprintf("%d\n", br.firstRow))

	// pattern
	fp.WriteString(fmt.Sprintf("%s\n", br.pattern))

	// marks
	for i := 1; i <= 9; i++ {
		fp.WriteString(fmt.Sprintf("%d ", br.marks[i]))
	}
	fp.WriteString("\n")

	fp.Close()
	return true
}

func readRcFile(br *browseObj) bool {
	var lbuf string

	filePath := fmt.Sprintf("%s/%s", os.Getenv("HOME"), RCFILENAME)

	fp, err := os.Open(filePath)

	if err != nil {
		return false
	}

	r := bufio.NewReader(fp)

	// fileName
	lbuf, _ = r.ReadString('\n')
	br.fileName = lbuf[:len(lbuf)-1]

	// firstRow
	lbuf, _ = r.ReadString('\n')
	fmt.Sscanf(lbuf, "%d", &br.firstRow)

	// pattern
	lbuf, _ = r.ReadString('\n')
	br.pattern = lbuf[:len(lbuf)-1]

	// marks
	lbuf, _ = r.ReadString('\n')
	fmt.Sscanf(lbuf, "%d %d %d %d %d %d %d %d %d",
		&br.marks[1], &br.marks[2], &br.marks[3],
		&br.marks[4], &br.marks[5], &br.marks[6],
		&br.marks[7], &br.marks[8], &br.marks[9])

	fp.Close()
	return true
}

// vim: set ts=4 sw=4 noet:
