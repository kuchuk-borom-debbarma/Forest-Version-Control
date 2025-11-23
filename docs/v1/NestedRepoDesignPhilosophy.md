
# 📘 MRVC Design Philosophy

### *Independent Repositories With Hierarchical Time-Travel Snapshots*

MRVC (Multi-Repo Version Control) is built on a powerful idea:

> **Each repository is independent, but MRVC can snapshot the entire hierarchical structure of repositories as a reproducible system-level state.**

This gives MRVC the flexibility of a multi-repo ecosystem with the historical accuracy of a monorepo.

---

# 🎯 Core Principles

## **1. Every Repository Is Truly Independent**

Each directory containing a `.mrvc/` folder is a fully self-contained repository:

* Its own object store
* Its own commit history
* Its own tree hashing
* Its own lifecycle

MRVC **never merges nested repos** or rewrites their internals.
This makes nested structures safe, predictable, and portable.

---

## **2. Hierarchies Are Explicitly Declared**

MRVC does not infer nested relationships automatically.

Developers create links manually:

```
mrvc link ./repoA
mrvc link ./repoB
```

This generates:

```
.mrvc/children.json
```

which defines:

* Where child repositories live
* How the hierarchy is structured

This approach preserves developer intent and avoids “automatic magic.”

---

## **3. Structure and State Are Separate Concepts**

MRVC distinguishes between:

| Concept       | Meaning                                            | Stored In       |
| ------------- | -------------------------------------------------- | --------------- |
| **Structure** | Which repos belong to this repo                    | `children.json` |
| **State**     | What version each repo was at (including subtrees) | Super Commit    |

This separation prevents accidental coupling and keeps repos lightweight.

---

# 🧩 The Key Challenge

### **How do super commits work in a hierarchical system where child repos may already have their own super commits?**

Consider:

```
RootA
 └── SecondRoot
      └── RepoX
```

### Scenario:

* `RepoX` commits normally → has only `.mrvc/HEAD`
* `SecondRoot` performs a super commit → now has `.mrvc/HEAD_SUPER`
* `RootA` performs a super commit afterwards

### Problem:

Which “state” should `RootA` reference for `SecondRoot`?

* Its **regular HEAD** (file-level state)?
* Its **super commit HEAD_SUPER** (hierarchy-level state)?
* Something else?

Without a rule, the entire hierarchical versioning model becomes ambiguous and inconsistent.

---

# ✅ The Solution: Dual-Head Architecture

To solve this cleanly and elegantly, MRVC introduces **two heads per repo**:

```
.mrvc/HEAD         → local commit (single repo state)
.mrvc/HEAD_SUPER   → super commit (hierarchy snapshot)
```

This provides a deterministic rule:

## **Super Commit Rule**

When creating a super commit:

1. **If child repo has HEAD_SUPER → use HEAD_SUPER**
   (Child has already snapshotted its hierarchy.)

2. **Else → use child’s HEAD**
   (Child has no hierarchy or has not super-committed yet.)

3. **Apply recursively** through the entire hierarchy.

This resolves the ambiguity completely.

---

# 🌲 Example: How the Hierarchy Resolves

### Step 1: RepoX commits normally

```
RepoX HEAD = c1
RepoX HEAD_SUPER = none
```

### Step 2: SecondRoot super commits

It inspects RepoX:

* RepoX has no HEAD_SUPER → use HEAD (`c1`)

SecondRoot stores:

```
{ path: "RepoX", head: "c1", type: "commit" }
```

and generates:

```
SecondRoot HEAD_SUPER = sc1
```

### Step 3: RootA super commits

It inspects SecondRoot:

* SecondRoot HAS HEAD_SUPER → use that

RootA stores:

```
{ path: "SecondRoot", head: "sc1", type: "super-commit" }
```

Result:

* RootA super commit references SecondRoot's **entire hierarchy snapshot**
* SecondRoot super commit references RepoX's state
* RepoX's state remains independent

This gives perfect hierarchical time travel.

---

# 🔥 Final Design Summary

## **📌 Linking Phase (Structure Only)**

Users link repos manually to define the hierarchy:

```
.mrvc/children.json
```

This is purely structural.

---

## **📌 Super Commit Phase (State Snapshot)**

A super commit records:

* Each child repo's current state
* Whether that state is:

    * a normal commit
    * or a super commit

MRVC creates a **SuperCommitObject** that stores the entire multi-repo state in a reproducible way.

---

## **📌 Time Travel**

With hierarchical super commits, MRVC can reconstruct:

* The entire multi-repo project
* At any moment in history
* Consistently and recursively

This is a capability that neither Git submodules nor monorepos provide cleanly.

---

# ⭐ Conclusion

MRVC’s design delivers:

* Independent repositories
* Explicit hierarchical structure
* Multi-level super commits
* Universal time travel across repo trees
* Clean, deterministic recursion
* A system that scales to arbitrarily deep hierarchies

By separating:

* **structure (children.json)**
  from
* **state (HEAD vs HEAD_SUPER)**

MRVC achieves a robust, elegant, hierarchical version control model that is both simple to understand and powerful in practice.
