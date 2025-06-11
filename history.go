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

const (
	commHistory = ".browse_shell"
	fileHistory = ".browse_files"
)

const maxHistorySize = 500

func loadHistory(historyFile string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	historyPath := filepath.Join(home, historyFile)
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
	// Return immediately if history is empty
	if len(history) == 0 {
		return
	}

	// Clean and validate history entries
	validHistory := make([]string, 0, len(history))
	for _, entry := range history {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		validHistory = append(validHistory, trimmed)
	}

	// Return if no valid entries after cleaning
	if len(validHistory) == 0 {
		return
	}

	// Remove duplicate consecutive entries
	if len(validHistory) >= 2 && validHistory[len(validHistory)-1] == validHistory[len(validHistory)-2] {
		validHistory = validHistory[:len(validHistory)-1]
	}

	// Ensure history doesn't exceed max size
	if len(validHistory) > maxHistorySize {
		validHistory = validHistory[len(validHistory)-maxHistorySize:]
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	historyPath := filepath.Join(home, historyFile)
	file, err := os.OpenFile(historyPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	// Truncate the file to zero length after opening
	if err := file.Truncate(0); err != nil {
		return
	}

	writer := bufio.NewWriter(file)
	for _, cmd := range validHistory {
		fmt.Fprintln(writer, cmd)
	}
	writer.Flush()
}

// vim: set ts=4 sw=4 noet:
