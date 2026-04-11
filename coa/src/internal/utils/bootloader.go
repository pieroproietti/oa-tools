package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	// ELIMINATO: "coa/src/internal/engine"
)

const BootloaderURL = "https://github.com/pieroproietti/penguins-bootloaders/releases/download/v26.1.16/bootloaders.tar.gz"

// EnsureBootloaders ora riceve il percorso come argomento
func EnsureBootloaders(targetDir string) (string, error) {
	// 1. Controllo se esistono già
	if _, err := os.Stat(targetDir); err == nil {
		return targetDir, nil
	}

	fmt.Printf("\033[1;33m[coa]\033[0m Bootloaders non trovati in %s. Inizio download...\n", targetDir)

	// 2. Download
	resp, err := http.Get(BootloaderURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("errore download: status %d", resp.StatusCode)
	}

	// 3. Estrazione
	if err := extractTarGz(resp.Body, targetDir); err != nil {
		return "", err
	}

	return targetDir, nil
}

func extractTarGz(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		cleanPath := filepath.Clean(header.Name)
		parts := strings.Split(cleanPath, string(filepath.Separator))

		var relPath string
		if len(parts) > 1 {
			relPath = filepath.Join(parts[1:]...)
		} else {
			continue // Salta la cartella radice del tarball
		}

		target := filepath.Join(dest, relPath)

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			f, _ := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			io.Copy(f, tr)
			f.Close()
		}
	}
	return nil
}
