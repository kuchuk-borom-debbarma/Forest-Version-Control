package constants

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"MultiRepoVC/src/internal/core/version_control/v1/model"
	"MultiRepoVC/src/internal/utils/fs"
	"MultiRepoVC/src/internal/utils/time"
)

type VersionControlV1 struct{}

func New() *VersionControlV1 {
	return &VersionControlV1{}
}

// ======================================================================
// INIT
// ======================================================================

func (v *VersionControlV1) Init(repoName string, author string) error {
	root := fs.GetCurrentDir()
	mrvc := filepath.Join(root, ".mrvc")

	if fs.IsDirPresent(mrvc) {
		return errors.New("repository already initialized")
	}

	if err := fs.CreateDir(mrvc); err != nil {
		return err
	}

	meta := model.Metadata{
		Name:      repoName,
		Author:    author,
		CreatedAt: strconv.FormatInt(time.GetCurrentTimestamp(), 10),
	}

	return fs.WriteJSON(filepath.Join(mrvc, "metadata.json"), meta)
}

// ======================================================================
// COMMIT
// ======================================================================

func (v *VersionControlV1) Commit(message string, author string, files []string) error {
	if len(files) == 0 {
		return errors.New("no files to commit")
	}

	repoRoot := fs.GetCurrentDir()

	// Expand wildcard
	if len(files) == 1 && files[0] == "*" {
		all, err := fs.ListFiles(repoRoot, fs.WalkOptions{
			IgnoreMRVC:          true,
			IgnoreNestedRepos:   true,
			ApplyIgnorePatterns: true,
		})
		if err != nil {
			return fmt.Errorf("failed to list files for wildcard commit: %w", err)
		}
		files = make([]string, 0, len(all))
		for _, f := range all {
			files = append(files, fs.NormalizePath(f))
		}
	} else {
		for i, f := range files {
			n := fs.NormalizePath(f)
			if !fs.FileExists(n) {
				return fmt.Errorf("file does not exist: %s", n)
			}
			files[i] = n
		}
	}

	// Build tree objects
	directoryTrees := make(map[string]model.TreeObject)
	children := make(map[string][]string)

	for _, filePath := range files {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		blobHash := HashContent(content)
		if err := SaveObject(blobHash, content); err != nil {
			return fmt.Errorf("failed to save blob object: %w", err)
		}

		fileDir := filepath.Dir(filePath)
		if fileDir == "." {
			fileDir = repoRoot
		}
		fileDir = fs.NormalizePath(fileDir)

		if _, exists := directoryTrees[fileDir]; !exists {
			directoryTrees[fileDir] = model.TreeObject{Entries: []model.TreeEntry{}}
		}

		tree := directoryTrees[fileDir]
		tree = addOrReplaceTreeEntry(tree, model.TreeEntry{
			Name:      filepath.Base(filePath),
			EntryType: "blob",
			Hash:      blobHash,
		})
		directoryTrees[fileDir] = tree

		// ensure parent directories are present
		current := fileDir
		for current != repoRoot {
			parent := filepath.Dir(current)
			if parent == "." {
				parent = repoRoot
			}
			parent = fs.NormalizePath(parent)
			if _, ok := directoryTrees[parent]; !ok {
				directoryTrees[parent] = model.TreeObject{Entries: []model.TreeEntry{}}
			}
			children[parent] = append(children[parent], current)
			current = parent
		}
	}

	// Sort directories deepest -> shallowest
	var dirs []string
	for d := range directoryTrees {
		dirs = append(dirs, d)
	}
	sort.Slice(dirs, func(i, j int) bool {
		return strings.Count(dirs[i], "/") > strings.Count(dirs[j], "/")
	})

	treeHashes := make(map[string]string)
	for _, dir := range dirs {
		tree := directoryTrees[dir]
		for _, child := range children[dir] {
			tree = addOrReplaceTreeEntry(tree, model.TreeEntry{
				Name:      filepath.Base(child),
				EntryType: "tree",
				Hash:      treeHashes[child],
			})
		}
		// deterministic ordering
		sort.Slice(tree.Entries, func(i, j int) bool {
			return tree.Entries[i].Name < tree.Entries[j].Name
		})
		hash, jsonBytes, err := HashTree(tree)
		if err != nil {
			return fmt.Errorf("failed to hash tree for %s: %w", dir, err)
		}
		if err := SaveObject(hash, jsonBytes); err != nil {
			return fmt.Errorf("failed to save tree object: %w", err)
		}
		treeHashes[dir] = hash
	}

	rootTreeHash, ok := treeHashes[repoRoot]
	if !ok {
		// If nothing was committed, create an empty tree
		emptyTree := model.TreeObject{Entries: []model.TreeEntry{}}
		hash, jsonBytes, err := HashTree(emptyTree)
		if err != nil {
			return fmt.Errorf("failed to hash empty tree: %w", err)
		}
		if err := SaveObject(hash, jsonBytes); err != nil {
			return fmt.Errorf("failed to save empty tree: %w", err)
		}
		rootTreeHash = hash
	}

	commit := model.CommitObject{
		Tree:      rootTreeHash,
		Parent:    readHEAD(),
		Message:   message,
		Author:    author,
		Timestamp: strconv.FormatInt(time.GetCurrentTimestamp(), 10),
	}

	commitHash, commitBytes, err := HashCommit(commit)
	if err != nil {
		return fmt.Errorf("failed to hash commit: %w", err)
	}
	if err := SaveObject(commitHash, commitBytes); err != nil {
		return fmt.Errorf("failed to save commit object: %w", err)
	}
	if err := updateHEAD(commitHash); err != nil {
		return fmt.Errorf("failed to update HEAD: %w", err)
	}

	log.Println("Commit created:", commitHash)
	return nil
}

// ======================================================================
// STATUS
// ======================================================================

func (v *VersionControlV1) Status() (string, error) {
	repoRoot := fs.GetCurrentDir()
	var sb strings.Builder

	// SECTION 1 - Normal status
	head := readHEAD()
	if head == "" {
		sb.WriteString("No commits yet.\n")
	} else {
		normalStatus, err := v.statusNormal(repoRoot, head)
		if err != nil {
			return "", fmt.Errorf("failed to compute normal status: %w", err)
		}
		sb.WriteString(normalStatus)
		sb.WriteString("\n")
	}

	// SECTION 2 - Super commit status
	headSuper := readHEADSUPER()
	if headSuper == "" {
		sb.WriteString("No super commits yet.\n")
		return sb.String(), nil
	}

	sb.WriteString(fmt.Sprintf("Super Commit: %s\n", headSuper))

	// Load super commit object
	objPath := filepath.Join(".mrvc", "objects", headSuper[:2], headSuper[2:])
	data, err := os.ReadFile(objPath)
	if err != nil {
		sb.WriteString(fmt.Sprintf("ERROR: Failed to load super commit object %s: %v\n", headSuper, err))
		return sb.String(), nil
	}
	var sc model.SuperCommitObject
	if err := json.Unmarshal(data, &sc); err != nil {
		sb.WriteString(fmt.Sprintf("ERROR: Corrupt super commit object: %v\n", err))
		return sb.String(), nil
	}

	sb.WriteString("\nSelf snapshot:\n")
	sb.WriteString(fmt.Sprintf("  commit: %s\n\n", sc.SelfHead))

	// Load working children.json (best-effort)
	childrenPath := filepath.Join(repoRoot, ".mrvc", childrenFileName)
	cf := model.ChildrenFile{Children: []model.ChildEntry{}}
	if fs.FileExists(childrenPath) {
		if err := fs.ReadJSON(childrenPath, &cf); err != nil {
			sb.WriteString("  ERROR: failed to read children.json\n")
			return sb.String(), nil
		}
	}

	// create map from repoName -> ChildEntry for quick lookup
	childMap := make(map[string]model.ChildEntry)
	for _, c := range cf.Children {
		childMap[c.RepoName] = c
	}

	sb.WriteString("Children:\n")
	for _, ch := range sc.Children {
		sb.WriteString(fmt.Sprintf("  %s (path: %s)\n", ch.RepoName, ch.Path))

		entry, ok := childMap[ch.RepoName]
		if !ok {
			sb.WriteString("    ✗ ERROR: child repo missing from children.json\n\n")
			continue
		}

		childAbs := filepath.Join(repoRoot, entry.Path)
		if !fs.IsDirPresent(childAbs) {
			sb.WriteString("    ✗ ERROR: child directory missing\n\n")
			continue
		}

		childMrvc := filepath.Join(childAbs, ".mrvc")
		metaPath := filepath.Join(childMrvc, "metadata.json")
		if !fs.FileExists(metaPath) {
			sb.WriteString("    ✗ ERROR: missing metadata.json in child\n\n")
			continue
		}

		var meta model.Metadata
		if err := fs.ReadJSON(metaPath, &meta); err != nil {
			sb.WriteString("    ✗ ERROR: failed reading child's metadata.json\n\n")
			continue
		}

		if meta.Name != ch.RepoName {
			sb.WriteString(fmt.Sprintf("    ✗ ERROR: repo identity mismatch (expected=%s, found=%s)\n\n", ch.RepoName, meta.Name))
			continue
		}

		// Make sure referenced object exists inside child .mrvc/objects
		refObj := filepath.Join(childMrvc, "objects", ch.Ref[:2], ch.Ref[2:])
		if !fs.FileExists(refObj) {
			sb.WriteString(fmt.Sprintf("    ✗ ERROR: referenced object missing: %s\n\n", ch.Ref))
			continue
		}

		sb.WriteString(fmt.Sprintf("    ✓ type: %s\n", ch.Type))
		sb.WriteString(fmt.Sprintf("    ✓ ref:  %s\n", ch.Ref))
		if ch.Type == "commit" {
			sb.WriteString("    ⚠ child has no super commit yet\n")
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func (v *VersionControlV1) statusNormal(repoRoot, head string) (string, error) {
	// head is expected non-empty
	if len(head) < 2 {
		return "", errors.New("invalid HEAD value")
	}

	// Load commit object
	commitPath := filepath.Join(".mrvc", "objects", head[:2], head[2:])
	commitBytes, err := os.ReadFile(commitPath)
	if err != nil {
		return "", fmt.Errorf("failed to read HEAD commit object: %w", err)
	}
	var commit model.CommitObject
	if err := json.Unmarshal(commitBytes, &commit); err != nil {
		return "", fmt.Errorf("failed to parse HEAD commit object: %w", err)
	}

	// Load tree object
	treeHash := commit.Tree
	if len(treeHash) < 2 {
		return "", errors.New("invalid tree hash in commit")
	}
	treePath := filepath.Join(".mrvc", "objects", treeHash[:2], treeHash[2:])
	treeBytes, err := os.ReadFile(treePath)
	if err != nil {
		return "", fmt.Errorf("failed to read tree object: %w", err)
	}
	var headTree model.TreeObject
	if err := json.Unmarshal(treeBytes, &headTree); err != nil {
		return "", fmt.Errorf("failed to parse tree object: %w", err)
	}

	headFiles := make(map[string]string)
	if err := flattenTree(repoRoot, "", headTree, headFiles); err != nil {
		return "", fmt.Errorf("failed to flatten tree: %w", err)
	}

	workingFiles, err := fs.ListFiles(repoRoot, fs.WalkOptions{
		IgnoreMRVC:          true,
		IgnoreNestedRepos:   true,
		ApplyIgnorePatterns: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to scan working directory: %w", err)
	}

	norm := make([]string, 0, len(workingFiles))
	for _, f := range workingFiles {
		norm = append(norm, fs.NormalizePath(f))
	}

	var modified []string
	var deleted []string
	var untracked []string
	seen := make(map[string]bool, len(norm))

	for _, w := range norm {
		rel, _ := filepath.Rel(repoRoot, w)
		rel = filepath.ToSlash(rel)
		seen[rel] = true

		headHash, ok := headFiles[rel]
		if !ok {
			untracked = append(untracked, rel)
			continue
		}

		currentHash, err := fs.CalculateFileHash(w)
		if err != nil {
			return "", fmt.Errorf("failed to hash working file %s: %w", w, err)
		}

		if currentHash != headHash {
			modified = append(modified, rel)
		}
	}

	for rel := range headFiles {
		if !seen[rel] {
			deleted = append(deleted, rel)
		}
	}

	// Build status string
	var sb strings.Builder
	if len(modified) == 0 && len(deleted) == 0 && len(untracked) == 0 {
		sb.WriteString("clean")
		return sb.String(), nil
	}

	if len(modified) > 0 {
		sb.WriteString("Modified:\n")
		for _, m := range modified {
			sb.WriteString("  " + m + "\n")
		}
		sb.WriteString("\n")
	}
	if len(deleted) > 0 {
		sb.WriteString("Deleted:\n")
		for _, d := range deleted {
			sb.WriteString("  " + d + "\n")
		}
		sb.WriteString("\n")
	}
	if len(untracked) > 0 {
		sb.WriteString("Untracked:\n")
		for _, u := range untracked {
			sb.WriteString("  " + u + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// ======================================================================
// LINK
// ======================================================================

func (v *VersionControlV1) Link(childPath string) error {
	if childPath == "" {
		return errors.New("child path cannot be empty")
	}
	childPath = fs.NormalizePath(childPath)

	parentRoot := fs.GetCurrentDir()
	childAbs, err := filepath.Abs(childPath)
	if err != nil {
		return fmt.Errorf("unable to resolve absolute path for '%s': %w", childPath, err)
	}

	childMrvc := filepath.Join(childAbs, ".mrvc")
	if !fs.IsDirPresent(childMrvc) {
		return fmt.Errorf("'%s' is not an MRVC repository (missing .mrvc directory)", childAbs)
	}

	metaPath := filepath.Join(childMrvc, "metadata.json")
	if !fs.FileExists(metaPath) {
		return fmt.Errorf("'%s' is not a valid MRVC repository (missing metadata.json)", childAbs)
	}

	var meta model.Metadata
	if err := fs.ReadJSON(metaPath, &meta); err != nil {
		return fmt.Errorf("failed to read child metadata.json: %w", err)
	}

	childRel, err := filepath.Rel(parentRoot, childAbs)
	if err != nil {
		return fmt.Errorf("unable to calculate relative path: %w", err)
	}
	childRel = filepath.ToSlash(childRel)
	if strings.HasPrefix(childRel, "../") {
		return fmt.Errorf("child repo '%s' must be inside parent repo '%s'", childRel, parentRoot)
	}

	childrenPath := filepath.Join(parentRoot, ".mrvc", childrenFileName)
	cf := model.ChildrenFile{Children: []model.ChildEntry{}}
	if fs.FileExists(childrenPath) {
		if err := fs.ReadJSON(childrenPath, &cf); err != nil {
			return fmt.Errorf("invalid children.json: %w", err)
		}
	}

	for _, existing := range cf.Children {
		if existing.Path == childRel {
			return fmt.Errorf("child path '%s' is already linked", childRel)
		}
		if existing.RepoName == meta.Name {
			return fmt.Errorf("child repoName '%s' is already linked", meta.Name)
		}
	}

	newEntry := model.ChildEntry{
		Path:     childRel,
		RepoName: meta.Name,
	}
	cf.Children = append(cf.Children, newEntry)

	if err := fs.WriteJSON(childrenPath, cf); err != nil {
		return fmt.Errorf("failed to save children.json: %w", err)
	}

	return nil
}

// ======================================================================
// SUPER COMMIT
// ======================================================================

func (v *VersionControlV1) SuperCommit(message string, author string) error {
	repoRoot := fs.GetCurrentDir()

	selfHead := readHEAD()
	if selfHead == "" {
		return errors.New("cannot create super commit: no commits yet in this repo")
	}

	childrenPath := filepath.Join(repoRoot, ".mrvc", childrenFileName)
	cf := model.ChildrenFile{Children: []model.ChildEntry{}}
	if fs.FileExists(childrenPath) {
		if err := fs.ReadJSON(childrenPath, &cf); err != nil {
			return fmt.Errorf("invalid children.json: %w", err)
		}
	}

	childRefs := make([]model.SuperCommitChildRef, 0, len(cf.Children))
	for _, child := range cf.Children {
		childAbs := filepath.Join(repoRoot, child.Path)
		childMrvc := filepath.Join(childAbs, ".mrvc")

		if !fs.IsDirPresent(childMrvc) {
			return fmt.Errorf("child repo missing or corrupted: %s", child.Path)
		}

		var meta model.Metadata
		if err := fs.ReadJSON(filepath.Join(childMrvc, "metadata.json"), &meta); err != nil {
			return fmt.Errorf("failed to read metadata for child '%s': %w", child.Path, err)
		}

		if meta.Name != child.RepoName {
			return fmt.Errorf("repo identity mismatch for '%s': expected '%s', found '%s'",
				child.Path, child.RepoName, meta.Name)
		}

		childHead := func() string {
			b, err := os.ReadFile(filepath.Join(childMrvc, "HEAD"))
			if err != nil {
				return ""
			}
			return strings.TrimSpace(string(b))
		}()

		childSuper := func() string {
			b, err := os.ReadFile(filepath.Join(childMrvc, "HEAD_SUPER"))
			if err != nil {
				return ""
			}
			return strings.TrimSpace(string(b))
		}()

		if childHead == "" {
			return fmt.Errorf("child repo '%s' has no commits", child.Path)
		}

		ref := childHead
		refType := "commit"
		if childSuper != "" {
			ref = childSuper
			refType = "super"
		}

		childRefs = append(childRefs, model.SuperCommitChildRef{
			Path:     child.Path,
			RepoName: child.RepoName,
			Ref:      ref,
			Type:     refType,
		})
	}

	superObj := model.SuperCommitObject{
		SelfHead:  selfHead,
		Children:  childRefs,
		Message:   message,
		Author:    author,
		Timestamp: strconv.FormatInt(time.GetCurrentTimestamp(), 10),
	}

	hash, jsonBytes, err := HashSuperCommit(superObj)
	if err != nil {
		return fmt.Errorf("failed to hash super commit: %w", err)
	}
	if err := SaveObject(hash, jsonBytes); err != nil {
		return fmt.Errorf("failed to save super commit object: %w", err)
	}

	if err := updateHEADSUPER(hash); err != nil {
		return fmt.Errorf("failed to update HEAD_SUPER: %w", err)
	}

	log.Println("Super commit created:", hash)
	return nil
}

// ======================================================================
// Helpers
// ======================================================================

func readHEAD() string {
	b, err := os.ReadFile(filepath.Join(".mrvc", "HEAD"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func updateHEAD(hash string) error {
	if hash == "" {
		return errors.New("empty hash")
	}
	return os.WriteFile(filepath.Join(".mrvc", "HEAD"), []byte(strings.TrimSpace(hash)), 0644)
}

func readHEADSUPER() string {
	b, err := os.ReadFile(filepath.Join(".mrvc", "HEAD_SUPER"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func updateHEADSUPER(hash string) error {
	if hash == "" {
		return errors.New("empty hash")
	}
	return os.WriteFile(filepath.Join(".mrvc", "HEAD_SUPER"), []byte(strings.TrimSpace(hash)), 0644)
}
