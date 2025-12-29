# Default recipe
default: build

# Version ldflags for injecting version at build time
version_pkg := "github.com/benjaminabbitt/claude-limits/internal/version"
ldflags := "-X " + version_pkg + ".Version=$(versionator version)"

# Build the binary to bin/
build:
    @mkdir -p bin
    go build -ldflags "{{ldflags}}" -o bin/claude-limits ./cmd/claude-limits

# Run go mod tidy
tidy:
    go mod tidy

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f claude-limits

# Run the limits command
limits *ARGS:
    @bin/claude-limits limits {{ARGS}}

# Run the MCP server
serve:
    @bin/claude-limits serve

# Install to $GOPATH/bin
install:
    go install -ldflags "{{ldflags}}" ./cmd/claude-limits

# Run tests
test:
    go test ./...

# Build for multiple platforms
release:
    @mkdir -p bin
    GOOS=linux GOARCH=amd64 go build -ldflags "{{ldflags}}" -o bin/claude-limits-linux-amd64 ./cmd/claude-limits
    GOOS=darwin GOARCH=amd64 go build -ldflags "{{ldflags}}" -o bin/claude-limits-darwin-amd64 ./cmd/claude-limits
    GOOS=darwin GOARCH=arm64 go build -ldflags "{{ldflags}}" -o bin/claude-limits-darwin-arm64 ./cmd/claude-limits
    GOOS=windows GOARCH=amd64 go build -ldflags "{{ldflags}}" -o bin/claude-limits-windows-amd64.exe ./cmd/claude-limits
