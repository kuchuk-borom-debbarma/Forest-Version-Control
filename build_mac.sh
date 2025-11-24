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
mkdir -p "$CHILD_AA" "$CHILD_A" "$CHILD_B" "$ROOT"

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

create_repo_files "$ROOT" "RootRepo"
create_repo_files "$CHILD_A" "ChildA"
create_repo_files "$CHILD_B" "ChildB"
create_repo_files "$CHILD_AA" "ChildAA"

echo "📁 Initializing all repos..."

( cd "$ROOT" && ./mrvc init RootRepo "Builder Script" )
( cd "$CHILD_A" && ./mrvc init ChildA "Builder Script" )
( cd "$CHILD_B" && ./mrvc init ChildB "Builder Script" )
( cd "$CHILD_AA" && ./mrvc init ChildAA "Builder Script" )

echo "🔗 Linking repos..."

# RootRepo -> ChildA, ChildB
(
    cd "$ROOT"
    ./mrvc link ChildA
    ./mrvc link ChildB
)

# ChildA -> ChildAA
(
    cd "$CHILD_A"
    ./mrvc link ChildAA
)

echo "📸 Performing first commits..."

( cd "$ROOT"     && ./mrvc commit "first commit for RootRepo" "Builder Script" "*" )
( cd "$CHILD_A"  && ./mrvc commit "first commit for ChildA"   "Builder Script" "*" )
( cd "$CHILD_B"  && ./mrvc commit "first commit for ChildB"   "Builder Script" "*" )
( cd "$CHILD_AA" && ./mrvc commit "first commit for ChildAA"  "Builder Script" "*" )

echo ""
echo "🎉 All repos initialized, linked, and committed!"
echo "📂 Final Structure:"
echo "RootRepo/"
echo " ├── ChildA/"
echo " │     └── ChildAA/"
echo " └── ChildB/"
echo ""
echo "🚀 Build complete."
