package util

import (
	"encoding/gob"
	"fmt"
	"os"
)

var File = FileUtil{} // exported singleton

type FileUtil struct{}

// ----------- PATH HELPERS ----------------

func (FileUtil) GetCurrentDir() (string, error) {
	return os.Getwd()
}

func (FileUtil) DoesDirExist(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateDir creates the directory and all parent directories.
func (FileUtil) CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// ----------- FILE WRITING ----------------

// WriteBinaryFile writes any struct using gob encoding.
func (FileUtil) WriteBinaryFile(path string, data any) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode binary: %w", err)
	}

	return nil
}

// WriteTextFile writes plain text (overwrites).
func (FileUtil) WriteTextFile(path string, text string) error {
	return os.WriteFile(path, []byte(text), 0644)
}
