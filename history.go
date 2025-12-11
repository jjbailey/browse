// history.go
// history management module
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
	"path/filepath"
	"strings"
)

func loadHistory(historyFile string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	historyPath := filepath.Join(home, (RCDIRNAME + "/" + historyFile))
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

func saveHistory(history []string, historyFile string) {
	if len(history) == 0 {
		return
	}

	// Clean history in-place to avoid new slice allocation
	n := 0
	for _, entry := range history {
		trimmed := strings.TrimSpace(entry)
		if trimmed != "" {
			history[n] = trimmed
			n++
		}
	}
	history = history[:n]

	if len(history) == 0 {
		return
	}

	// Remove duplicate consecutive entries
	if len(history) >= 2 && history[len(history)-1] == history[len(history)-2] {
		history = history[:len(history)-1]
	}

	// Ensure history doesn't exceed max size
	if len(history) > maxHistorySize {
		history = history[len(history)-maxHistorySize:]
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	historyPath := filepath.Join(home, (RCDIRNAME + "/" + historyFile))
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

func updateDirHistory(savDir, curDir string) {
	// save the previous directory
	updateHistory(savDir, dirHistory)

	// save the current directory
	updateHistory(curDir, dirHistory)
}

func updateHistory(newEntry, historyFile string) {
	// saveHistory checks for dups

	if newEntry == "" {
		return
	}

	if strings.ContainsAny(newEntry, " ") && !strings.ContainsAny(newEntry, "'") {
		newEntry = "'" + newEntry + "'"
	}

	history := append(loadHistory(historyFile), newEntry)
	saveHistory(history, historyFile)
}

// vim: set ts=4 sw=4 noet:
