package model

type Metadata struct {
	Name      string
	Author    string
	CreatedAt string
}

type TreeObject struct {
	entries []TreeEntry
}

type TreeEntry struct {
	name      string
	entryType string // "blob" or "tree"
	hash      string
}

type CommitObject struct {
	tree      string
	parent    string
	message   string
	author    string
	timestamp string
}
