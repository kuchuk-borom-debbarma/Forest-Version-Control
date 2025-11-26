package versioncontrol

import (
	"fmt"
	"main/internal/util"
	"main/internal/versioncontrol/model"
	"path/filepath"
	"sort"
	"time"
)

func Commit(message string, author string, files []string) error {

	// --------------------------------------------
	// 1. Basic validation
	// --------------------------------------------
	if len(files) == 0 {
		return fmt.Errorf("no files to commit")
	}

	// --------------------------------------------
	// 2. Normalize file paths
	// We convert "src//file.go" → "src/file.go"
	// This ensures consistent directory grouping & hashing
	// --------------------------------------------
	var normalizedPaths []string
	for _, f := range files {
		p, err := util.File.NormalizePath(f)
		if err != nil {
			return fmt.Errorf("failed normalizing path %s: %w", f, err)
		}
		normalizedPaths = append(normalizedPaths, p)
	}

	// Sorting ensures deterministic tree structure
	// (Two commits with same files must produce same hash)
	sort.Strings(normalizedPaths)

	// --------------------------------------------
	// 3. Directory tree structure
	//
	//	dirTree["src/main"] = {
	//	    Files:        [{Name:"a.go", Hash:"xxx"}],
	//	    ChildrenDirs: ["utils"]
	//	}
	//
	// This structure lets us later compute tree hashes
	// from the deepest directory → upwards.
	// --------------------------------------------
	type FileAndHash struct {
		Name string
		Hash string
	}

	type DirData struct {
		Files        []FileAndHash
		ChildrenDirs []string
	}

	dirTree := make(map[string]DirData)

	cwd, _ := util.File.GetCurrentDir()
	objectsDir := filepath.Join(cwd, ".fvc", "objects")

	// --------------------------------------------
	// 4. Process files: hash them + store blob objects
	//
	// Also fill dirTree with file entries.
	// --------------------------------------------
	for _, path := range normalizedPaths {

		// Step A: compute file hash
		hash, err := util.File.CalculateHashOfFile(path)
		if err != nil {
			return fmt.Errorf("cannot hash file %s: %w", path, err)
		}

		// Step B: store file in .fvc/objects/xx/yyyy object style
		if err := util.File.SaveObject(hash, path, objectsDir); err != nil {
			return fmt.Errorf("cannot save object %s: %w", path, err)
		}

		// Step C: add file into directory entry
		fileName := filepath.Base(path)
		dirName := filepath.Dir(path)

		d := dirTree[dirName]
		d.Files = append(d.Files, FileAndHash{Name: fileName, Hash: hash})
		dirTree[dirName] = d
	}

	// --------------------------------------------
	// 5. Build parent → child directory relationships
	//
	// Example:
	// If we have dir "src/main/utils"
	// Then parent is "src/main"
	//
	// We do this *after* processing all files.
	// --------------------------------------------
	for dir := range dirTree {
		parent := filepath.Dir(dir)

		// Skip root
		if parent == "." || parent == dir {
			continue
		}

		parentData := dirTree[parent]
		childName := filepath.Base(dir)

		// Add child directory if not already present
		exists := false
		for _, c := range parentData.ChildrenDirs {
			if c == childName {
				exists = true
				break
			}
		}
		if !exists {
			parentData.ChildrenDirs = append(parentData.ChildrenDirs, childName)
		}

		dirTree[parent] = parentData
	}

	// --------------------------------------------
	// 6. Compute tree hashes (MOST IMPORTANT PART)
	//
	// We need to hash children → before → parents.
	// Example:
	// utils/ → hashA
	// main/  → depends on utils hash → hashB
	// src/   → depends on main hash → hashC
	//
	// Solution:
	// Sort directory paths by depth (longest → first)
	// --------------------------------------------
	dirs := make([]string, 0, len(dirTree))
	for dir := range dirTree {
		dirs = append(dirs, dir)
	}

	// deeper directories have longer paths
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	treeHash := make(map[string]string)

	// --------------------------------------------
	// 7. Build and hash each directory (bottom-up)
	// --------------------------------------------
	for _, dir := range dirs {
		data := dirTree[dir]

		// Create a tree object in memory
		tree := model.DirectoryTree{
			Path:    filepath.Base(dir),
			Parent:  filepath.Dir(dir),
			Entries: make([]model.TreeEntry, 0),
		}

		// Add file entries (blob)
		for _, f := range data.Files {
			tree.Entries = append(tree.Entries, model.TreeEntry{
				Name: f.Name,
				Type: "blob",
				Hash: f.Hash,
			})
		}

		// Add subdirectory entries (tree)
		for _, child := range data.ChildrenDirs {
			childFullPath := filepath.Join(dir, child)
			tree.Entries = append(tree.Entries, model.TreeEntry{
				Name: child,
				Type: "tree",
				Hash: treeHash[childFullPath],
			})
		}

		// Save tree object to temp file so we can hash its contents
		tempTree := filepath.Join(objectsDir, "tmp_tree")
		if err := util.File.WriteBinaryFile(tempTree, tree); err != nil {
			return err
		}

		// Hash the tree object
		hash, err := util.File.CalculateHashOfFile(tempTree)
		if err != nil {
			return err
		}

		// Save tree object with its hash
		if err := util.File.SaveObject(hash, tempTree, objectsDir); err != nil {
			return err
		}

		// Store hash for parent directories to use
		treeHash[dir] = hash
	}

	// --------------------------------------------
	// 8. Commit object → references only the ROOT tree
	// Root is the directory of the first file
	// --------------------------------------------
	rootDir := filepath.Dir(normalizedPaths[0])
	rootTreeHash := treeHash[rootDir]

	commit := model.Commit{
		Message:  message,
		Author:   author,
		Date:     time.Now().UnixMilli(),
		TreeHash: rootTreeHash,
	}

	// Encode commit
	tempCommit := filepath.Join(objectsDir, "tmp_commit")
	if err := util.File.WriteBinaryFile(tempCommit, commit); err != nil {
		return err
	}

	// Hash commit object
	commitHash, err := util.File.CalculateHashOfFile(tempCommit)
	if err != nil {
		return err
	}

	// Save commit object
	if err := util.File.SaveObject(commitHash, tempCommit, objectsDir); err != nil {
		return err
	}

	// --------------------------------------------
	// 9. Update HEAD
	// --------------------------------------------
	headPath := filepath.Join(cwd, ".fvc", "HEAD")
	if err := util.File.WriteTextFile(headPath, commitHash); err != nil {
		return err
	}

	fmt.Println("Committed as", commitHash)
	return nil
}
