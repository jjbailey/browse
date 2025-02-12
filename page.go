// page.go
// paging and some support functions
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"strings"
)

func (br *browseObj) pageUp() {
	br.printPage(br.firstRow - br.dispRows)
}

func (br *browseObj) pageCurrent() {
	br.printPage(br.firstRow)
}

func (br *browseObj) pageDown() {
	br.printPage(br.firstRow + br.dispRows)
}

func (br *browseObj) pageHeader() {
	// print the header line

	// if the title is too long, fit to size, include ellipsis
	dispTitle := br.title
	lenDiff := (len(br.title) - br.dispWidth) + 6 + 3

	if lenDiff > 0 {
		dispTitle = "..." + br.title[lenDiff:]
	}

	// minus tees and spaces
	lineLen := br.dispWidth - 4
	oneSide := (lineLen - len(dispTitle)) >> 1

	// -----| title |-----
	header := fmt.Sprintf("%s%s%s%s%s %s %s%s%s%s%s",
		ENTERGRAPHICS, strings.Repeat(HORIZLINE, oneSide), LEFTTEE, EXITGRAPHICS,
		VIDBOLDREV, dispTitle, VIDOFF,
		ENTERGRAPHICS, RIGHTTEE, strings.Repeat(HORIZLINE, oneSide+1), EXITGRAPHICS)

	resetScrRegion()
	moveCursor(1, 1, true)
	fmt.Print(CLEARSCREEN)
	fmt.Print(LINEWRAPOFF)
	fmt.Print(header)
	setScrRegion(2, br.dispHeight)
}

func (br *browseObj) pageLast() {
	br.printPage(br.mapSiz)
}

func (br *browseObj) pageMarked(lineno int) {
	br.printPage(br.marks[lineno])
}

// vim: set ts=4 sw=4 noet:
