
# 🚨 Project Discontinued — Please Switch to **GitGroove**

### 👉 **Next-generation multi-repo tooling is now developed under: [GitGroove](https://github.com/kuchuk-borom-debbarma/GitGrove)**

*(Built with the same goal in mind, but practical, maintainable, and powered by Git.)*

---

# 🌲 ForestVC — *Project on Hold*

ForestVC was an experimental attempt to build a **complete version control system from scratch** — including:

* A custom **CLI client**
* A custom **object storage model**
* A full **commit and tree object pipeline**
* And eventually a **remote hosting platform**, similar to GitHub or Gitea, to store repositories

After progressing deep into the core architecture (object hashing, directory tree modeling, commit structure, command registry, gob-based storage, etc.), the reality became clear:

### ❗ Creating a full VCS requires:

* A fully-developed command-line interface
* A stable networking/push/pull layer
* A server platform for hosting repositories
* A branching/merging system
* Diff algorithms, index/staging logic
* Countless edge cases and recovery mechanisms
* And a huge development investment

This is not something I can continue allocating time to right now.

---

# 🌿 Successor Project: **GitGroove**

The goals behind ForestVC are **not abandoned** — they are simply moving to a more realistic and more powerful foundation.

Introducing:

## 👉 **[GitGroove](https://github.com/kuchuk-borom-debbarma/GitGrove)**

GitGroove is the **official continuation** of the same mission:
✔ Handling **multiple nested repositories inside a single repo**
✔ Managing multi-repo workflows cleanly
✔ Providing simple, intuitive commands
✔ Fully powered by **Git**, not replacing it

GitGroove is:

* **Easier to build**
* **More reliable**
* **Actually useful to developers**
* **Maintains the spirit of ForestVC’s goals**

If you came here looking for that multi-repo functionality —
🔄 **please switch to GitGroove instead**.

---

# 📁 ForestVC Code Overview (Frozen)

This repo contains an educational implementation of:

### ✓ Core VCS internals:

* SHA-256 object hashing
* Git-style object storage layout (`xx/yyyy`)
* Blob/tree/commit object generation
* Directory tree hashing (deep-to-root)

### ✓ Repository setup:

* `.fvc` directory structure
* Metadata storage
* HEAD pointer management

### ✓ Command system:

* Dynamic command registry
* Implemented commands: `init`, `commit`

### ✓ Utilities:

* Path normalization
* Binary/text file writers
* Gob-based serialization
* Directory creation & filesystem helpers

The code is left intact as a learning resource.

---

# 📌 Project Status

> **ForestVC is no longer under active development.**
> For the continued and improved version of this idea, please use **GitGroove**.

---

# 📜 License

MIT License — feel free to experiment or reuse code for educational purposes.

---

If you want, I can also:

