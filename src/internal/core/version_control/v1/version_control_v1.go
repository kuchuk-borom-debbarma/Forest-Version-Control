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

func (v *VersionControlV1) Init(repoName string, author string) error {
	currentDir := fs.GetCurrentDir()
	repoDir := filepath.Join(currentDir, ".mrvc")

	log.Printf("Initializing MultiRepoVC %s, author %s on path %s",
		repoName, author, currentDir)

	// 1. Check if repo exists
	if fs.IsDirPresent(repoDir) {
		return errors.New("repository already initialized")
	}

	// 2. Create .mrvc
	if err := fs.CreateDir(repoDir); err != nil {
		return err
	}

	// 3. Create metadata
	metadata := model.Metadata{
		Name:      repoName,
		Author:    author,
		CreatedAt: strconv.FormatInt(time.GetCurrentTimestamp(), 10),
	}

	// 4. Write metadata JSON
	if err := fs.WriteJSON(filepath.Join(repoDir, "metadata.json"), metadata); err != nil {
		return err
	}

	return nil
}

func (v *VersionControlV1) Commit(message string, author string, files []string) error {
	if len(files) == 0 {
		return errors.New("no files to commit")
	}

	repoRoot := fs.GetCurrentDir()

	// ------------------------------
	// WILDCARD HANDLING
	// ------------------------------
	if len(files) == 1 && files[0] == "*" {
		log.Println("Wildcard commit: all non-ignored files")

		allFiles, err := fs.ListFilesExcludingIgnore(repoRoot)
		if err != nil {
			return err
		}

		normalized := make([]string, 0, len(allFiles))
		for _, f := range allFiles {
			normalized = append(normalized, fs.NormalizePath(f))
		}
		files = normalized
	} else {
		// Validate user-specified files
		for i, f := range files {
			files[i] = fs.NormalizePath(f)
			if !fs.FileExists(files[i]) {
				return errors.New("file does not exist: " + f)
			}
		}
	}

	log.Println("Committing files:", files)

	// --------------------------------------------------------------------
	// DIRECTORY TREE MAP
	// Key: absolute normalized directory path
	// Value: TreeObject of that directory
	// --------------------------------------------------------------------
	directoryTrees := make(map[string]model.TreeObject)

	// --------------------------------------------------------------------
	// PROCESS EACH FILE → CREATE BLOBS AND TREE ENTRIES
	// --------------------------------------------------------------------
	for _, filePath := range files {

		// 1. Calculate hash
		hash, err := fs.CalculateFileHash(filePath)
		if err != nil {
			return err
		}

		// 2. Save blob object
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if err := SaveObject(hash, content); err != nil {
			return err
		}

		// 3. Resolve directory path
		fileDir := filepath.Dir(filePath)
		if fileDir == "." {
			fileDir = repoRoot
		}
		fileDir = fs.NormalizePath(fileDir)

		// 4. Ensure directoryTrees entry
		tree := directoryTrees[fileDir]
		tree = addOrReplaceTreeEntry(tree, model.TreeEntry{
			Name:      filepath.Base(filePath),
			EntryType: "blob",
			Hash:      hash,
		})
		directoryTrees[fileDir] = tree

		// ----------------------------------------------------------------
		// 5. Recursively walk up to the repo root, building empty trees
		//    So if file is at: /root/a/b/c.txt
		//    we ensure directories: /root/a/b, /root/a, /root
		// ----------------------------------------------------------------
		current := fileDir
		for current != repoRoot {
			parent := filepath.Dir(current)
			if parent == "." {
				parent = repoRoot
			}
			parent = fs.NormalizePath(parent)

			// ensure parent exists
			if _, ok := directoryTrees[parent]; !ok {
				directoryTrees[parent] = model.TreeObject{Entries: []model.TreeEntry{}}
			}

			current = parent
		}
	}

	// ------------------------------------------------------------
	// BUILD TREE HIERARCHY — bottom up
	// ------------------------------------------------------------

	// Sorting keys ensures deterministic tree builds
	var dirs []string
	for d := range directoryTrees {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	treeHashes := make(map[string]string)

	// Process directories deepest-first
	sort.Slice(dirs, func(i, j int) bool {
		return strings.Count(dirs[i], "/") > strings.Count(dirs[j], "/")
	})

	for _, dir := range dirs {
		tree := directoryTrees[dir]

		// Add child directory entries
		for _, childDir := range dirs {
			if filepath.Dir(childDir) == dir {
				// add child tree entry
				hash := treeHashes[childDir]
				tree = addOrReplaceTreeEntry(tree, model.TreeEntry{
					Name:      filepath.Base(childDir),
					EntryType: "tree",
					Hash:      hash,
				})
			}
		}

		// Hash final directory tree
		jsonBytes, _ := json.Marshal(tree)
		treeHash := sha256Hex(jsonBytes)
		SaveObject(treeHash, jsonBytes)

		treeHashes[dir] = treeHash
	}

	// root tree is the repo root
	rootTreeHash := treeHashes[repoRoot]

	// ------------------------------------------------------------
	// CREATE COMMIT
	// ------------------------------------------------------------
	commit := model.CommitObject{
		Tree:      rootTreeHash,
		Parent:    strings.TrimSpace(readHEAD()),
		Message:   message,
		Author:    author,
		Timestamp: strconv.FormatInt(time.GetCurrentTimestamp(), 10),
	}

	data, _ := json.Marshal(commit)
	commitHash := sha256Hex(data)
	SaveObject(commitHash, data)

	// Update HEAD
	updateHEAD(commitHash)

	log.Println("Commit created:", commitHash)
	return nil
}

func (v *VersionControlV1) Status() (string, error) {
	return "clean", nil
}
