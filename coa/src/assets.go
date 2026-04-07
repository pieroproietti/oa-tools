// Copyright 2026 Piero Proietti <piero.proietti@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed assets/*
var internalConfigs embed.FS

func ExtractConfigs(destRoot string) error {
	return fsCopy(internalConfigs, "assets", destRoot)
}

func fsCopy(fs embed.FS, src, dest string) error {
	entries, err := fs.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// NOTA: Name() è un metodo, servono le parentesi!
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			if err := fsCopy(fs, srcPath, destPath); err != nil {
				return err
			}
		} else {
			data, err := fs.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
