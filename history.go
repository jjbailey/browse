// history.go
// history management module
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loadHistory reads the history file and returns recent entries.
func loadHistory(historyFile string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	historyPath := filepath.Join(home, RCDIRNAME, historyFile)
	file, err := os.OpenFile(historyPath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return []string{}
	}
	defer file.Close()

	var history []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		history = append(history, scanner.Text())
	}

	// If history is too large, keep only the most recent entries
	if len(history) > maxHistorySize {
		history = history[len(history)-maxHistorySize:]
	}

	return history
}

// saveHistory writes history entries to disk with trimming and cleanup.
func saveHistory(history []string, historyFile string) {
	if len(history) == 0 {
		return
	}

	// Clean history in-place to avoid new slice allocation
	n := 0
	for _, entry := range history {
		if historyFile != searchHistory {
			entry = strings.TrimSpace(entry)
		}
		if entry != "" {
			history[n] = entry
			n++
		}
	}
	history = history[:n]

	if len(history) == 0 {
		return
	}

	// Ensure history doesn't exceed max size
	if len(history) > maxHistorySize {
		history = history[len(history)-maxHistorySize:]
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	historyPath := filepath.Join(home, RCDIRNAME, historyFile)
	file, err := os.OpenFile(historyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, cmd := range history {
		fmt.Fprintln(writer, cmd)
	}
	writer.Flush()
}

// updateDirHistory records directory changes in the directory history.
func updateDirHistory(savDir, curDir string) {
	// save the previous directory
	updateHistory(savDir, dirHistory)

	// save the current directory
	updateHistory(curDir, dirHistory)
}

// updateHistory appends a single entry to the specified history file.
func updateHistory(newEntry, historyFile string) {
	if historyFile != searchHistory {
		newEntry = strings.TrimSpace(newEntry)
	}

	if newEntry == "" {
		return
	}

	if historyFile == commHistory || historyFile == searchHistory {
		newEntry = unQuote(newEntry)
	} else {
		if strings.ContainsAny(newEntry, " ") && !strings.ContainsAny(newEntry, "'") {
			newEntry = "'" + newEntry + "'"
		}
	}

	history := loadHistory(historyFile)

	// Remove duplicate consecutive entries
	if len(history) > 0 && history[len(history)-1] == newEntry {
		return
	}

	history = append(history, newEntry)
	saveHistory(history, historyFile)
}

// vim: set ts=4 sw=4 noet:
