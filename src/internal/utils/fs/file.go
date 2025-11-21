package fs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// GetCurrentDir returns the absolute path of the current working directory.
func GetCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

// IsDirPresent checks if a directory exists at the given path.
func IsDirPresent(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CreateDir creates a directory with 755 permissions.
// It creates all parent dirs if needed.
func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// WriteJSON writes any Go struct/map as pretty JSON to a file.
// It automatically creates directories if the parent folder doesn't exist.
func WriteJSON(path string, data any) error {
	parent := filepath.Dir(path)

	if !IsDirPresent(parent) {
		if err := CreateDir(parent); err != nil {
			return err
		}
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0644)
}

// FileExists checks whether a file exists.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ReadJSON reads JSON file into the target struct.
func ReadJSON(path string, target any) error {
	if !FileExists(path) {
		return errors.New("file not found: " + path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}
