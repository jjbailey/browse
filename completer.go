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

	// Get hostname title
	title := getHostnameTitle()
	fmt.Printf("\033]0;%s\007", title)

	ttyBrowser()
	return input, flag
}

func userFileComp() (string, bool) {
	searchType = searchFiles
	input, flag := runCompleter("File: ", fileHistory)

	// Get hostname title
	title := getHostnameTitle()
	fmt.Printf("\033]0;%s\007", title)

	ttyBrowser()
	return input, flag
}

func runCompleter(promptStr, historyFile string) (string, bool) {
	// Load history from file
	history := loadHistory(historyFile)

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

	// Get hostname title
	title := getHostnameTitle()

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
			return "", true
		}
		return input, false

	case <-ctx.Done():
		return "", true
	}
}

func getHostnameTitle() string {
	// Can't prevent prompt.New from meddling with the title, so...

	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	if dotIndex := strings.Index(hostname, "."); dotIndex != -1 {
		hostname = hostname[:dotIndex]
	}

	return hostname
}

func completer(d prompt.Document) []prompt.Suggest {
	var suggestions []prompt.Suggest

	// Get the word being completed
	word := strings.ReplaceAll(d.GetWordBeforeCursor(), "//", "/")

	// Handle home directory expansion
	if strings.HasPrefix(word, "~") {
		if len(word) == 1 || word[1] == '/' {
			// ~ or ~/something
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return suggestions
			}
			word = filepath.Join(homeDir, word[1:])
		} else {
			// ~username or ~username/something
			slashIndex := strings.IndexRune(word, '/')
			var userName, pathSuffix string
			if slashIndex == -1 {
				userName = word[1:]
				pathSuffix = ""
			} else {
				userName = word[1:slashIndex]
				pathSuffix = word[slashIndex:]
			}

			word, err := GetHomeDir(userName)
			if err != nil {
				return suggestions
			}
			word = filepath.Join(word, pathSuffix)
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
