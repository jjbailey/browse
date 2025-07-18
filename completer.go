// completer.go
// central completion module (bash + file)
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/c-bata/go-prompt"
)

const (
	dispSuggestions = 5
	maxSuggestions  = 1000
	searchFiles     = 1
	searchPath      = 2
)

var searchType int

func userBashComp() (string, bool) {
	searchType = searchPath
	input, flag := runCompleter("$ ", commHistory)
	ttyBrowser()
	return input, flag
}

func userFileComp() (string, bool) {
	searchType = searchFiles
	input, flag := runCompleter("File: ", fileHistory)
	ttyBrowser()
	return input, flag
}

func runCompleter(promptStr, historyFile string) (string, bool) {
	// Load history from file
	history := loadHistory(historyFile)

	// Can't prevent prompt.New from meddling with the title, so...
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	} else {
		if dotIndex := strings.Index(hostname, "."); dotIndex != -1 {
			hostname = hostname[:dotIndex]
		}
	}
	title := hostname

	// Create a context that can be cancelled
	ctx, cancelled := context.WithCancel(context.Background())
	defer cancelled()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	defer signal.Stop(sigChan)

	// Start a goroutine to handle Ctrl+C
	go func() {
		<-sigChan
		fmt.Printf("Ctrl+C pressed\n")
		cancelled()
	}()

	p := prompt.New(
		func(in string) { /* no-op executor */ },
		completer,
		prompt.OptionDescriptionTextColor(prompt.Green),
		prompt.OptionHistory(history),
		prompt.OptionMaxSuggestion(dispSuggestions),
		prompt.OptionPrefix(promptStr),
		prompt.OptionPrefixTextColor(prompt.White),
		prompt.OptionScrollbarBGColor(prompt.DefaultColor),
		prompt.OptionScrollbarThumbColor(prompt.DefaultColor),
		prompt.OptionSelectedSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSelectedSuggestionTextColor(prompt.Yellow),
		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSwitchKeyBindMode(prompt.EmacsKeyBind),
		prompt.OptionTitle(title),
	)

	// Start a goroutine to handle input
	inputChan := make(chan string)
	go func() {
		inputChan <- p.Input()
	}()

	// Wait for either input, Ctrl+C, or context cancellation
	// Restore shell title on empty input
	select {
	case input := <-inputChan:
		if len(input) == 0 {
			fmt.Printf("\033]0;%s\007", hostname)
			return "", true
		}
		return input, false

	case <-ctx.Done():
		fmt.Printf("\033]0;%s\007", hostname)
		return "", true
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	var suggestions []prompt.Suggest

	// Get the word being completed
	word := strings.ReplaceAll(d.GetWordBeforeCursor(), "//", "/")

	// Handle home directory expansion
	if strings.HasPrefix(word, "~") {
		// Check if it's a user-specific home directory (e.g., ~jjb)
		if len(word) > 1 && word[1] != '/' {
			// Extract username (everything between ~ and / or end of string)
			username := word[1:]
			if idx := strings.Index(username, "/"); idx != -1 {
				username = username[:idx]
			}

			// Look up the user
			u, err := user.Lookup(username)
			if err == nil && u != nil {
				// Replace ~username with the user's home directory
				word = u.HomeDir + word[len(username)+1:]
			} else {
				// User does not exist, skip expansion and return current suggestions
				return suggestions
			}
		} else {
			// Regular home directory expansion
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return suggestions
			}

			// Replace ~ with the home directory path
			word = homeDir + word[1:]
		}
	}

	// Handle absolute paths and relative paths
	if strings.HasPrefix(word, "/") ||
		strings.HasPrefix(word, ".") ||
		strings.HasPrefix(word, "..") ||
		strings.HasPrefix(word, "~") ||
		strings.Contains(word, "/") {

		// Get the directory part of the path
		dir := filepath.Dir(word)

		// List files in the specified directory
		files, err := os.ReadDir(dir)
		if err != nil {
			return suggestions
		}

		// If the word ends with a slash, we're looking for everything in that directory
		// Otherwise, we're looking for files/dirs that match the base name
		var baseWord string
		if strings.HasSuffix(word, "/") {
			baseWord = ""
		} else {
			baseWord = filepath.Base(word)
		}

		for _, file := range files {
			if len(suggestions) >= maxSuggestions {
				break
			}

			if file.IsDir() {
				if baseWord == "" || strings.HasPrefix(file.Name(), baseWord) {
					suggestions = append(suggestions, prompt.Suggest{
						Text: filepath.Join(dir, file.Name()),
					})
				}

				continue
			}

			// Include all files
			if baseWord == "" || strings.HasPrefix(file.Name(), baseWord) {
				suggestions = append(suggestions, prompt.Suggest{
					Text: filepath.Join(dir, file.Name()),
				})
			}
		}

		return suggestions
	}

	var pathDirs []string

	if searchType == searchPath {
		// Handle $PATH
		pathDirs = strings.Split(os.Getenv("PATH"), ":")
	} else {
		// Handle current directory
		pathDirs = []string{"."}
	}

	for _, dir := range pathDirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if len(suggestions) >= maxSuggestions {
				break
			}

			if file.IsDir() {
				if strings.HasPrefix(file.Name(), word) {
					suggestions = append(suggestions, prompt.Suggest{
						Text: file.Name(),
					})
				}

				continue
			}

			// Include all files
			if strings.HasPrefix(file.Name(), word) {
				suggestions = append(suggestions, prompt.Suggest{
					Text: file.Name(),
				})
			}
		}
	}

	return suggestions
}

// vim: set ts=4 sw=4 noet:
