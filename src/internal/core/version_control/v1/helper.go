package v1

import (
	"MultiRepoVC/src/internal/core/version_control/v1/model"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func SaveObject(hash string, content []byte) error {
	if len(hash) < 3 {
		return errors.New("invalid hash length")
	}

	// Path breakdown: first 2 chars → directory, rest → file name
	dir := filepath.Join(".mrvc", "objects", hash[:2])
	file := filepath.Join(dir, hash[2:])

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write object content
	return os.WriteFile(file, content, 0644)
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func addOrReplaceTreeEntry(tree model.TreeObject, entry model.TreeEntry) model.TreeObject {
	updated := false
	for i, e := range tree.Entries {
		if e.Name == entry.Name && e.EntryType == entry.EntryType {
			tree.Entries[i] = entry
			updated = true
			break
		}
	}
	if !updated {
		tree.Entries = append(tree.Entries, entry)
	}
	return tree
}

func readHEAD() string {
	data, err := os.ReadFile(".mrvc/HEAD")
	if err != nil {
		return "" // no HEAD yet
	}
	return strings.TrimSpace(string(data))
}

func updateHEAD(hash string) error {
	return os.WriteFile(".mrvc/HEAD", []byte(hash), 0644)
}
