
# 📘 MRVC — New Architecture & Design Philosophy

*A Centralized, Hierarchical Multi-Repository Version Control Model*

---

# ❗ Drawbacks of the Old Design

Before introducing the new architecture, it is important to understand **why the original MRVC design had fundamental limitations**.

### 1. Each Repository Stored Its Own Objects Independently

In the old model, **every repository had its own `.mrvc` directory** storing:

* blob objects
* tree objects
* commit objects
* super-commit objects
* metadata

This meant:

* A parent repo **did not contain** the objects of its children
* Parents only stored **references** to child repos
* If a child repo:

    * was moved
    * was deleted
    * became corrupted
    * or its `.mrvc` folder changed

then the **parent could no longer reconstruct** that child’s state.

This broke one of the core guarantees of a version control system:
➡ **Historical reproducibility.**

---

### 2. Heavy Reliance on Filesystem Paths

Repositories were linked **only through file paths**, e.g.:

```
parent/
   child/
```

This was fragile:

* Moving a directory broke the links
* Renaming repos broke the links
* Restructuring folders broke the links
* Externalizing a repo was impossible
* Parent repos were tightly coupled to the physical directory layout

This resulted in a poor developer experience and an unstable hierarchy.

---

# 🚀 The New Architecture (Central Storage Model)

The redesigned MRVC architecture removes these limitations entirely by introducing a **central storage system** that becomes the authoritative source of all version-control data.

Repositories become **lightweight workspaces**, and central storage holds all version-control objects, hierarchy information, and tree snapshots.

Below is the complete flow of how the new system works.

---

# 🟦 1. Initialize Your First Repository

You begin with a normal project, for example a **hospital management system**:

```
HospitalRoot/
```

You initialize MRVC:

```bash
mrvc init
```

You now continue working normally:

* creating commits
* merging
* branching

Everything behaves like a typical single-repo version control system.

At this point, the repo is standalone — there is **no hierarchy**, and no central storage yet.

---

# 🟧 2. Create Another Repository (Future Child)

Later, you build a new component — for example:

```
PatientService/
```

You initialize MRVC here as well:

```bash
mrvc init
```

This is also a normal, standalone repository.

---

# 🟥 3. Attempt to Add PatientService as a Child Repository

At this point you run:

```bash
mrvc add-child ../PatientService
```

But this **fails** with a clear message:

> ❌ Cannot add child repository.
> This repository is not linked to a central storage.

Why?

Because **hierarchy can only be created once repositories share the same central storage**.

This ensures:

* consistent snapshot storage
* centralized history
* no duplication
* full reconstructability

---

# 🟩 4. Create the Central Storage

You now create a central storage for the entire hospital project:

```bash
mrvc create-storage HospitalManagementStorage
```

This creates:

```
HospitalManagementStorage/
  .mrvc/
    objects/
    repo_index.json
    tree_index.json
    tree_snaps/
```

This is the **global MRVC storage**.

It will hold:

* all blobs
* all tree objects
* all commit objects
* all super-snap objects
* hierarchy mapping
* HEAD pointers
* all snapshots, forever

Repos will no longer store their own objects.

---

# 🟦 5. Link the Root Repository to Central Storage

In the root project (`HospitalRoot`), you now run:

```bash
mrvc link-storage ../HospitalManagementStorage
```

This performs three things:

### ✔ 1. Transfers all existing commit objects

All objects from the root repo’s `.mrvc` directory are moved to the central storage **once**.

### ✔ 2. Creates a `.mrvc-link.json` inside the repository

This file simply points the repo to central storage.

### ✔ 3. Registers the repo inside central storage

It becomes a managed repository.

From this point forward:

**All future commits in HospitalRoot will be stored inside the central storage, not locally.**

---

# 🟧 6. Add the Child Repository to the Same Storage

Now you repeat the same linking process:

```bash
mrvc link-storage ../HospitalManagementStorage
```

inside `PatientService`.

* All its existing objects are transferred once
* It receives its `.mrvc-link.json`
* It becomes a managed repo inside central storage

Now **both** repos share the same central storage.

---

# 🟨 7. Add Child to Parent (Now Allowed)

Now the command succeeds:

```bash
mrvc add-child ../PatientService
```

Central storage updates:

* hierarchy (`tree_index.json`)
* metadata about the relationship

From now on, MRVC knows:

```
HospitalRoot → PatientService
```

and can manage snapshots across both.

---

# 🟫 8. All Future Commits Are Stored in Central Storage

Whether you commit in:

```
HospitalRoot/
```

or

```
PatientService/
```

MRVC will:

* compute blob/tree/commit objects
* send them to the central storage
* update the central storage HEAD
* update the central hierarchy
* ensure consistency across the whole project

Local repos no longer maintain `.mrvc/objects`.

They are pure **workspaces**.

---

# 🌳 9. Super-Snap (Global Snapshot)

Super-snap now works at the **tree level**, not inside children.

When you run:

```bash
mrvc super-snap
```

MRVC:

1. Finds all children of the current repo
2. Performs a commit of each repo (centralized)
3. Creates a **tree snapshot** representing the full hierarchy
4. Stores the tree snapshot in central storage
5. Ensures full historical reproducibility — even if repos move or disappear

This fully fixes the original system’s worst limitation.

---

# 🔮 10. Future: Server-Based Central Storage

Currently, central storage is a folder on the filesystem.

In the future, MRVC will be able to point to a **server-based** storage backend:

```
mrvc link-storage http://mrvc-storage.mycompany.com
```

This removes the last remaining dependency on file paths completely.

---

# ⭐ Summary — Why the New Architecture is Superior

### ✔ No more duplicated snapshots

### ✔ No more filesystem path dependency

### ✔ Central, authoritative object store

### ✔ Repos behave like real workspaces

### ✔ Fully reconstructable history

### ✔ Super-snap is consistent and simple

### ✔ Clean separation of:

* storage
* repo
* hierarchy
