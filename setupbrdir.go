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
)

// setupBrDir ensures the browse config directory exists.
func setupBrDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(filepath.Join(home, RCDIRNAME), 0700)
}

// vim: set ts=4 sw=4 noet:
