package versioncontrol

import (
	"fmt"
	"main/internal/util"
	"path/filepath"
	"sort"
)

func Commit(message string, author string, files []string) error {
	if len(files) <= 0 {
		return fmt.Errorf("no files to commit")
	}
	if files[0] == "*" {
		//wild card :- all files in directory
	}
	var normalizedPath []string
	for _, f := range files {
		path, err := util.File.NormalizePath(f)
		if err != nil {
			return err
		}
		normalizedPath = append(normalizedPath, path)
	}
	//sort so that its naturally sorted path
	sort.Sort(sort.Reverse(sort.StringSlice(normalizedPath)))

	type FileAndHash struct {
		Name string
		Hash string
	}
	type DirData struct {
		Files        []FileAndHash
		ChildrenDirs []string
	}
	dirTree := make(map[string]DirData)

	// Step 1: Process all files and build directory entries
	for _, f := range normalizedPath {
		//Calculate hash and save object
		hashString, err := util.File.CalculateHashOfFile(f)
		if err != nil {
			return err
		}
		currentDir, err := util.File.GetCurrentDir()
		if err != nil {
			return err
		}
		util.File.SaveObject(hashString, f, filepath.Join(currentDir, ".fvc", "objects"))

		//Store the file in the dirTree map
		fileName := filepath.Base(f)
		fileDir := filepath.Dir(f)

		dirTreeData, exists := dirTree[fileDir]
		if !exists {
			dirTreeData = DirData{
				Files:        make([]FileAndHash, 0),
				ChildrenDirs: make([]string, 0),
			}
		}

		dirTreeData.Files = append(dirTreeData.Files, FileAndHash{
			Name: fileName,
			Hash: hashString,
		})

		dirTree[fileDir] = dirTreeData
	}

	// Step 2: Build parent-child relationships ONCE after all files processed
	for dir := range dirTree {
		parentDir := filepath.Dir(dir)

		if parentDir != "." && parentDir != dir {
			parentDirData, exists := dirTree[parentDir]
			if !exists {
				parentDirData = DirData{
					Files:        make([]FileAndHash, 0),
					ChildrenDirs: make([]string, 0),
				}
			}

			// Add child directory name (not full path)
			childName := filepath.Base(dir)
			// Check if not already added
			alreadyExists := false
			for _, existingChild := range parentDirData.ChildrenDirs {
				if existingChild == childName {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				parentDirData.ChildrenDirs = append(parentDirData.ChildrenDirs, childName)
			}

			dirTree[parentDir] = parentDirData
		}
	}

	//Step 3: Calculate the hashes of the trees, save it etc
	return nil
}
