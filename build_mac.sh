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

echo "📁 Creating sensible testRepo..."

TEST_REPO="build/testRepo"
SUB_REPO="$TEST_REPO/subRepo"

mkdir -p "$TEST_REPO/internal/math"
mkdir -p "$TEST_REPO/pkg/greetings"
mkdir -p "$TEST_REPO/assets"

# Move built mrvc into testRepo root
cp build/mrvc "$TEST_REPO/mrvc"


# -------------------------------
# Create .mrvcignore
# -------------------------------

cat > "$TEST_REPO/.mrvcignore" <<EOF
# Ignore build artifacts
build/

# Ignore temporary files
*.tmp
*.log

# Ignore Go vendor folder
vendor/

# Ignore macOS metadata files
.DS_Store
EOF

# -------------------------------
# Create sample files
# -------------------------------

cat > "$TEST_REPO/README.md" <<EOF
# TestRepo

A small example Go project used for testing the MultiRepoVC version control system.
EOF

cat > "$TEST_REPO/app.go" <<EOF
package main

import (
    "fmt"
    "testRepo/internal/math"
    "testRepo/pkg/greetings"
)

func main() {
    fmt.Println("Hello from TestRepo!")
    fmt.Println("2 + 3 =", math.Add(2, 3))
    fmt.Println(greetings.Hello("Kuku"))
}
EOF

cat > "$TEST_REPO/internal/math/add.go" <<EOF
package math

func Add(a, b int) int {
    return a + b
}
EOF

cat > "$TEST_REPO/internal/math/multiply.go" <<EOF
package math

func Multiply(a, b int) int {
    return a * b
}
EOF

cat > "$TEST_REPO/pkg/greetings/hello.go" <<EOF
package greetings

func Hello(name string) string {
    return "Hello, " + name + "!"
}
EOF

cat > "$TEST_REPO/assets/sample.txt" <<EOF
This is a sample asset file for snapshot testing with MultiRepoVC.
EOF

echo "📁 TestRepo created."

# ================================================================================
# CREATE NESTED REPO
# ================================================================================

echo "📁 Creating nested MRVC repository inside testRepo/subRepo..."

mkdir -p "$SUB_REPO"

# Move mrvc binary into nested repo
cp build/mrvc "$SUB_REPO/mrvc"

# Initialize nested repo structure
mkdir -p "$SUB_REPO/src"
mkdir -p "$SUB_REPO/assets"

cat > "$SUB_REPO/README.md" <<EOF
# SubRepo

A nested MRVC repository inside TestRepo.
Used to test hierarchical versioning.
EOF

cat > "$SUB_REPO/src/module.go" <<EOF
package src

func SubValue() int {
    return 42
}
EOF

cat > "$SUB_REPO/assets/info.txt" <<EOF
This is a nested repo asset inside subRepo.
EOF

# Create subRepo .mrvcignore
cat > "$SUB_REPO/.mrvcignore" <<EOF
*.tmp
ignore-this/
EOF

echo "📁 Initializing both repos with mrvc..."

# Initialize parent repo
(
    cd "$TEST_REPO"
    ./mrvc init testRepo "Builder Script"
)

# Initialize child repo
(
    cd "$SUB_REPO"
    ./mrvc init subRepo "Builder Script"
)

echo "🎉 Nested repository created at: $SUB_REPO"
echo "🎉 MRVC binary copied into both repos"
echo "🎉 Build complete!"
