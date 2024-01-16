// page.go
// paging and some support functions
//
// Copyright (c) 2024 jjb
// All rights reserved.

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

	// minus tees and spaces
	linelen := x.dispWidth - 4
	oneside := (linelen - len(x.screenName)) / 2

	resetScrRegion()
	movecursor(1, 1, true)
	fmt.Printf("%s", CLEARSCREEN)
	fmt.Printf("%s", ENTERGRAPHICS)
	fmt.Printf("%s", strings.Repeat(HORIZLINE, oneside))
	fmt.Printf("%s%s", LEFTTEE, EXITGRAPHICS)
	fmt.Printf("%s %s %s", VIDBOLDREV, x.screenName, VIDOFF)
	fmt.Printf("%s%s", ENTERGRAPHICS, RIGHTTEE)
	fmt.Printf("%s", strings.Repeat(HORIZLINE, oneside+1))
	fmt.Printf("%s", EXITGRAPHICS)
	setScrRegion(2, x.dispHeight)
}

func (x *browseObj) pageLast() {
	x.printPage(x.mapSiz)
}

func (x *browseObj) pageMarked(lineno int) {
	x.printPage(x.marks[lineno])
}

// vim: set ts=4 sw=4 noet:
