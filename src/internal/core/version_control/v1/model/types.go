package model

type Metadata struct {
	Name      string `json:"name"`
	Author    string `json:"author"`
	CreatedAt string `json:"created_at"`
}

// TREE ---------------------------------------------------------------------

type TreeObject struct {
	Entries []TreeEntry `json:"entries"`
}

type TreeEntry struct {
	Name      string `json:"name"`
	EntryType string `json:"entry_type"` // "blob" or "tree"
	Hash      string `json:"hash"`
}

// COMMIT --------------------------------------------------------------------

type CommitObject struct {
	Tree      string `json:"tree"`
	Parent    string `json:"parent"`
	Message   string `json:"message"`
	Author    string `json:"author"`
	Timestamp string `json:"timestamp"`
}

type ChildEntry struct {
	Path     string `json:"path"`
	RepoName string `json:"repoName"`
}

type ChildrenFile struct {
	Children []ChildEntry `json:"children"`
}

type SuperCommitChildRef struct {
	Path     string `json:"path"`
	RepoName string `json:"repoName"`
	Ref      string `json:"ref"`
	Type     string `json:"type"` // "super" or "commit"
}

type SuperCommitObject struct {
	SelfHead  string                `json:"self_head"`
	Children  []SuperCommitChildRef `json:"children"`
	Message   string                `json:"message"`
	Author    string                `json:"author"`
	Timestamp string                `json:"timestamp"`
}
