package v1

import (
	"MultiRepoVC/src/internal/core/version_control/v1/model"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// SaveObject OBJECT STORAGE
// These helpers implement Git-style loose object storage.
// An object is stored at:
//
//	.mrvc/objects/<first2>/<rest>
//
// This keeps directories small and lookup fast.
func SaveObject(hash string, content []byte) error {
	if len(hash) < 3 {
		return errors.New("invalid hash length")
	}

	// directory split improves filesystem scalability
	dir := filepath.Join(".mrvc", "objects", hash[:2])
	file := filepath.Join(dir, hash[2:])

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(file, content, 0644)
}

// HashContent HASH HELPERS
// Each object (blob, tree, commit) is content-addressable.
// HashContent hashes raw bytes.
// HashTree hash the JSON representation of a directory tree.
// HashCommit hash the JSON representation of a commit.
// This mirrors Git's core behavior.
func HashContent(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func HashTree(tree model.TreeObject) (string, []byte, error) {
	jsonBytes, err := json.Marshal(tree)
	if err != nil {
		return "", nil, err
	}
	h := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(h[:]), jsonBytes, nil
}

func HashCommit(commit model.CommitObject) (string, []byte, error) {
	jsonBytes, err := json.Marshal(commit)
	if err != nil {
		return "", nil, err
	}
	h := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(h[:]), jsonBytes, nil
}

// TREE HELPERS
// addOrReplaceTreeEntry ensures no duplicate directory or file entries
// exist inside a tree. If an entry already exists, it updates it.
func addOrReplaceTreeEntry(tree model.TreeObject, entry model.TreeEntry) model.TreeObject {
	for i, e := range tree.Entries {
		if e.Name == entry.Name && e.EntryType == entry.EntryType {
			tree.Entries[i] = entry
			return tree
		}
	}
	tree.Entries = append(tree.Entries, entry)
	return tree
}

// HEAD HELPERS
// readHEAD returns the current commit hash (or empty if no commits)
// updateHEAD moves HEAD to a new commit
func readHEAD() string {
	data, err := os.ReadFile(".mrvc/HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func updateHEAD(hash string) error {
	return os.WriteFile(".mrvc/HEAD", []byte(strings.TrimSpace(hash)), 0644)
}
