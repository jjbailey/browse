// page.go
// paging and some support functions
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
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
	var i int

	linelen := x.dispWidth - 4 // minus tees and spaces
	oneside := (linelen - len(x.screenName)) / 2

	movecursor(1, 1, true)
	fmt.Printf("%s", CLEARSCREEN)
	fmt.Printf("%s", ENTERGRAPHICS)

	for i = 0; i < oneside; i++ {
		fmt.Printf("%s", HORIZLINE)
	}

	fmt.Printf("%s%s", LEFTTEE, EXITGRAPHICS)
	fmt.Printf("%s %s %s", VIDBOLDREV, x.screenName, VIDOFF)
	fmt.Printf("%s%s", ENTERGRAPHICS, RIGHTTEE)

	for i = 0; i <= oneside; i++ {
		fmt.Printf("%s", HORIZLINE)
	}

	fmt.Printf("%s", EXITGRAPHICS)
}

func (x *browseObj) pageLast() {
	x.printPage(x.mapSiz)
}

func (x *browseObj) pageMarked(lineno int) {
	x.printPage(x.marks[lineno])
}

// vim: set ts=4 sw=4 noet:
