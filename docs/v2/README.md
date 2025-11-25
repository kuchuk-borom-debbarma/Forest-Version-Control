# 🌲 Forest Version Control (FVC)

**Unified Architecture & Design Specification**  
*Snapshot-Based, Multi-Repository, Hierarchical Version Control System*

---

<details>
<summary><strong>1. Introduction</strong></summary>

Forest Version Control (FVC) is a multi-repository version control system designed to manage projects composed of
several independent but related repositories.

It provides:

* **Local repository independence** — Each repository can function standalone
* **Central Tree Storage** — Unifies versioning across repositories
* **Versioned hierarchical structure** — Track relationships over time
* **Full-system or subtree snapshots** — Super Snap mechanism
* **Immutable, hash-based object storage** — Ensures data integrity

FVC is built to solve the limitations of path-based linking between repositories and ensures reproducibility, safety,
and scalability for projects that grow across multiple codebases.

</details>

---

<details>
<summary><strong>2. Limitations of the Previous Design</strong></summary>

The original approach used a fully independent repository model where each repository:

* Had its own `.mrvc/.fvc` folder
* Stored its own objects, commits, and metadata
* Identified child repositories by filesystem paths

This created several major issues:

### 2.1 Loss of Reproducibility

If a child repository was removed, renamed, or moved, the parent could not reconstruct its state because all
relationships depended on absolute paths.

### 2.2 Heavy Reliance on Filesystem Layout

Relationships between repositories depended on directory structure. Any movement of repositories invalidated links,
breaking history and snapshot integrity.

### 2.3 No Centralized History

Since every repository stored its own objects, no system-level consistency was possible. Capturing a unified snapshot of
multiple repositories was nearly impossible.

**These limitations led to the development of a central Tree Storage and a complete redesign of how repository
relationships and multi-repository snapshots are handled.**

</details>

---

<details>
<summary><strong>3. New Architecture Overview</strong></summary>

FVC introduces a **Tree Storage** that acts as a central, unified object store and relationship manager for multiple
repositories.

### Core Principles

✔ **Local repositories remain independent by default**  
They support commits, snapshots, and metadata.

✔ **Repositories can link to a Tree Storage**  
This transitions them from standalone mode to shared-storage mode.

✔ **The Tree Storage maintains:**

* A central object store
* Metadata for all linked repositories
* A versioned hierarchy of repository relationships
* Independent Super Snap histories for each repository

✔ **Commit history and hierarchy history are fully decoupled**  
Both systems evolve independently.

✔ **Super Snap captures hierarchical commit state**  
A full-tree or subtree snapshot is represented as a tree of Super Snap Objects.

This architecture is scalable, reproducible, and robust against filesystem changes.

</details>

---

<details>
<summary><strong>4. Repository Design</strong></summary>

Repositories are **local and independent** when first initialized.

```bash
fvc init --name repoName --author "optional"
```

After initialization:

```
repo/
  .fvc/
    metadata.bin
    objects/
    HEAD
```

### 4.1 Metadata

Stored in `metadata.bin`, containing:

* Repository name
* Author
* Creation timestamp
* Optional description

### 4.2 Commit Operation

```bash
fvc commit --message "..." --author "..." --files "file1,file2,*"
```

**Commit behavior:**

1. Select files (explicit paths or `*` for all except ignored)
2. Compute file hashes
3. Store files as blob objects
4. Build directory trees bottom-up as Tree Objects
5. Create commit objects linking:
    * Root tree hash
    * Parent commit
    * Metadata
6. Update the local `HEAD`

### 4.3 Isolation

Repositories do **not** track or store child repositories locally. Each repo is fully self-contained until linked to a
Tree Storage.

</details>

---

<details>
<summary><strong>5. Tree Storage</strong></summary>

*A central shared storage and relationship manager*

A Tree Storage is a dedicated directory that hosts:

* **Central object store** — Deduplicated storage for all linked repos
* **Linked repository metadata** — Information about each connected repo
* **Versioned repository hierarchy** — Tree Objects tracking relationships
* **Super Snap objects** — Hierarchical snapshots

**A repository can only join a hierarchy after linking to a Tree Storage.**

</details>

---

<details>
<summary><strong>6. Linking a Repository to Tree Storage</strong></summary>

Repositories link to a central storage using:

```bash
fvc link-storage --storage "/abs/path/to/storage" \
                  --path "logical/tree/path"
```

### 6.1 Linking Steps

**1. Create link file**

Stored in `repo/.fvc/link.json`:

```json
{
  "repoID": "uuid-123",
  "repoName": "PatientService",
  "storagePath": "/abs/path/to/storage",
  "desiredTreeLocation": "services/patient",
  "workspace_path": "/users/admin/PatientService",
  "linkedAt": "2025-03-01T14:55:22Z",
  "version": 1
}
```

**2. Transfer all objects**

From repo → tree storage. Deduplication is handled automatically.

**3. Create repository directory in Tree Storage**

```
storage/.fvc/repositories/<repoID>/
    metadata.bin
    HEAD
    repoName.txt
    relativePath.txt
    workspace_path.txt
```

**4. Remove local object storage**

The repo now becomes a "thin workspace," delegating storage to Tree Storage.

After linking, the repository is eligible to become part of the FVC hierarchy.

</details>

---

<details>
<summary><strong>7. Versioned Hierarchy of Repositories</strong></summary>

*Relationship modeling via Tree Objects*

Once repositories are linked, relationships between them are defined via:

```bash
fvc add-child --parent <path> --child <path1,path2,...>
```

This creates a versioned hierarchy in Tree Storage.

### 7.1 Tree Objects

Each hierarchy node is represented by a **Tree Object** stored under:

```
.fvc/trees/<hash>
```

### 7.2 Tree Object Structure

```json
{
  "repoID": "uuid-root",
  "children": [
    {
      "repoID": "uuid-a",
      "childTreeHash": "..."
    },
    {
      "repoID": "uuid-b",
      "childTreeHash": "..."
    }
  ],
  "parentTreeHash": "previous_hierarchy_hash_or_null",
  "message": "Hierarchy updated",
  "author": "username",
  "timestamp": "2025-03-01T10:32:14Z"
}
```

### 7.3 Tree HEAD

The latest hierarchy version is stored in `.fvc/treeHEAD`

**Hierarchy changes:**

* Never modify old Tree Objects
* Always create a new one
* Update treeHEAD

This gives a fully versioned structure of repository relationships.

</details>

---

<details>
<summary><strong>8. Super Snap</strong></summary>

*Hierarchical snapshotting of repo commit states*

Super Snap captures a consistent snapshot of:

* A repository's commit HEAD
* All of its descendants' commit HEADs
* The hierarchical structure at that moment

Super Snap can be taken at **any level** of the hierarchy. Each repository has its own **independent super snap history
**.

### 8.1 Super Snap Object

For each repository in the snapshot's scope:

```
tree_storage/.fvc/supersnaps/<repoID>/<hash>
```

**Object Structure:**

```json
{
  "repoID": "uuid-a",
  "commitHash": "commit_of_a",
  "children": [
    {
      "repoID": "uuid-c",
      "superSnapHash": "hash_of_c_supersnap"
    }
  ],
  "timestamp": "2025-03-02T12:10:44Z",
  "author": "John Doe",
  "message": "Snapshot at A",
  "previousSuperSnapHash": "previous_supersnap_or_null"
}
```

Each object:

* Stores its own HEAD
* Stores only direct child heads
* Links to child super snap objects
* Forms a version chain via `previousSuperSnapHash`

### 8.2 Super Snap Flow

**Step 1 — Identify Snapshot Root**

```bash
fvc super-snap --message "..."
```

Determine which repo the command is invoked from.

**Step 2 — Get Current Hierarchy**

Use treeHEAD to traverse the Tree Objects and identify descendants.

**Step 3 — Recursively Snapshot**

For each repo:

1. Get commit HEAD
2. Snapshot child repositories recursively
3. Create a Super Snap Object containing:
    * Its HEAD
    * Child super snap hashes
    * Metadata
    * Previous super snap hash
4. Hash and store the object

**Step 4 — Return Root Super Snap Hash**

This hash defines the snapshot of the entire subtree.

### 8.3 Independent Level Histories

Snapshots occur independently at each level:

* Root may have many super snaps
* A child repo may have its own super snaps
* A parent's snapshot may incorporate a child's latest snapshot

This creates a flexible, multi-level versioning system.

</details>

---

<details>
<summary><strong>9. Object Storage Structure Summary</strong></summary>

### 9.1 Repository Local

Before linking:

```
.fvc/
  objects/
  metadata.bin
  HEAD
```

### 9.2 Tree Storage

```
.fvc/
  objects/
  repositories/
    <repoID>/
       metadata.bin
       HEAD
       repoName.txt
       relativePath.txt
       workspace_path.txt
  trees/
    <hash>   (Tree Objects — hierarchy versions)
  supersnaps/
    <repoID>/
      <hash>  (Super Snap Objects)
  treeHEAD
```

</details>

---

<details>
<summary><strong>10. Glossary</strong></summary>

| Term              | Meaning                               |
|-------------------|---------------------------------------|
| Repository        | Independent version-controlled unit   |
| Tree Storage      | Central storage for linked repos      |
| Tree Object       | Immutable hierarchy node              |
| treeHEAD          | Pointer to latest hierarchy version   |
| Super Snap        | Hierarchical snapshot of commit HEADs |
| Super Snap Object | Snapshot object for one repo          |
| repoID            | Unique ID assigned at linking         |
| Thin workspace    | Repo without local object store       |

</details>

---

<details>
<summary><strong>11. Conclusion</strong></summary>

This unified architecture provides a clean, scalable way to:

* ✅ Maintain independent repositories
* ✅ Link them under a central object store
* ✅ Version their hierarchical relationships
* ✅ Capture full consistent snapshots of any subtree

**FVC ensures reproducibility, integrity, and clarity in multi-repository environments without relying on fragile
filesystem paths or monolithic project structures.**


</details>