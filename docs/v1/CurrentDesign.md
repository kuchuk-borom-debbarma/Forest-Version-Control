
# 📘 MRVC — Current Design Summary (Before New RootRepo Architecture)

This summarizes the entire system as it exists **right now**, based on my implementation and discussions.

---

# 1. Core Concepts

MRVC is a **snapshot-based version control system** with support for **hierarchical repositories (nested repos)**.

Each repo contains:

```
.mrvc/
  HEAD
  HEAD_SUPER           # optional, super commit pointer
  metadata.json
  children.json         # list of direct children (path + repoName)
  objects/
```

---

# 2. Normal Commits (HEAD)

A normal commit:

* Captures the **file tree snapshot** for the repo.
* Stores blob + tree objects.
* Updates **HEAD**.

Nested repos are excluded from commits:

* Via `IgnoreNestedRepos: true`.

Commit object format:

```json
{
  "tree": "<hash>",
  "parent": "<hash or empty>",
  "message": "...",
  "author": "...",
  "timestamp": "..."
}
```

---

# 3. Linking Child Repos

`mrvc link <path>`

* Path must be *inside* the parent repo.
* Child must have `.mrvc/` and `metadata.json`.
* Metadata `repoName` is used as the **unique identifier** for that child.
* Writes entry to `children.json`:

```json
{
  "children": [
    { "path": "ChildA", "repoName": "repoA" }
  ]
}
```

Each repo may have multiple children, but only **direct children** are considered.

---

# 4. Super Commits (HEAD_SUPER)

A super commit is a **hierarchical snapshot**, referencing:

* The repo’s own latest HEAD
* Each direct child’s latest:

    * `HEAD_SUPER` (preferred)
    * OR `HEAD` (fallback)

Super commit object:

```json
{
  "self_head": "<HEAD>",
  "children": [
    {
      "path": "ChildA",
      "repoName": "childA",
      "ref": "<child HEAD or SUPER>",
      "type": "super|commit"
    }
  ],
  "message": "...",
  "author": "...",
  "timestamp": "..."
}
```

Super commit does *not* recurse deeper:
**only direct children are captured**.
Children carry *their own* super commits (which may include grandchildren).

Super commit updates:

```
.mrvc/HEAD_SUPER = <superCommitHash>
```

---

# 5. Status

`mrvc status` now shows:

### ✔ Normal commit state:

* Modified
* Deleted
* Untracked
  Compared against **HEAD**.

### ✔ Super commit state:

* Displays the hash of HEAD_SUPER
* Loads the super commit object
* Resolves all linked children
* Checks:

    * Child path exists
    * `.mrvc` exists
    * Metadata identity matches repoName
    * Snapshot object exists
* Shows warnings (e.g., missing super commit)
* Shows errors for broken links

This provides a complete hierarchical integrity check.

---

# 6. Error Handling

Super commit errors:

* Child missing
* Child not a repo
* Child has no HEAD
* RepoName mismatch
* Snapshot object missing

Link errors:

* Duplicate repoName
* Duplicate path
* Invalid path (outside)

Commit errors:

* No files
* Missing file
* Invalid tree structure

---

# 7. Current Limitations (Known + Accepted)

* Hierarchical filesystem dependency:
  Root repo depends on child directories being present.
* Super commit object references child objects stored *in child repos*.
* Moving a child repo breaks links (unless re-linked).
* Parent repo does not store actual child objects (no embedding).
* No recursive rebuilding—only direct child snapshot usage.

These were accepted constraints for the current architecture.

---

# 8. Planned New Architecture (Upcoming Work)

I decided the next iteration will switch to:

### ✔ A **Root Repository as a Pure Tree Controller**

* Does *not* store files.
* Does *not* make commits.
* Only manages:

    * Hierarchy tree
    * Child repo metadata
    * Unified hierarchical snapshots
    * Global versioning
    * Snapshot storage for the whole multi-repo system
