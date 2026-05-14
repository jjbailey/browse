// completer.go
// central completion module
//
// Copyright (c) 2024-2026 jjb
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

	"github.com/jjbailey/go-prompt"
)

// Completion search modes.
const (
	dispSuggestions = 5
	maxSuggestions  = 1000
	searchFiles     = 1
	searchPath      = 2
	searchSearch    = 3
	searchDirs      = 4
)

// Completion filters for file types.
const (
	onlyDirs         = 1
	onlyExec         = 2
	onlyFiles        = 3
	onlyFilesAndDirs = 4
)

// SearchType controls which completion mode is active.
var SearchType int

type completionCandidate struct {
	name       string
	suggestion prompt.Suggest
}

type pathCompletionCache struct {
	path       string
	candidates []completionCandidate
	loaded     bool
}

var pathCache pathCompletionCache

// userDirComp prompts for a directory with completion.
func userDirComp() (string, bool) {
	SearchType = searchDirs
	return runCompleter("Dir: ", dirHistory)
}

// userFileComp prompts for a file with completion.
func userFileComp() (string, bool) {
	SearchType = searchFiles
	return runCompleter("File: ", fileHistory)
}

// userBashComp prompts for a command with PATH-aware completion.
func userBashComp() (string, bool) {
	SearchType = searchPath
	return runCompleter("$ ", commHistory)
}

// userSearchComp prompts for a search pattern with completion.
func userSearchComp(searchDir bool) (string, bool) {
	promptStr := "/"
	if !searchDir {
		promptStr = "?"
	}

	SearchType = searchSearch
	return runCompleter(promptStr, searchHistory)
}

// runCompleter starts the prompt UI and returns user input and cancellation state.
func runCompleter(promptStr, historyFile string) (string, bool) {
	history := loadHistory(historyFile)
	pathCache = pathCompletionCache{}

	// reset go-prompt BackedOut flag
	prompt.BackedOut = false

	// set RawPrefix to allow escape chars in prefix, turns off color
	// escape char usage works only in simple cases, not here
	prompt.RawPrefix = true

	p := prompt.New(
		func(in string) { /* no-op executor */ },
		completer,
		prompt.OptionDescriptionBGColor(prompt.DarkGray),
		prompt.OptionDescriptionTextColor(prompt.Yellow),
		prompt.OptionHistory(history),
		prompt.OptionMaxSuggestion(dispSuggestions),
		prompt.OptionPrefix(promptStr),
		prompt.OptionScrollbarBGColor(prompt.DefaultColor),
		prompt.OptionScrollbarThumbColor(prompt.DefaultColor),
		prompt.OptionSelectedSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSelectedSuggestionTextColor(prompt.Yellow),
		prompt.OptionSwitchKeyBindMode(prompt.EmacsKeyBind),
		prompt.OptionDisableTitle(),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(buf *prompt.Buffer) {
				fmt.Print("\r" + CURUP)
			},
		}),
	)

	input := p.Input()
	ttyBrowser()

	if len(input) == 0 {
		return "", prompt.BackedOut
	}

	return input, false
}

