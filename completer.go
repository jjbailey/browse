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
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	return runCompleter("$ ", commHistory)
}

func userFileComp() (string, bool) {
	searchType = searchFiles
	return runCompleter("File: ", fileHistory)
}

func runCompleter(promptStr, historyFile string) (string, bool) {
	// Load history from file
	history := loadHistory(historyFile)

	// Get hostname title
	title := getHostnameTitle()

	p := prompt.New(
		func(in string) { /* no-op executor */ },
		completer,
		prompt.OptionDescriptionBGColor(prompt.DarkGray),
		prompt.OptionDescriptionTextColor(prompt.Yellow),
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
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(buf *prompt.Buffer) {
				fmt.Print("\r" + CURUP)
			},
		}),
	)

	input := p.Input()
	// Restore terminal state after prompt exits
	fmt.Printf(XTERMTITLE, title)
	ttyBrowser()

	if len(input) == 0 {
		return "", true
	}

	return input, false
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
	word := strings.ReplaceAll(d.GetWordBeforeCursor(), "//", "/")

	// Detect and handle home directory expansion (including ~user)
	if strings.HasPrefix(word, "~") {
		word = expandHome(word)
	}

	// Decide completion type
	if isAbsOrRelPath(word) {
		return fileCompleter(word)
	}

	// If searchType == searchPath, complete $PATH executables
	if searchType == searchPath {
		return pathCompleter(word)
	}

	// Fallback: complete from current directory
	return dirCompleter(".", word, false)
}

func expandHome(word string) string {
	if len(word) == 1 || word[1] == '/' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return word
		}

		return filepath.Join(homeDir, word[1:])
	}

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

	// Use getHomeDir which reads /etc/passwd directly
	wordHome, err := getHomeDir(userName)
	if err != nil || wordHome == "" {
		// Return original word if user not found or error occurred
		return word
	}

	// Verify the home directory exists
	if _, err := os.Stat(wordHome); os.IsNotExist(err) {
		return word
	}

	return filepath.Join(wordHome, pathSuffix)
}

func isAbsOrRelPath(word string) bool {
	return strings.HasPrefix(word, "/") ||
		strings.HasPrefix(word, ".") ||
		strings.HasPrefix(word, "..") ||
		strings.Contains(word, "/")
}

func fileCompleter(word string) []prompt.Suggest {
	dir := filepath.Dir(word)
	_, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	prefix := word
	if !strings.HasSuffix(word, "/") {
		prefix = filepath.Base(word)
	} else {
		prefix = ""
	}

	return dirCompleter(dir, prefix, true)
}

func pathCompleter(word string) []prompt.Suggest {
	var suggestions []prompt.Suggest

	paths := strings.Split(os.Getenv("PATH"), ":")
	for _, dir := range paths {
		// Use short-circuit if enough suggestions
		if len(suggestions) >= maxSuggestions {
			break
		}

		_, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		suggestions = append(suggestions, dirCompleter(dir, word, false)...)
	}

	if len(suggestions) > maxSuggestions {
		suggestions = suggestions[:maxSuggestions]
	}

	return suggestions
}

func dirCompleter(dir, prefix string, useFullPath bool) []prompt.Suggest {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	suggestions := make([]prompt.Suggest, 0, dispSuggestions)

	for _, file := range files {
		if len(suggestions) >= maxSuggestions {
			break
		}

		name := file.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		text := name
		if useFullPath {
			text = filepath.Join(dir, name)
		}

		desc := ""

		switch {

		case file.Type()&os.ModeSymlink != 0:
			desc = "-> " + resolveSymlink(filepath.Join(dir, name))

		case file.Type()&os.ModeNamedPipe != 0:
			desc = "named pipe"

		case file.IsDir():
			desc = "directory"
		}

		suggestions = append(suggestions, prompt.Suggest{
			Text:        text,
			Description: desc,
		})
	}

	return suggestions
}

// vim: set ts=4 sw=4 noet:
