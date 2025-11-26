package util

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

var File = FileUtil{
	hasher: sha256.New(),
} // exported singleton

type FileUtil struct {
	hasher hash.Hash
}

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

func (fu FileUtil) CalculateHashOfFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := fu.hasher
	// io.Copy handles chunking automatically
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString, nil
}

func (fu FileUtil) SaveObject(hashString string, path string, savePath string) error {
	// Open source file
	sourceFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create directory structure: savePath/first_2_letters/
	first2 := hashString[:2]
	restOfHash := hashString[2:]

	dirPath := filepath.Join(savePath, first2)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	// Create destination file path
	destPath := filepath.Join(dirPath, restOfHash)

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Stream copy from source to destination
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return nil
}

// NormalizePath cleans and normalizes a file path by removing redundant separators,
// resolving relative path elements (. and ..), and converting all path separators
// to forward slashes (/). The path remains relative if it was relative, or absolute
// if it was absolute.
//
// Examples:
//   - "folder\subfolder\file.txt" -> "folder/subfolder/file.txt"
//   - "folder//subfolder/../file.txt" -> "folder/file.txt"
//   - "./folder/./file.txt" -> "folder/file.txt"
func (fu FileUtil) NormalizePath(path string) (string, error) {
	// Clean the path (removes redundant separators, resolves . and ..)
	cleanPath := filepath.Clean(path)

	// Convert all backslashes to forward slashes
	normalizedPath := filepath.ToSlash(cleanPath)

	return normalizedPath, nil
}
