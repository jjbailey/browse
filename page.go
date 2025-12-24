// page.go
// Paging and some support functions
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
	// Minimum required width for the header (tees, spaces, and some title)
	const minHeaderWidth = 10

	// Validate display width
	if br.dispWidth < minHeaderWidth {
		return
	}

	// Calculate available width for title (account for tees and spaces)
	availableWidth := br.dispWidth - 4

	// Prepare title with ellipsis if needed
	dispTitle := br.title
	if len(br.title) > availableWidth {
		// Calculate start index for the title substring
		// Leave room for ellipsis (3 chars) and some title text
		startIndex := len(br.title) - (availableWidth - 7)
		if startIndex < 0 {
			startIndex = 0
		}
		if startIndex > len(br.title) {
			startIndex = len(br.title)
		}
		dispTitle = "..." + br.title[startIndex:]
	}

	// Calculate padding for centering
	padding := (availableWidth - len(dispTitle)) >> 1

	// build header
	// ─────┤ title ├─────

	var sb strings.Builder

	sb.Grow(br.dispWidth + 20)

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
	header := sb.String()
	sb.Reset()

	sb.WriteString(fmt.Sprintf(CURPOS, 1, 1))
	sb.WriteString(CLEARSCREEN)
	sb.WriteString(LINEWRAPOFF)
	sb.WriteString(fmt.Sprintf(SCROLLREGION, 2, br.dispHeight))
	sb.WriteString(header)
	fmt.Print(sb.String())
}

func (br *browseObj) pageLast() {
	br.printPage(br.mapSiz)
}

func (br *browseObj) pageMarked(lineno int) {
	br.printPage(br.marks[lineno])
}

// vim: set ts=4 sw=4 noet:
