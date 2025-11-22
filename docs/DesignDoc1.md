
# 📘 MultiRepo VC — Architecture & Roadmap

*A Snapshot-Based Version Control System With Nested Repository Support (Planned)*
*(Updated based on current implementation)*

---

# 📍 Overview

MultiRepo VC is a **snapshot-based version control system**.
Each commit captures a complete tree of directory structures and file contents at commit time.

Key characteristics:

* **No staging area**
* **Content-addressed objects** (SHA-256)
* **Tree-based snapshots**
* **Simple CLI commands**
* **Selective file commits (explicit list only)**
* `"*"` wildcard allows *commit everything*
* `.mrvcignore` exclusion system

Pattern-based commits (e.g., `src/*.go`) are **not supported yet** and appear only in the roadmap.

---

# 🧭 Design Philosophy

### 1. **Immutable Object Storage**

* Blobs, trees, and commits are hashed using pure SHA-256 of their serialized content.
* Changing content produces new hashes → new objects → new snapshot.

### 2. **No Staging Area**

Current commit modes:

* `mrvc commit --message="msg" --files file1 file2`
* `mrvc commit --message="msg" --files *`

File selection is explicit.
Pattern matching is a future feature.

### 3. **Directory Snapshot Model**

Each commit stores a complete **tree of directory objects** representing the state of the repository.

Trees are:

* Built bottom-up
* Deterministically ordered
* Content-hashed after serialization

---

# 📝 Repository Initialization

`mrvc init --name <repoName> --author <author>`

Creates:

```
.mrvc/
  metadata.json
  HEAD
  objects/
```

### `metadata.json`

Matches actual code ():

```json
{
  "name": "MyRepo",
  "author": "Kuku",
  "created_at": "1732211000"
}
```

### `HEAD`

Initially empty:

```
<empty>
```

### `objects/`

Stores all blobs, trees, and commits:

```
objects/<first2>/<remaining>
```

---

# 📦 Blob Objects

Blobs contain **raw file bytes** exactly as read.

Hashing:

```
sha256(fileBytes)
```

Storage:

```
.mrvc/objects/ab/cdef1234...
```

---

# 🌳 Tree Objects

A tree represents a directory.

Updated actual format ():

```json
{
  "entries": [
    { "name": "main.go", "entry_type": "blob", "hash": "..." },
    { "name": "src", "entry_type": "tree", "hash": "..." }
  ]
}
```

### Bottom-up Construction (Actual Behavior)

The implementation:

* Tracks each directory touched by committed files.
* Ensures all parent directories exist.
* Sorts directories by depth (deep → shallow).
* Adds subtree entries after hashing children.
* Sorts entries alphabetically to guarantee deterministic hashing.

---

# 🔗 Commit Objects

Actual commit format ():

```json
{
  "tree": "rootTreeHash",
  "parent": "previousCommitHash or empty",
  "message": "Commit message",
  "author": "Author",
  "timestamp": "1732212000"
}
```

Differences from conceptual doc:

* `parent` is a **single string**, not a list.
* Commit JSON does **not** include `"type": "commit"`.
* Timestamp is stored as **stringified milliseconds**.

Storage:

```
objects/ab/cdef123...
```

---

# 🧱 Commit Model (Current Behavior)

Two modes:

---

## 1. **Explicit file commit**

Example:

```
mrvc commit --message="update" --files src/main.go README.md
```

Behavior:

* Paths are normalized to absolute.
* Each file must exist (no globbing).
* Only listed files are included in the snapshot.
* Parent directories are ensured automatically.

Unlisted files:

* Are **not included**, unless `"*"` is used.

---

## 2. `"*"` → Commit entire repository

```
mrvc commit --message="all" --files *
```

Behavior:

* Runs `ListFilesExcludingIgnore()` ()
* Reads `.mrvcignore` rules
* Excludes system folder `.mrvc`

This is the only wildcard currently supported.

---

# 🧾 Ignore System

`.mrvcignore` supports simple patterns:

* `*.ext`
* `prefix*`
* `folder/`
* exact matches

These are implemented in `IsIgnored()` ().

---

# 🔨 CLI Command System

Commands are registered dynamically via `init()` in each command file.

Current commands:

* `init`
* `commit`

### Argument Model

The parser supports:

* `--key value1 value2`
* `--key=value`
* `--flag` (boolean)
* positional arguments

Stored as `map[string][]string`.

---

# 🏗️ Roadmap (Planned Features)

These were in the original conceptual doc but **not yet implemented in code**.

### 1. **Pattern-Based Selective Commits**

Examples:

```
mrvc commit src/*.go
mrvc commit docs/** src/**/*.java
```

Requires:

* Full globbing engine
* Pattern → path resolution logic
* Integration with ignore rules

### 2. **Nested Repository Support**

Allow directories to act as independent sub-repositories.

Planned behaviors:

* Tree isolation
* Cross-references between repos
* Selective linking of history

### 3. **Multi-parent Commits**

Support merges:

```
"parents": ["hash1", "hash2"]
```

### 4. **Delta Compression**

Store diffs for large unchanged files.

### 5. **Packfile System**

Periodically compress object store.
