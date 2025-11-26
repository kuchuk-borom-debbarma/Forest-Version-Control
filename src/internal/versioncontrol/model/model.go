package model

type RepoMetadata struct {
	Name      string
	Author    string
	CreatedAt int64
}

type DirectoryTree struct {
	Path    string
	Parent  string
	Entries []TreeEntry
}

type TreeEntry struct {
	Name string
	Type string
	Hash string
}
