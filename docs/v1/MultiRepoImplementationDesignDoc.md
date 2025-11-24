
# 📘 MRVC Hierarchical Versioning — Implementation Design Document

*(Finalized Architecture & Implementation Plan)*

---

# 1. Overview

MRVC is a **snapshot-based version control system** that supports **hierarchical, nested repositories**.
Each repository is fully independent but can link other MRVC repositories as **direct children**.

MRVC provides:

* **Normal commits**, generating `HEAD`
* **Super commits**, generating `HEAD_SUPER`
* **Stable hierarchical snapshots**, representing the repo + its direct children's stable snapshots
* **children.json**, representing the *working topology*
* **Super commit objects**, representing *versioned topology*

This design allows independent teams to work in isolated repos while offering stable integration points through super commits.

This document describes:

* Repository structure
* Linking child repositories
* children.json behavior
* Super commit behavior
* Strict vs. default snapshot modes
* Non-recursive hierarchical behavior
* Integration with current commit workflow
* Error handling and warnings

---

# 2. Repository Structure

Every MRVC repository contains:

```
.mrvc/
  HEAD                ← Latest normal commit (local snapshot)
  HEAD_SUPER          ← Latest super commit (stable snapshot)
  metadata.json       ← Repo metadata
  children.json       ← List of direct child repos (working topology)
  objects/            ← All MRVC objects (blobs, trees, commits, super commits)
```

### 2.1 children.json Format

This file defines the **current working hierarchy** (not versioned).

```json
{
  "children": [
    "childA",
    "libs/api",
    "packages/frontend"
  ]
}
```

Rules:

* Paths must be **relative** to parent repo root
* Paths must use **forward slashes**
* Each child must be a valid MRVC repository (`.mrvc` folder + `metadata.json`)

children.json is **not part of a commit**, but **super commits snapshot it**.

---

# 3. Super Commit Object Format

A super commit describes:

* The repo's own HEAD at that time
* The exact list of children that existed at that time
* The snapshot (HEAD_SUPER or HEAD) selected from each child
* The stable topology for checkouts

### JSON Structure

```json
{
  "self_head": "<commit-hash>",
  "children": [
    {
      "path": "childA",
      "ref": "<hash>",
      "type": "super"
    },
    {
      "path": "childB",
      "ref": "<hash>",
      "type": "commit"
    }
  ],
  "message": "Stable snapshot",
  "author": "Author",
  "timestamp": "<utc-ms>"
}
```

Meaning:

* `self_head` — Pointer to this repo's HEAD at super-commit time
* `children[]` — Frozen hierarchy:

    * `path`: relative path to child
    * `ref`: child HEAD_SUPER or HEAD
    * `type`: `"super"` or `"commit"`
* `message`, `author`, `timestamp`: metadata

---

# 4. Linking Children

### Command

```
mrvc link <path-to-child>
```

### Behavior

1. Normalize input path
2. Convert to absolute path
3. Validate directory contains:

    * `.mrvc/`
    * `.mrvc/metadata.json`
4. Convert absolute child path → **relative to parent repo root**
5. Load `.mrvc/children.json` (or create it)
6. Add entry if not duplicate
7. Save back to `.mrvc/children.json`
8. No commit or super commit is created

**Linking updates only children.json, not history.**

### Why?

children.json = **working topology**
Super commits = **versioned topology**

---

# 5. Normal Commit Behavior

A normal commit:

* Scans the working directory
* Ignores child MRVC repos (via `IgnoreNestedRepos`)
* Creates a blob/tree snapshot
* Updates `.mrvc/HEAD`
* Does NOT affect `.mrvc/HEAD_SUPER`
* Does NOT snapshot children.json

### Normal commits represent file-level history only.

---

# 6. Super Commit Behavior

### Command

```
mrvc super-commit --message "msg" [--strict]
```

### Purpose

Create a **stable hierarchical snapshot** capturing:

* Repo's current HEAD
* Current children.json structure
* Each child repo's stable snapshot (HEAD_SUPER)
* Or, if absent, child’s HEAD (default mode only)

### KEY RULE

**Super commits look ONLY at direct children.**
No recursion into grandchildren.

Child super commits contain their own children recursively — maintaining clean team boundaries.

---

# 6.1 Super Commit Algorithm (Final)

Given repo **R**:

1. Ensure R has at least one normal commit:

   ```
   selfHead := readHEAD()
   if selfHead == "" → error
   ```
2. Load children.json (may be empty)
3. For each child in children.json:

    * Validate it is a valid MRVC repo
    * Read child HEAD
    * Read child HEAD_SUPER (stable snapshot)
4. Apply resolution rules:

### Resolution Rules

#### **Case 1 — Child has HEAD_SUPER**

Always use it:

```
ref = childSuper
type = "super"
```

#### **Case 2 — Child has no super commit**

* **Strict mode** → throw error
* **Default mode** → use child HEAD + warn:

```
ref = childHead
type = "commit"
(WARN: child has no stable snapshot)
```

5. Build `SuperCommitObject`
6. Hash + store object in `.mrvc/objects/`
7. Update `.mrvc/HEAD_SUPER` with the new hash
8. Done

---

# 7. Strict Mode (Optional)

Enable with:

```
mrvc super-commit --strict
```

Strict mode enforces:

* All children MUST have HEAD_SUPER
* Using child HEAD is NOT allowed
* Guarantees fully stable tree
* Ideal for CI/CD and release pipelines

Default mode is user-friendly; strict mode is enforcement.

---

# 8. Why Super Commits Do NOT Recurse

Parent repositories must NOT inspect grandchildren.
Child repos are responsible for their own subtree.

This ensures:

* Clean module boundaries
* Predictable behavior
* No accidental inclusion of unstable deep changes
* Simple mental model
* Hierarchy scales arbitrarily deep

Example:

```
RootA
 └── RepoB
      └── RepoC
```

* RepoC super-commit → C_SC1
* RepoB super-commit → includes C_SC1
* RootA super-commit → includes B_SC1

Perfect delegation.

---

# 9. Versioned Topology Model

### children.json is NOT history

It is only the current working structure.

### Super commits ARE the history

They store:

* which children existed
* which snapshot each child contributed

Thus:

* Linking/unlinking modifies children.json only
* The next super commit records the new topology

This ensures:

* Full reproducibility of past topology
* Clean, explicit historical records

---

# 10. Status / Checkout Implications

### Status

* Only inspects file changes
* Ignores nested repos
* Requires no modification

### Checkout (future)

Super commit checkout must:

* Read SuperCommitObject
* Restore:

    * repo self HEAD
    * each child to exact snapshot referenced
* Recursively apply for deeper levels

**Checkout design is future work, separate document.**

---

# 11. Integration With Codebase

Recommended new files:

```
src/internal/core/version_control/v1/super_commit.go
src/internal/core/version_control/v1/children.go
src/internal/commands/super_commit.go
src/internal/commands/link.go
```

Key helper functions:

* `readHEADSUPER(path string)`
* `updateHEADSUPER(path, hash string)`
* `LoadChildren(path string)`
* `ResolveChildSnapshot(path string, strict bool)`
* `IsValidMRVCRepo(path string)`

---

# 12. Error Conditions

### Super Commit Errors

* Repo has no HEAD
* Invalid children.json
* Child repo missing `.mrvc/`
* Child repo missing metadata
* Strict mode and child missing HEAD_SUPER
* Unable to read or write MRVC object files

### Warnings (non-strict mode)

* Child has no super commit → using HEAD

---

# 13. Real Workflow Example

```
RepoC:
  mrvc commit
  mrvc super-commit -m "stable C1"

RepoB:
  mrvc link ../RepoC
  mrvc commit
  mrvc super-commit -m "stable B1"

RootA:
  mrvc link ./RepoB
  mrvc super-commit -m "root stable snapshot"
```

RootA super commit contains:

```
children: [
  { path: "RepoB", ref: "B1", type: "super" }
]
```

RepoB super commit contains:

```
children: [
  { path: "RepoC", ref: "C1", type: "super" }
]
```
