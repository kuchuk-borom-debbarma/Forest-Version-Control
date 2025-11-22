package v1

import (
	"MultiRepoVC/src/internal/core/version_control/v1/model"
	"MultiRepoVC/src/internal/utils/fs"
	"MultiRepoVC/src/internal/utils/time"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type VersionControlV1 struct{}

func New() *VersionControlV1 {
	return &VersionControlV1{}
}

// ======================================================================
// INIT
// ======================================================================

func (v *VersionControlV1) Init(repoName string, author string) error {
	root := fs.GetCurrentDir()
	mrvc := filepath.Join(root, ".mrvc")

	if fs.IsDirPresent(mrvc) {
		return errors.New("repository already initialized")
	}

	if err := fs.CreateDir(mrvc); err != nil {
		return err
	}

	meta := model.Metadata{
		Name:      repoName,
		Author:    author,
		CreatedAt: strconv.FormatInt(time.GetCurrentTimestamp(), 10),
	}

	return fs.WriteJSON(filepath.Join(mrvc, "metadata.json"), meta)
}

// ======================================================================
// COMMIT
// ======================================================================

func (v *VersionControlV1) Commit(message string, author string, files []string) error {
	if len(files) == 0 {
		return errors.New("no files to commit")
	}

	repoRoot := fs.GetCurrentDir()

	// -----------------------------
	// Wildcard "*" → commit all files
	// -----------------------------
	if len(files) == 1 && files[0] == "*" {
		all, err := fs.ListFiles(repoRoot, fs.WalkOptions{
			IgnoreMRVC:          true,
			IgnoreNestedRepos:   true, // IMPORTANT
			ApplyIgnorePatterns: true,
		})
		if err != nil {
			return err
		}

		files = make([]string, 0, len(all))
		for _, f := range all {
			files = append(files, fs.NormalizePath(f))
		}
	} else {
		for i, f := range files {
			normalized := fs.NormalizePath(f)
			files[i] = normalized
			if !fs.FileExists(normalized) {
				return errors.New("file does not exist: " + normalized)
			}
		}
	}

	// -----------------------------
	// Build directory → TreeObject
	// -----------------------------
	directoryTrees := make(map[string]model.TreeObject)

	// Parent → children mapping (optimization)
	children := make(map[string][]string)

	for _, filePath := range files {

		// --------------------------------------
		// 1. Blob
		// --------------------------------------
		//TODO stream large files instead of reading all at once
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		blobHash := HashContent(content)

		if err := SaveObject(blobHash, content); err != nil {
			return err
		}

		// --------------------------------------
		// 2. Determine directory of file
		// --------------------------------------
		fileDir := filepath.Dir(filePath)
		if fileDir == "." { //if file is in repo root directory, set to repo root
			fileDir = repoRoot
		}

		fileDir = fs.NormalizePath(fileDir)
		if _, exists := directoryTrees[fileDir]; !exists {
			directoryTrees[fileDir] = model.TreeObject{Entries: []model.TreeEntry{}}
		}

		// Add file entry into this directory tree
		tree := directoryTrees[fileDir]
		tree = addOrReplaceTreeEntry(tree, model.TreeEntry{
			Name:      filepath.Base(filePath),
			EntryType: "blob",
			Hash:      blobHash,
		})
		directoryTrees[fileDir] = tree

		// --------------------------------------
		// 3. Ensure all parent directories exist
		// --------------------------------------
		current := fileDir
		for current != repoRoot {
			parent := filepath.Dir(current)
			if parent == "." {
				parent = repoRoot
			}
			parent = fs.NormalizePath(parent)

			if _, ok := directoryTrees[parent]; !ok {
				directoryTrees[parent] = model.TreeObject{Entries: []model.TreeEntry{}}
			}

			children[parent] = append(children[parent], current)

			current = parent
		}
	}

	// ==================================================================
	// We must sort directories from deepest → shallowest because tree
	// hashes must be built bottom-up.
	//
	// A tree object contains the hashes of its children (files or
	// subtrees). Therefore, a parent directory cannot be hashed until
	// all of its subdirectories have already been hashed.
	//
	// By processing deeper directories first, we guarantee that when we
	// build a parent tree, all child tree hashes are already available.
	// This ensures deterministic, correct tree construction—just like
	// Git’s own object model.
	// ==================================================================

	var dirs []string
	for d := range directoryTrees {
		dirs = append(dirs, d)
	}

	sort.Slice(dirs, func(i, j int) bool {
		return strings.Count(dirs[i], "/") > strings.Count(dirs[j], "/")
	})
	// ==================================================================
	// BUILD TREES BOTTOM-UP (single pass)  O(N)
	//
	// After sorting folders deepest → shallowest, this loop constructs
	// the tree objects for every directory. For each folder:
	//   • Insert subtree entries using child directory hashes
	//   • Sort entries for deterministic hashing
	//   • Compute the tree hash
	//   • Save the tree object
	//
	// Processing bottom-up ensures that when we hash a directory, all
	// its children (files and subtrees) already have hashes available.
	// ==================================================================
	treeHashes := make(map[string]string)

	for _, dir := range dirs {
		tree := directoryTrees[dir]

		// Add subtree entries
		for _, child := range children[dir] {
			tree = addOrReplaceTreeEntry(tree, model.TreeEntry{
				Name:      filepath.Base(child),
				EntryType: "tree",
				Hash:      treeHashes[child],
			})
		}

		// Deterministic ordering
		sort.Slice(tree.Entries, func(i, j int) bool {
			return tree.Entries[i].Name < tree.Entries[j].Name
		})

		hash, jsonBytes, err := HashTree(tree)
		if err != nil {
			return err
		}

		if err := SaveObject(hash, jsonBytes); err != nil {
			return err
		}

		treeHashes[dir] = hash
	}

	rootTreeHash := treeHashes[repoRoot]

	// ==================================================================
	// CREATE COMMIT OBJECT
	// ==================================================================

	commit := model.CommitObject{
		Tree:      rootTreeHash,
		Parent:    readHEAD(),
		Message:   message,
		Author:    author,
		Timestamp: strconv.FormatInt(time.GetCurrentTimestamp(), 10),
	}

	commitHash, commitBytes, err := HashCommit(commit)
	if err != nil {
		return err
	}

	if err := SaveObject(commitHash, commitBytes); err != nil {
		return err
	}

	err = updateHEAD(commitHash)
	if err != nil {
		return err
	}

	log.Println("Commit created:", commitHash)
	return nil
}

// ======================================================================
// STATUS
// ======================================================================

func (v *VersionControlV1) Status() (string, error) {
	repoRoot := fs.GetCurrentDir()

	head := readHEAD()
	if head == "" {
		return "No commits yet.", nil
	}

	// ------------------------------------------------------
	// Load HEAD commit
	// ------------------------------------------------------
	commitPath := filepath.Join(".mrvc", "objects", head[:2], head[2:])
	commitBytes, err := os.ReadFile(commitPath)
	if err != nil {
		return "", err
	}

	var commit model.CommitObject
	if err := json.Unmarshal(commitBytes, &commit); err != nil {
		return "", err
	}

	// ------------------------------------------------------
	// Load HEAD tree
	// ------------------------------------------------------
	treeHash := commit.Tree
	treePath := filepath.Join(".mrvc", "objects", treeHash[:2], treeHash[2:])
	treeBytes, err := os.ReadFile(treePath)
	if err != nil {
		return "", err
	}

	var headTree model.TreeObject
	if err := json.Unmarshal(treeBytes, &headTree); err != nil {
		return "", err
	}

	// Convert HEAD tree to map path → hash
	headFiles := make(map[string]string)
	err = flattenTree(repoRoot, "", headTree, headFiles)
	if err != nil {
		return "", err
	}

	// ------------------------------------------------------
	// Scan working directory
	// ------------------------------------------------------
	workingFiles, err := fs.ListFiles(repoRoot, fs.WalkOptions{
		IgnoreMRVC:          true,
		IgnoreNestedRepos:   true,
		ApplyIgnorePatterns: true,
	})
	if err != nil {
		return "", err
	}

	// Normalize paths to match headFiles keys
	normalized := make([]string, 0, len(workingFiles))
	for _, f := range workingFiles {
		normalized = append(normalized, fs.NormalizePath(f))
	}

	// ------------------------------------------------------
	// Compare
	// ------------------------------------------------------
	var modified []string
	var deleted []string
	var untracked []string

	seen := make(map[string]bool)

	for _, w := range normalized {
		rel, _ := filepath.Rel(repoRoot, w)
		rel = filepath.ToSlash(rel)

		seen[rel] = true

		// In HEAD?
		headHash, exists := headFiles[rel]
		if !exists {
			untracked = append(untracked, rel)
			continue
		}

		// Compare content hash
		currentHash, err := fs.CalculateFileHash(w)
		if err != nil {
			return "", err
		}

		if currentHash != headHash {
			modified = append(modified, rel)
		}
	}

	// Deleted files: in HEAD but not in working dir
	for rel := range headFiles {
		if !seen[rel] {
			deleted = append(deleted, rel)
		}
	}

	// ------------------------------------------------------
	// Build output
	// ------------------------------------------------------
	var sb strings.Builder

	if len(modified) == 0 && len(deleted) == 0 && len(untracked) == 0 {
		return "clean", nil
	}

	if len(modified) > 0 {
		sb.WriteString("Modified:\n")
		for _, m := range modified {
			sb.WriteString("  " + m + "\n")
		}
		sb.WriteString("\n")
	}

	if len(deleted) > 0 {
		sb.WriteString("Deleted:\n")
		for _, d := range deleted {
			sb.WriteString("  " + d + "\n")
		}
		sb.WriteString("\n")
	}

	if len(untracked) > 0 {
		sb.WriteString("Untracked:\n")
		for _, u := range untracked {
			sb.WriteString("  " + u + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
