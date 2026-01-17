// setupbrdir.go
// create .browse directory
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"os"
	"path/filepath"
	"strings"
)

// setupBrDir ensures the browse config directory exists and migrates old files.
func setupBrDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	newDir := filepath.Join(home, RCDIRNAME)
	if err := os.MkdirAll(newDir, 0700); err != nil {
		return err
	}

	// remove this pre-0.67 default file code at some point

	pattern := filepath.Join(home, (RCDIRNAME + "*"))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return nil
	}

	for _, oldPath := range matches {
		info, err := os.Stat(oldPath)
		if err != nil {
			continue
		}

		// Skip directories
		if info.IsDir() {
			continue
		}

		base := filepath.Base(oldPath)
		if !strings.HasPrefix(base, RCDIRNAME) {
			continue
		}

		// Remove the leading dot for the new filename
		newName := strings.TrimPrefix(base, ".")
		newPath := filepath.Join(newDir, newName)

		// Skip if destination already exists
		if _, err := os.Stat(newPath); err == nil {
			continue
		}

		// Move (rename) the file
		os.Rename(oldPath, newPath)
	}

	return nil
}

// vim: set ts=4 sw=4 noet:
