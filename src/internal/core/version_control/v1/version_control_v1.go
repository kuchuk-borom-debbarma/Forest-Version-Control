package v1

import (
	"MultiRepoVC/src/internal/core/version_control/v1/model"
	"MultiRepoVC/src/internal/utils/fs"
	"MultiRepoVC/src/internal/utils/time"
	"errors"
	"log"
	"path/filepath"
	"strconv"
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

	// Wildcard commit
	if len(files) == 1 && files[0] == "*" {
		log.Println("Wildcard commit detected. Tracking all files excluding the ones in .mrvcignore")

		allFiles, err := fs.ListFilesExcludingIgnore(fs.GetCurrentDir())
		if err != nil {
			return err
		}

		// Normalize paths
		normalized := make([]string, 0, len(allFiles))
		for _, f := range allFiles {
			normalized = append(normalized, fs.NormalizePath(f))
		}

		files = normalized

	} else {
		// validate files exist (with normalization)
		for i, f := range files {
			norm := fs.NormalizePath(f)
			files[i] = norm

			if !fs.FileExists(norm) {
				return errors.New("file does not exist: " + f)
			}
		}
	}
	log.Println("Committing files:", files)

	// For each file:- 1. Store its directory + parent directory, store the hash and content, save them and create the blob and tree. Also track the root tree i.e root dir

	return nil
}

func (v *VersionControlV1) Status() (string, error) {
	return "clean", nil
}