// completer provides suggestions based on the current SearchType and input.
func completer(d prompt.Document) []prompt.Suggest {
	word := strings.ReplaceAll(d.GetWordBeforeCursor(), "//", "/")
	originalWord := word

	// Handle tilde expansions early to prevent duplicated code

	if strings.HasPrefix(word, "~-") {
		word = prevDirectory()
	}

	if strings.HasPrefix(word, "~") {
		word = expandHome(word)
	}

	switch SearchType {

	case searchSearch:
		return searchCompleter()

	case searchDirs:
		return dirCompleter(word)

	case searchFiles:
		if hasPathSeparator(word) {
			return fileCompleter(word)
		}
		return anyCompleter(".", originalWord, onlyFiles)

	case searchPath:
		if hasPathSeparator(word) {
			return fileCompleter(word)
		}

		// Current user input before the cursor
		text := d.TextBeforeCursor()
		parts := strings.Fields(text)
		n := len(parts)
		tokenIndex := n
		if word != "" {
			tokenIndex = n - 1
		}
		isCommand := tokenIndex == 0 || (tokenIndex > 0 && parts[tokenIndex-1] == "|")

		if isCommand {
			// If currently typing (partial) after a pipe: "|cmd"
			if strings.HasPrefix(word, "|") {
				if word == "|" {
					return nil
				}

				suggestions := pathCompleter(word[1:])
				for i := range suggestions {
					suggestions[i].Text = "|" + suggestions[i].Text
				}
				return suggestions
			}

			// By default, normal path completion
			return pathCompleter(word)
		}

		// Otherwise, complete shell arguments
		return anyCompleter(".", originalWord, onlyFilesAndDirs)

	default:
		return anyCompleter(".", originalWord, onlyFiles)
	}
}

// hasPathSeparator reports whether the word contains a path separator.
func hasPathSeparator(word string) bool {
	return strings.Contains(word, "/")
}

// searchCompleter returns suggestions for search history (currently none).
func searchCompleter() []prompt.Suggest {
	return nil
}

// fileCompleter completes file paths for a given word.
func fileCompleter(word string) []prompt.Suggest {
	dir := filepath.Dir(word)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	prefix := word
	if !strings.HasSuffix(word, "/") {
		prefix = filepath.Base(word)
	} else {
		prefix = ""
	}

	return matchFiles(files, dir, prefix, true, onlyFilesAndDirs)
}

// pathCompleter completes executable names from PATH entries.
func pathCompleter(word string) []prompt.Suggest {
	candidates := pathCompleterCandidates()
	suggestions := make([]prompt.Suggest, 0, minimum(len(candidates), maxSuggestions))

	for _, candidate := range candidates {
		if !strings.HasPrefix(candidate.name, word) {
			continue
		}

		suggestions = append(suggestions, candidate.suggestion)
		if len(suggestions) >= maxSuggestions {
			break
		}
	}

	return suggestions
}

// pathCompleterCandidates caches executable candidates for one prompt session.
func pathCompleterCandidates() []completionCandidate {
	path := os.Getenv("PATH")
	if pathCache.loaded && pathCache.path == path {
		return pathCache.candidates
	}

	pathCache.path = path
	pathCache.candidates = nil
	pathCache.loaded = true

	// cannot test PATH with go run .
	// go build .

	paths := strings.Split(path, ":")
	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "") {
		paths = []string{"/usr/local/bin", "/usr/bin", "/usr/sbin"}
	}

	for _, dir := range paths {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		pathCache.candidates = append(pathCache.candidates,
			matchFileCandidates(files, dir, "", false, onlyExec, 0)...)
	}

	return pathCache.candidates
}

// dirCompleter completes directories using absolute paths and CDPATH.
func dirCompleter(word string) []prompt.Suggest {
	var suggestions []prompt.Suggest

	if hasPathSeparator(word) {
		dir := filepath.Dir(word)
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}

		prefix := word
		if !strings.HasSuffix(word, "/") {
			prefix = filepath.Base(word)
		} else {
			prefix = ""
		}

		suggestions = matchFiles(files, dir, prefix, true, onlyDirs)

		if len(suggestions) > maxSuggestions {
			suggestions = suggestions[:maxSuggestions]
			return suggestions
		}
	}

	// Handle other paths

	paths := strings.Split(os.Getenv("CDPATH"), ":")
	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "") {
		paths = []string{".", ".."}
	}

	for _, dir := range paths {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		suggestions = append(suggestions,
			matchFiles(files, dir, word, true, onlyDirs)...)

		if len(suggestions) > maxSuggestions {
			suggestions = suggestions[:maxSuggestions]
			break
		}
	}

	return suggestions
}

// anyCompleter completes entries in a directory with optional filtering.
func anyCompleter(dir, prefix string, onlyType int) []prompt.Suggest {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	// Match all directory entries
	return matchFiles(files, dir, prefix, false, onlyType)
}

// matchFiles filters directory entries into prompt suggestions.
func matchFiles(files []os.DirEntry, dir, prefix string,
	useFullPath bool, onlyType int) []prompt.Suggest {

	candidates := matchFileCandidates(files, dir, prefix, useFullPath, onlyType, maxSuggestions)
	suggestions := make([]prompt.Suggest, 0, len(candidates))

	for _, candidate := range candidates {
		suggestions = append(suggestions, candidate.suggestion)
	}

	return suggestions
}

// matchFileCandidates filters directory entries and keeps raw names for caching.
func matchFileCandidates(files []os.DirEntry, dir, prefix string,
	useFullPath bool, onlyType, limit int) []completionCandidate {

	candidates := make([]completionCandidate, 0, minimum(len(files), maxSuggestions))

	for _, file := range files {
		if limit > 0 && len(candidates) >= limit {
			break
		}

		name := file.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		fullPath := filepath.Join(dir, name)

		if !matchesFileType(file, fullPath, onlyType) {
			continue
		}

		// Determine display name (quote if contains spaces)
		var displayName string

		if useFullPath {
			displayName = fullPath
		} else {
			displayName = name
		}

		if strings.Contains(displayName, " ") {
			if strings.Contains(displayName, "'") {
				displayName = "\"" + strings.ReplaceAll(displayName, "\"", "\\\"") + "\""
			} else {
				displayName = "'" + displayName + "'"
			}
		}

		var desc string

		switch {

		case file.Type()&os.ModeSymlink != 0:
			linkTarget, _ := resolveSymlink(fullPath)
			desc = "-> " + linkTarget

		case file.IsDir():
			desc = "directory"

		case file.Type()&os.ModeNamedPipe != 0:
			desc = "named pipe"

		case file.Type()&os.ModeSocket != 0:
			desc = "socket"

		case file.Type()&os.ModeCharDevice != 0:
			desc = "character device"

		case file.Type()&os.ModeDevice != 0:
			desc = "block device"

		default:
			if SearchType == searchFiles {
				if isBinaryFile(fullPath) {
					desc = "binary file"
				} else {
					desc = "regular file"
				}
			}
		}

		candidates = append(candidates, completionCandidate{
			name: name,
			suggestion: prompt.Suggest{
				Text:        displayName,
				Description: desc,
			},
		})
	}

	return candidates
}

// matchesFileType avoids Stat where ReadDir already supplied enough type data.
func matchesFileType(file os.DirEntry, fullPath string, onlyType int) bool {
	modeType := file.Type()
	isSymlink := modeType&os.ModeSymlink != 0

	switch onlyType {

	case onlyDirs:
		if !isSymlink {
			return file.IsDir()
		}

		info, err := os.Stat(fullPath)
		return err == nil && info.IsDir()

	case onlyExec:
		info, err := statForCompletion(file, fullPath)
		return err == nil && !info.IsDir() && info.Mode().Perm()&0111 != 0

	case onlyFiles:
		if !isSymlink && modeType != 0 {
			return false
		}

		info, err := statForCompletion(file, fullPath)
		return err == nil && info.Mode().IsRegular()

	case onlyFilesAndDirs:
		if !isSymlink && file.IsDir() {
			return true
		}
		if !isSymlink && modeType != 0 {
			return false
		}

		info, err := statForCompletion(file, fullPath)
		return err == nil && (info.Mode().IsRegular() || info.IsDir())
	}

	return true
}

// statForCompletion follows symlinks but uses DirEntry info for normal entries.
func statForCompletion(file os.DirEntry, fullPath string) (os.FileInfo, error) {
	if file.Type()&os.ModeSymlink != 0 {
		return os.Stat(fullPath)
	}

	return file.Info()
}

// vim: set ts=4 sw=4 noet:
