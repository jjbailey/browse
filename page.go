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
	// calculate available width for title
	// account for tees and spaces
	availableWidth := br.dispWidth - 4

	// prepare title with ellipsis if needed
	dispTitle := br.title
	if len(br.title) > availableWidth {
		// leave room for ellipsis
		dispTitle = "..." + br.title[len(br.title)-availableWidth+4+3:]
	}

	// calculate padding for centering
	padding := (availableWidth - len(dispTitle)) >> 1

	// build header
	// -----| title |-----
	var sb strings.Builder
	sb.Grow(br.dispWidth + 20) // Pre-allocate space

	// left side
	sb.WriteString(ENTERGRAPHICS)
	sb.WriteString(strings.Repeat(HORIZLINE, padding))
	sb.WriteString(LEFTTEE)
	sb.WriteString(EXITGRAPHICS)

	// title
	sb.WriteString(VIDBOLDREV)
	sb.WriteString(" ")
	sb.WriteString(dispTitle)
	sb.WriteString(" ")
	sb.WriteString(VIDOFF)

	// right side
	sb.WriteString(ENTERGRAPHICS)
	sb.WriteString(RIGHTTEE)
	sb.WriteString(strings.Repeat(HORIZLINE, availableWidth-padding-len(dispTitle)))
	sb.WriteString(EXITGRAPHICS)

	// display header
	resetScrRegion()
	moveCursor(1, 1, true)
	fmt.Print(CLEARSCREEN)
	fmt.Print(LINEWRAPOFF)
	fmt.Print(sb.String())
	setScrRegion(2, br.dispHeight)
}

func (br *browseObj) pageLast() {
	br.printPage(br.mapSiz)
}

func (br *browseObj) pageMarked(lineno int) {
	br.printPage(br.marks[lineno])
}

// vim: set ts=4 sw=4 noet:
