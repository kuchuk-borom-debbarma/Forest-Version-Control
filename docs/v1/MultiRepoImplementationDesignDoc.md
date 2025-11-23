
# 📘 MRVC Hierarchical Versioning — Implementation Design Document

*(Based on current codebase + finalized design decisions)*


---

# 1. Overview

MRVC is a **snapshot-based version control system** with support for **hierarchical repositories** ("nested repos").
Each repo manages its own snapshots, and higher-level repos can create **super commits** that reference the stable snapshots of their direct children.

This allows teams to independently version their repos and provide stable integration points to parent repos.

This document describes:

* How hierarchy is structured
* How linking is stored
* How child super commits are resolved
* How parent super commits are created
* What errors/warnings should occur
* How this integrates with existing commit logic
* New files, structures, and algorithms to add

---

# 2. Repo Structure Extensions

Each MRVC repo currently stores:

```
.mrvc/
  HEAD
  metadata.json
  objects/
```

We will extend this with:

```
.mrvc/
  HEAD_SUPER          <-- NEW: pointer to last super commit (optional)
  children.json       <-- NEW: list of direct child repos
```

### 2.1 children.json format

```json
{
  "children": [
    "path/to/childA",
    "path/to/childB"
  ]
}
```

Paths must be:

* relative to parent repo root
* normalized forward-slash format
* validated to contain `.mrvc/` (i.e., a valid MRVC repo)

---

# 3. Super Commit Object Format

A super commit represents a **stable snapshot** of a repo AND its direct children.

### SuperCommitObject JSON:

```json
{
  "self_head": "<commit-hash>",
  "children": [
    { "path": "childA", "ref": "<child-super-or-head>", "type": "super" },
    { "path": "childB", "ref": "<child-super-or-head>", "type": "commit" }
  ],
  "message": "Stable snapshot",
  "author": "Author",
  "timestamp": "<utc-ms>"
}
```

### Notes

* `self_head` = repo HEAD at time of super commit
* Children entries reference the exact snapshot hash used
* For each child, `type` will be either:

    * `"super"` → referencing child’s HEAD_SUPER
    * `"commit"` → referencing child’s HEAD (if they have no super commit)

---

# 4. Linking Children (New Command)

Add CLI:

```
mrvc link --child path/to/repo
```

### Steps

1. Normalize path
2. Verify `.mrvc/` exists under child
3. Append to children.json
4. Ensure no duplicate entries
5. Store relative path from parent root

---

# 5. Behavior When Commit is Made (Normal Commit)

Normal commit should:

* NEVER include nested repos
  Already ensured because `ListFiles(... IgnoreNestedRepos: true)`
* Only affect `HEAD`, not `HEAD_SUPER`
* Only update local working tree snapshot

No changes needed to commit logic.

---

# 6. Behavior When Super Commit is Made

Add CLI:

```
mrvc super-commit --message "..." [--strict]
```

### **6.1 High-Level Algorithm**

Given repo R:

1. Load R’s own HEAD
2. Load R’s `children.json`
3. For each DIRECT child C:

    * Validate C is an MRVC repo
    * If C has HEAD_SUPER → use it
    * Else if strict mode → error
    * Else → use HEAD, but warn
4. Build `SuperCommitObject`
5. Hash + store
6. Update `.mrvc/HEAD_SUPER`
7. Print summary

---

# 7. Detailed Super Commit Flow (Step-by-Step)

### Function: `CreateSuperCommit(message, author, strict bool)`

#### Step 1 — Load Repo Root + Validate MRVC

```go
repoRoot := fs.GetCurrentDir()
if !IsDirPresent(repoRoot+"/.mrvc") → error
```

#### Step 2 — Load self HEAD

Use existing helper:

```
selfHead := readHEAD()
```

If empty: error
Super commit requires at least one commit.

#### Step 3 — Load children list

`children.json` format:

```json
{ "children": ["childA", "childB"] }
```

If file doesn't exist → children = empty array.

#### Step 4 — Resolve each child

For each child:

1. childAbs := repoRoot + "/" + childPath
2. Verify `.mrvc` exists
3. Determine:

```
childSuper := readHEADSUPER(childAbs)
childHead  := readHEAD(childAbs)
```

4. Decision logic:

* If childSuper != "" →
  Use {ref: childSuper, type: "super"}
* Else if strict →
  throw error: "Child X has no super commit"
* Else
  Use {ref: childHead, type: "commit"}
  Print warning: "Child X has no stable snapshot; using HEAD"

#### Step 5 — Build `SuperCommitObject`

```
{
  self_head: selfHead,
  children: resolved [],
  message,
  author,
  timestamp
}
```

#### Step 6 — Hash + store

Use `HashContent` and `SaveObject` (same as commits)

Store under:

```
.mrvc/objects/<hash>
```

#### Step 7 — Update HEAD_SUPER file

```
.mrvc/HEAD_SUPER = hash
```

#### Step 8 — Done

---

# 8. Super Commit Strict Mode

Activate with:

```
mrvc super-commit --strict
```

Behavior:

* If any child has NO super commit → immediate error
* Prevents unstable snapshots
* Ideal for CI pipelines / release branches

---

# 9. Why We Do NOT Recurse Deeply

Parent repo only sees **direct children**.

Child super commits carry subtree snapshots.

Parent super commits must TRUST child super commits — do NOT walk into grandchildren.
This preserves team boundaries and ensures predictable snapshots.

---

# 10. Status / Checkout Implications

### Status

* Does NOT consider child repos
* Only compares working tree vs HEAD
* No change needed

### Checkout (future)

* When implementing checkout:

    * HEAD restores local repo
    * HEAD_SUPER restores hierarchical snapshot by recursively walking super commits
    * This will require implementing a restore mechanism for each repo

(Not needed now)

---

# 11. Integration With Current Code

## Where new logic should be added:

### New Files (recommended):

```
src/internal/core/version_control/v1/super_commit.go
src/internal/core/version_control/v1/children.go
src/internal/commands/super_commit.go
src/internal/commands/link.go
```

## New helpers:

* `readHEADSUPER(repoRoot string)`
* `updateHEADSUPER(repoRoot, hash string)`
* `LoadChildren(repoRoot string)`
* `ValidateChildRepo(path string)`
* `ResolveChildSnapshot(childPath string, strict bool)`

---

# 12. Error Conditions

### Super Commit Errors

* No HEAD exists
* Child repo missing or invalid
* In strict mode: child has no HEAD_SUPER
* Invalid children.json entries

### Warnings (non-strict)

* Child has no super commit → using HEAD

---

# 13. Example Workflow

```
RepoB:
  mrvc commit ...
  mrvc super-commit -m "stable B1"

RepoA:
  mrvc link --child path/to/RepoB
  mrvc super-commit -m "root stable snapshot"
```

RepoA’s super commit will contain:

```
children: [
  { path: "path/to/RepoB", ref: "B1", type: "super" }
]
```

---

# 14. Final Notes

This design:

* keeps repos independent
* supports team workflows cleanly
* guarantees stable snapshots
* avoids recursive complexity
* ensures predictable, explicit behavior
* naturally scales to large hierarchies
