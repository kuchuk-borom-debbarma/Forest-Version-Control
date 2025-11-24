
# 📘 MRVC Hierarchical Versioning — Implementation Design Document

*(Finalized, RepoName-based Identity Model)*


---

# 1. Overview

MRVC is a **snapshot-based version control system** that supports **hierarchical nested repositories**.

Each repository is independent, but a parent repo can link other MRVC repositories as **direct children**.

MRVC supports:

* **Normal commits** → file-level snapshots (`HEAD`)
* **Super commits** → hierarchical snapshots (`HEAD_SUPER`)
* **children.json** → working (non-versioned) repo topology
* **Super commit objects** → versioned topology snapshots

This design enables multi-repo projects where each repo evolves independently,
but parents can take stable snapshots of their children when desired.

---

# 2. Repository Structure

Each MRVC repo contains:

```
.mrvc/
  HEAD                 ← Latest normal commit
  HEAD_SUPER           ← Latest super commit
  metadata.json        ← Repo metadata including repoName
  children.json        ← Direct child repos (working topology)
  objects/             ← All MRVC objects (blobs/trees/commits/super-commits)
```

---

# 3. Identity Model

Instead of using UUIDs, **repoName** is used as the repository identity.

### Rules:

1. repoName is defined in `metadata.json` at init and **must not change**.
2. repoName must be **unique among direct siblings** of a parent.
3. children.json stores repoName so the parent can detect moved repos.
4. Repo movement requires only path update—not identity change.

This keeps identity simple and predictable.

---

# 4. children.json Format (Working Topology)

Unlike history, this file represents the current, editable structure.

```json
{
  "children": [
    {
      "path": "libs/api",
      "repoName": "api"
    },
    {
      "path": "services/user",
      "repoName": "user"
    }
  ]
}
```

### Notes:

* Paths are **relative** to the parent repo.
* repoName is read from child `.mrvc/metadata.json`.

children.json is *not* versioned—it only affects the next super commit.

---

# 5. Linking Child Repositories

### Command

```
mrvc link <path-to-child>
```

### Behavior

1. Validate target folder contains a valid MRVC repo.
2. Load child’s `metadata.json` and extract `repoName`.
3. Convert absolute path → relative path.
4. Ensure no duplicate path or duplicate repoName.
5. Append `{path, repoName}` to children.json.

### Result

* Only children.json changes.
* No commit or super commit is created.
* Working topology is updated.

---

# 6. Normal Commit Behavior

A normal commit:

* Snapshots the current working tree (excluding nested repos)
* Updates `.mrvc/HEAD`
* Does NOT read children.json
* Does NOT touch HEAD_SUPER
* Does NOT include children or topology

Normal commits are **local-only**, file-level snapshots.

---

# 7. Super Commit Behavior

### Command

```
mrvc super-commit --message "..." [--strict]
```

Super commit creates a **hierarchical snapshot** that freezes:

1. The repo’s own current HEAD
2. The current list of children
3. Each child's stable snapshot (HEAD_SUPER) or HEAD (fallback mode)

Super commits are **not recursive**:

* They only inspect **direct children**
* Children handle their own subtree snapshots

This keeps boundaries clean and predictable.

---

# 8. Super Commit Resolution Rules (Final)

For repo **R**, super commit does:

---

## 8.1 Self Snapshot

Super commit **always** uses:

```
self_head = HEAD
```

Even if HEAD_SUPER exists.

Because HEAD represents the latest file-level state of the parent.

---

## 8.2 Child Snapshot

For each `{path, repoName}`:

1. Validate the child folder exists.
2. Validate `.mrvc/metadata.json` exists.
3. Validate metadata.Name matches repoName.
4. Load child HEAD and HEAD_SUPER.

### Resolution Logic

| Condition                | Used in super commit | type     |
| ------------------------ | -------------------- | -------- |
| Child has HEAD_SUPER     | HEAD_SUPER           | "super"  |
| Child has no HEAD_SUPER  | HEAD                 | "commit" |
| Strict mode and no super | ❌ error              | —        |

---

## 8.3 Building the Super Commit Object

```json
{
  "self_head": "<head>",
  "children": [
    {
      "path": "libs/api",
      "repoName": "api",
      "ref": "<childHash>",
      "type": "super"
    }
  ],
  "message": "...",
  "author": "...",
  "timestamp": "..."
}
```

---

## 8.4 Saving the Super Commit

Steps:

1. Serialize object
2. Hash it (content-addressed)
3. Save in `.mrvc/objects/<hash>`
4. Write hash to `.mrvc/HEAD_SUPER`

---

# 9. Strict Mode

```
mrvc super-commit --strict
```

Strict mode enforces:

* All children must have HEAD_SUPER
* Using child HEAD is prohibited
* Ensures fully stable, reproducible hierarchy

Default mode = flexible
Strict mode = controlled releases (ideal for CI)

---

# 10. No Recursion

Parents never traverse children’s children.

Child super commits already contain child refs recursively.

This gives:

* Clean boundaries
* Clear responsibilities
* Predictable behavior
* Infinite hierarchy without complexity

---

# 11. Versioned Topology vs Working Topology

| File              | Meaning                     |
| ----------------- | --------------------------- |
| children.json     | Working topology (editable) |
| SuperCommitObject | Versioned topology (frozen) |

Changing children.json does NOT modify history.
History is updated only when a new super commit is made.

---

# 12. Error & Warning Behavior

### Errors

* No HEAD present in parent
* Child repo does not contain `.mrvc/`
* Child metadata.json missing
* repoName mismatch
* Strict mode: child has no HEAD_SUPER
* Invalid children.json

### Warnings (non-strict)

* Child has no super commit → using HEAD

---

# 13. Example Hierarchy Flow

```
RepoC:
  mrvc commit
  mrvc super-commit → SC_C1

RepoB:
  mrvc link ../RepoC
  mrvc commit
  mrvc super-commit → SC_B1

RepoA:
  mrvc link ./RepoB
  mrvc super-commit → SC_A1
```

RepoA super commit will contain:

```
children: [
  { path: "RepoB", repoName: "repoB", ref: "SC_B1", type: "super" }
]
```

RepoB super commit includes:

```
{ path: "RepoC", ref: "SC_C1" }
```

This creates a complete stable chain.

---

# 14. Checkout (Future Work)

Super commit checkout will need to:

1. Restore self HEAD
2. Restore each child repo to `children[i].ref`
3. Recursively apply child super commits
4. Reconstruct full working directory and topology
5. Solely rely on .mrvc objects and not file system

This will be addressed in a separate design spec.
