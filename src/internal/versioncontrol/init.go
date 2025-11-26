package versioncontrol

import (
	"encoding/gob"
	"fmt"
	"main/internal/versioncontrol/model"
	"os"
	"path/filepath"
	"time"
)

func Init(repoName string, authorName string) error {
	/*
		Overview:
		- Create .fvc directory (error if already exists)
		- Create metadata.bin containing repo name, author, timestamp
		- Create HEAD file with empty content
	*/

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fvcDir := filepath.Join(currentDir, ".fvc")

	// Correct existence check:
	if _, err := os.Stat(fvcDir); err == nil {
		return fmt.Errorf("repository already initialized at: %s", fvcDir)
	}

	// Create .fvc directory
	if err := os.Mkdir(fvcDir, 0755); err != nil {
		return fmt.Errorf("failed to create .fvc directory: %w", err)
	}

	// Create metadata
	metadata := model.RepoMetadata{
		Name:      repoName,
		Author:    authorName,
		CreatedAt: time.Now().UnixMilli(),
	}

	metadataPath := filepath.Join(fvcDir, "metadata.bin")
	metadataFile, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer metadataFile.Close()

	encoder := gob.NewEncoder(metadataFile)
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	// Create empty head
	headPath := filepath.Join(fvcDir, "HEAD")
	headFile, err := os.Create(headPath)
	if err != nil {
		return fmt.Errorf("failed to create HEAD file: %w", err)
	}
	defer headFile.Close()

	// Write empty content (no commit yet)
	if _, err := headFile.WriteString(""); err != nil {
		return fmt.Errorf("failed to write to HEAD file: %w", err)
	}

	return nil
}
