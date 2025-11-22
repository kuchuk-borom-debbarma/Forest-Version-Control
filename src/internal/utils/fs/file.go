package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// LoadIgnore reads .mrvcignore and returns all patterns.
func LoadIgnore(rootDir string) ([]string, error) {
	ignorePath := filepath.Join(rootDir, ".mrvcignore")

	data, err := os.ReadFile(ignorePath)
	if err != nil {
		// No ignore file → no patterns
		return []string{}, nil
	}

	lines := strings.Split(string(data), "\n")
	patterns := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}

	return patterns, nil
}

// IsIgnored checks if a file path matches ignore rules.
func IsIgnored(rootDir, path string, patterns []string) bool {
	rel, err := filepath.Rel(rootDir, path)
	if err != nil {
		return false
	}

	rel = filepath.ToSlash(rel)

	for _, p := range patterns {

		// match *.ext
		if strings.HasPrefix(p, "*") {
			if strings.HasSuffix(rel, p[1:]) {
				return true
			}
		}

		// match prefix*
		if strings.HasSuffix(p, "*") {
			if strings.HasPrefix(rel, p[:len(p)-1]) {
				return true
			}
		}

		// folder ignore
		if strings.HasSuffix(p, "/") {
			if strings.HasPrefix(rel, p) {
				return true
			}
		}

		// exact match
		if rel == p {
			return true
		}
	}

	return false
}

// WalkOptions defines optional scanning behavior.
type WalkOptions struct {
	IgnoreMRVC          bool // ignore the root repo's .mrvc folder
	IgnoreNestedRepos   bool // ignore nested repos (folders containing their own .mrvc)
	ApplyIgnorePatterns bool // apply .mrvcignore rules
}

// ListFiles returns files respecting the given walk options.
func ListFiles(rootDir string, opts WalkOptions) ([]string, error) {
	var files []string
	var patterns []string

	if opts.ApplyIgnorePatterns {
		patterns, _ = LoadIgnore(rootDir)
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// ------------------------------------------------
		// 1. Skip .mrvc folder of *this* repo
		// ------------------------------------------------
		if opts.IgnoreMRVC && strings.Contains(path, "/.mrvc") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// ------------------------------------------------
		// 2. Skip nested repos -> any folder with its own .mrvc
		// ------------------------------------------------
		if opts.IgnoreNestedRepos && info.IsDir() {
			if IsDirPresent(filepath.Join(path, ".mrvc")) {
				return filepath.SkipDir
			}
		}

		// ------------------------------------------------
		// 3. Apply ignore patterns from .mrvcignore
		// ------------------------------------------------
		if opts.ApplyIgnorePatterns && IsIgnored(rootDir, path, patterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// ------------------------------------------------
		// 4. Collect files only
		// ------------------------------------------------
		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// NormalizePath converts a file path into an absolute, clean, slash-normalized path.
func NormalizePath(p string) string {
	if p == "" {
		return ""
	}

	abs, err := filepath.Abs(p)
	if err != nil {
		abs = p
	}

	// Clean redundant components, then convert all separators to "/"
	clean := filepath.Clean(abs)
	return filepath.ToSlash(clean)
}

func CalculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()

	// Copy file content into the hasher
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	// Generate final hex-encoded hash string
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
