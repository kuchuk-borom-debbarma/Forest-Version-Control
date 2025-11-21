
# 📘 MultiRepo VC — Architecture & Roadmap

*A Snapshot-Based Version Control System With Nested Repository Support*

---

# 📍 Roadmap Overview

MultiRepo VC is built on a **snapshot-per-commit** model, where each commit represents a complete snapshot of the project directory at that point in time.

The overall project roadmap:

1. **Implement Commit Snapshot Model**

    * Blob objects
    * Tree objects
    * Commit objects
    * Content-addressed storage backend
    * Selective commit (no staging area)

2. **Implement Nested Repository Support**

    * Allow repositories inside repositories
    * Maintain tree isolation and reference semantics
    * Provide controlled linking of sub-repos

3. **Implement Optimization Layer**

    * Optional diff storage (delta compression)
    * Periodic snapshot packing (similar to Git packfiles)
    * History compaction/cleanup

This layered approach ensures a solid foundation before advanced optimizations.

---

# 🧭 Design Philosophy

MultiRepo VC follows three core principles:

### 1. **Immutable Objects**

Every blob, tree, and commit is content-addressed and immutable.

> Changing the content produces a new hash → new object → new snapshot.

### 2. **Staging-less Workflow**

There is **no staging area**.
Commits operate directly on the **working directory**, making the system easier to reason about.

Users can:

* commit everything (`mrvc commit .`)
* commit **specific** files or patterns (`mrvc commit README.md src/*.go docs/**`)

### 3. **Directory Snapshot Trees**

Each commit stores a **tree of trees** representing the entire repository structure.

---

# 📝 Repository Initialization

When initializing a new repository:

### 📂 `.mrvc/` folder is created:

```
.mrvc/
  HEAD
  metadata.json
  objects/
```

### 📌 Components

#### **`metadata.json`**

```json
{
  "repoName": "MyProject",
  "createdAt": 1732211000,
  "formatVersion": 1
}
```

#### **`HEAD`**

```
commit: null
```

Repository starts with no commits.

#### **`objects/`**

Initially empty.
Later stores:

* blob objects
* tree objects
* commit objects

Each stored using a **split-path hash layout**:

```
objects/ac/42f3d1e3ab...
```

---

# 📦 Blob Objects

A **blob** holds file contents exactly as-is.

### Storage path:

```
objects/ac/42f3...
```

### Blob representation:

```
<raw file bytes>
```

### Hashing:

```
sha256(file bytes)
```

---

# 🌳 Tree Objects

A **tree** describes one directory.

### A tree contains:

* file → blob reference
* subdirectory → tree reference
* future: nested repository reference

### Example:

```json
{
  "type": "tree",
  "entries": [
    { "name": "README.md", "type": "blob", "hash": "a1b2..." },
    { "name": "src", "type": "tree", "hash": "d4e5..." }
  ]
}
```

### Hashing:

1. Serialize tree in stable format
2. Compute SHA256
3. Store under `.mrvc/objects/<hash>`

---

# 🔗 Commit Objects

A commit stores:

* root tree hash
* parent commit hash
* message
* timestamp

### Example:

```json
{
  "type": "commit",
  "tree": "88cc...",
  "parents": ["72fa..."],
  "message": "Initial commit",
  "author": "User",
  "timestamp": 1732212000
}
```

Stored as:

```
objects/88/cc....
```

---

# 🧱 Commit Model (No Staging)

MultiRepo VC supports **two types of commits**:

---

## ⭐ 1. `mrvc commit .` → Commit EVERYTHING

This creates a full snapshot of the entire working directory:

1. Walk every file (except ignored)
2. Hash contents
3. Write blob objects
4. Build directory trees
5. Build commit object
6. Update HEAD

Equivalent to Git’s “commit all changes” but without staging.

---

## ⭐ 2. `mrvc commit <files|patterns>` → Selective Commit

Users can commit a **subset of files**, defined by paths or wildcards:

Examples:

```
mrvc commit README.md
mrvc commit src/*.java
mrvc commit docs/** src/**/*.go
```

### Internal steps:

1. Load parent commit snapshot
2. Resolve file patterns into actual file paths
3. For each selected file:

    * Read working directory content
    * Hash and write blob
    * Replace corresponding entries inside parent tree
4. If selected files no longer exist → treat as deletion
5. Unselected files remain unchanged (inherited from parent snapshot)
6. Build new commit tree
7. Write commit object
8. Update HEAD

This allows **partial commits** without the complexity of a staging index.

---

# 🔄 Rename / Move Handling

Because commit selection uses **paths evaluated at commit time**, renames are naturally handled:

* If a selected path no longer exists → deletion
* If a wildcard matches a new location → updated content is committed
* No tracking IDs or staging needed

This keeps the system intuitive and predictable.

---

# 📐 Full Commit Algorithm (Unified)

Regardless of commit mode:

1. Load parent commit’s tree (if exists)
2. Determine which files to read:

    * ALL for `commit .`
    * SELECTED for `commit patterns`
3. Read file content
4. Compute blob hashes
5. Build or update trees
6. Write new commit object
7. Update `HEAD`

---

# 🧩 Why Snapshot Model First?

Snapshot-based commits:

* Are simple to reason about
* Enable fast diffs
* Are easy to revert
* Enable nested repository design
* Avoid staging/index complexities
* Provide structural persistence

Diffing and delta compression can come later.

---

# 🌀 Future Optimizations

### 🟧 Delta Compression

Store diffs for large files or repeated content.

### 🟦 Periodic Packfiles

Compact objects for performance and size savings.

### 🟩 Nested Repository Support

Allow directories to embed sub-repositories with independent histories.

---

# 🎉 Final Summary

MultiRepo VC provides:

* **Simple, staging-less workflow**
* **Full or selective commits via patterns**
* **Immutable object storage**
* **Tree-based directory snapshots**
* **Easy revert + predictable behavior**
* **Extensible base for nested repos and diffs**

This architecture is clean, modern, and far easier to maintain than a Git-like index system while still supporting partial commits.
