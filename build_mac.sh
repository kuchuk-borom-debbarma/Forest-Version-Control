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
mkdir -p "$TEST_REPO/internal/math"
mkdir -p "$TEST_REPO/pkg/greetings"
mkdir -p "$TEST_REPO/assets"

# .mrvcignore
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

# README.md
cat > "$TEST_REPO/README.md" <<EOF
# TestRepo

A small example Go project used for testing the MultiRepoVC version control system.
EOF

# app.go
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

# internal/math/add.go
cat > "$TEST_REPO/internal/math/add.go" <<EOF
package math

func Add(a, b int) int {
    return a + b
}
EOF

# internal/math/multiply.go
cat > "$TEST_REPO/internal/math/multiply.go" <<EOF
package math

func Multiply(a, b int) int {
    return a * b
}
EOF

# pkg/greetings/hello.go
cat > "$TEST_REPO/pkg/greetings/hello.go" <<EOF
package greetings

func Hello(name string) string {
    return "Hello, " + name + "!"
}
EOF

# assets/sample.txt
cat > "$TEST_REPO/assets/sample.txt" <<EOF
This is a sample asset file for snapshot testing with MultiRepoVC.
EOF

echo "✅ testRepo created at build/testRepo/"
echo "🎉 Build complete: build/mrvc"
