// page.go
// paging and some support functions
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"strings"
)

func (x *browseObj) pageUp() {
	x.printPage(x.firstRow - x.dispRows)
}

func (x *browseObj) pageCurrent() {
	x.printPage(x.firstRow)
}

func (x *browseObj) pageDown() {
	x.printPage(x.firstRow + x.dispRows)
}

func (x *browseObj) pageHeader() {
	// print the header line

	// if title is too long, fit to size, include elipsis
	dispTitle := x.title
	lenDiff := (len(x.title) - x.dispWidth) + 6 + 3

	if lenDiff > 0 {
		dispTitle = "..." + x.title[lenDiff:]
	}

	// minus tees and spaces
	lineLen := x.dispWidth - 4
	oneside := int((lineLen - len(dispTitle)) / 2)

	resetScrRegion()
	moveCursor(1, 1, true)
	fmt.Print(CLEARSCREEN)
	fmt.Print(LINEWRAPOFF)
	fmt.Print(ENTERGRAPHICS)
	fmt.Print(strings.Repeat(HORIZLINE, oneside))
	fmt.Printf("%s%s", LEFTTEE, EXITGRAPHICS)
	fmt.Printf("%s %s %s", VIDBOLDREV, dispTitle, VIDOFF)
	fmt.Printf("%s%s", ENTERGRAPHICS, RIGHTTEE)
	fmt.Print(strings.Repeat(HORIZLINE, oneside+1))
	fmt.Print(EXITGRAPHICS)
	setScrRegion(2, x.dispHeight)
}

func (x *browseObj) pageLast() {
	x.printPage(x.mapSiz)
}

func (x *browseObj) pageMarked(lineno int) {
	x.printPage(x.marks[lineno])
}

// vim: set ts=4 sw=4 noet:
