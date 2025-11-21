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
	return nil
}

func (v *VersionControlV1) Status() (string, error) {
	return "clean", nil
}
