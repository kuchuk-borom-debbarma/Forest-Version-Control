#!/bin/bash

set -e  # stop on first error

echo "🔨 Building mrvc for macOS..."

# Create build directory
mkdir -p build

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    echo "🧠 Detected Apple Silicon (arm64)"
    GOARCH=arm64
else
    echo "🖥 Detected Intel (amd64)"
    GOARCH=amd64
fi

# Build CLI binary
GOOS=darwin \
GOARCH=$GOARCH \
go build -o build/mrvc ./src/cmd/mrvc

echo "📁 Creating nested repo hierarchy..."

ROOT="build/RootRepo"
CHILD_A="$ROOT/ChildA"
CHILD_B="$ROOT/ChildB"
CHILD_AA="$CHILD_A/ChildAA"

# Create folder structure
mkdir -p "$ROOT" "$CHILD_A" "$CHILD_B" "$CHILD_AA"

# Copy mrvc into all repos
for repo in "$ROOT" "$CHILD_A" "$CHILD_B" "$CHILD_AA"; do
    cp build/mrvc "$repo/mrvc"
done

# Function to create sample files
create_repo_files() {
    local REPO_PATH="$1"
    local NAME="$2"

    mkdir -p "$REPO_PATH/src"
    mkdir -p "$REPO_PATH/assets"

    cat > "$REPO_PATH/README.md" <<EOF
# $NAME
This is repository $NAME created for hierarchical MRVC testing.
EOF

    cat > "$REPO_PATH/src/module.go" <<EOF
package src

func Value() string {
    return "Hello from $NAME"
}
EOF

    cat > "$REPO_PATH/assets/sample.txt" <<EOF
Asset file for $NAME.
EOF

    # default ignore rules
    cat > "$REPO_PATH/.mrvcignore" <<EOF
*.tmp
*.log
EOF
}

# Create files
create_repo_files "$ROOT" "RootRepo"
create_repo_files "$CHILD_A" "ChildA"
create_repo_files "$CHILD_B" "ChildB"
create_repo_files "$CHILD_AA" "ChildAA"

echo "📁 Initializing all repos..."

(
    cd "$ROOT"
    ./mrvc init --name RootRepo --author "Builder Script"
)

(
    cd "$CHILD_A"
    ./mrvc init --name ChildA --author "Builder Script"
)

(
    cd "$CHILD_B"
    ./mrvc init --name ChildB --author "Builder Script"
)

(
    cd "$CHILD_AA"
    ./mrvc init --name ChildAA --author "Builder Script"
)

echo "🔗 Linking repos..."

# RootRepo -> ChildA + ChildB
(
    cd "$ROOT"
    ./mrvc link --path ChildA
    ./mrvc link --path ChildB
)

# ChildA -> ChildAA
(
    cd "$CHILD_A"
    ./mrvc link --path ChildAA
)

echo "📸 Performing first commits..."

(
    cd "$ROOT"
    ./mrvc commit --message "first commit for RootRepo" --author "Builder Script" --files "*"
)

(
    cd "$CHILD_A"
    ./mrvc commit --message "first commit for ChildA" --author "Builder Script" --files "*"
)

(
    cd "$CHILD_B"
    ./mrvc commit --message "first commit for ChildB" --author "Builder Script" --files "*"
)

(
    cd "$CHILD_AA"
    ./mrvc commit --message "first commit for ChildAA" --author "Builder Script" --files "*"
)

echo "🌲 Creating hierarchical super commits..."

# SUPER COMMIT FOR ChildA → includes ChildAA
(
    cd "$CHILD_A"
    ./mrvc super-commit --message "super commit for ChildA" --author "Builder Script"
)

# SUPER COMMIT FOR RootRepo → includes ChildA (super) + ChildB (commit)
(
    cd "$ROOT"
    ./mrvc super-commit --message "super commit for RootRepo" --author "Builder Script"
)

echo ""
echo "🎉 All repos initialized, linked, committed, and super committed!"
echo "📂 Final Structure:"
echo "RootRepo/"
echo " ├── ChildA/   (super committed)"
echo " │     └── ChildAA/ (normal committed)"
echo " └── ChildB/   (normal committed)"
echo ""
echo "🚀 Build complete!"
