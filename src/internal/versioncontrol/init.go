package versioncontrol

import (
	"fmt"
	"main/internal/util"
	"main/internal/versioncontrol/model"
	"path/filepath"
	"time"
)

func Init(repoName string, authorName string) error {
	currentDir, err := util.File.GetCurrentDir()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fvcDir := filepath.Join(currentDir, ".fvc")

	// Check dir existence
	exists, err := util.File.DoesDirExist(fvcDir)
	if err != nil {
		return fmt.Errorf("failed checking .fvc directory: %w", err)
	}
	if exists {
		return fmt.Errorf("repository already initialized at: %s", fvcDir)
	}

	// Create .fvc directory
	if err := util.File.CreateDir(fvcDir); err != nil {
		return fmt.Errorf("failed to create .fvc directory: %w", err)
	}

	// Create metadata.bin using FileUtil
	metadata := model.RepoMetadata{
		Name:      repoName,
		Author:    authorName,
		CreatedAt: time.Now().UnixMilli(),
	}
	metadataPath := filepath.Join(fvcDir, "metadata.bin")
	if err := util.File.WriteBinaryFile(metadataPath, metadata); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Create HEAD
	headPath := filepath.Join(fvcDir, "HEAD")
	if err := util.File.WriteTextFile(headPath, ""); err != nil {
		return fmt.Errorf("failed to write HEAD file: %w", err)
	}

	return nil
}
