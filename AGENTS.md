# AGENTS.md - Development Guidelines for httpfromtcp

This file contains essential information for AI coding agents working on the httpfromtcp project. Follow these guidelines to maintain code quality and consistency.

## Build, Lint, and Test Commands

### Basic Commands
```bash
# Build the project
go build .

# Build all packages
go build ./...

# Run the program
go run main.go
```

### Testing Commands
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run a specific test (when tests are added)
go test -run TestFunctionName ./...

# Run tests in a specific package
go test ./path/to/package

# Run benchmarks
go test -bench=. ./...
```

### Linting and Code Quality
```bash
# Format code
go fmt ./...

# Check for common issues
go vet ./...

# Install and run golangci-lint (recommended linter suite)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run

# Install and run goimports (formats imports)
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .

# Check for security issues
go install github.com/securecodewarrior/govulncheck@latest
govulncheck ./...
```

### Module Management
```bash
# Tidy dependencies
go mod tidy

# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Clean module cache
go clean -modcache
```

## Code Style Guidelines

### General Principles
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for consistent formatting
- Prefer clarity over cleverness
- Keep functions short and focused
- Use meaningful names for variables, functions, and types

### Package Naming
- Use lowercase, single-word names
- Avoid package names like `util`, `common`, or `misc`
- Package name should match the directory name
- Example: `package httpfromtcp`

### File Naming
- Use lowercase with underscores: `file_reader.go`
- Test files: `file_reader_test.go`
- Don't use camelCase or PascalCase for filenames

### Imports
```go
// Standard library imports first
import (
    "bufio"
    "fmt"
    "io"
    "os"
    "strings"
)

// Blank line separates standard library from third-party
import (
    "github.com/pkg/errors"
    "golang.org/x/net/context"
)

// Local imports last
import (
    "yourproject/internal/config"
    "yourproject/pkg/utils"
)
```

- Group imports by category with blank lines
- Use factored import statements (parentheses)
- Remove unused imports
- Use import aliases only when necessary to avoid conflicts

### Variable and Constant Naming
```go
// Good
var userName string
var maxRetries int
const defaultPort = 8080

// Bad - unclear or too generic
var x string
var data int
var DEFAULT_PORT = 8080
```

- Use camelCase for variables and functions
- Use PascalCase for exported identifiers
- Use ALL_CAPS for constants
- Be descriptive but not verbose

### Function Naming
```go
// Good
func parseConfig(filePath string) (*Config, error)
func validateInput(input string) bool

// Bad
func doStuff(s string) error
func process() *Config
```

- Start with a verb
- Use PascalCase for exported functions
- camelCase for unexported functions
- Return errors as the last return value

### Error Handling
```go
// Good - immediate check after assignment
data, err := readFile(filename)
if err != nil {
    return fmt.Errorf("failed to read file %s: %w", filename, err)
}

// Good - wrap errors with context
if err := processData(data); err != nil {
    return errors.Wrap(err, "failed to process data")
}

// Bad - don't ignore errors
data, _ := readFile(filename) // Don't do this

// Bad - don't use panic for expected errors
if err != nil {
    panic(err)
}
```

- Always check for errors immediately after they occur
- Don't ignore errors with `_`
- Add context to errors using `fmt.Errorf` with `%w` verb
- Use `errors.Wrap` from `github.com/pkg/errors` if available
- Return errors, don't panic for expected error conditions
- Use `log.Fatal` or `panic` only for unrecoverable errors

### Type Declarations
```go
// Good
type Config struct {
    Host     string
    Port     int
    Timeout  time.Duration
    Enabled  bool
}

type Server interface {
    Start() error
    Stop() error
}

type handlerFunc func(w http.ResponseWriter, r *http.Request)

// Bad - unclear or generic names
type C struct {
    H string
    P int
    T time.Duration
    E bool
}
```

- Use PascalCase for type names
- Keep types simple and focused
- Use interfaces to define contracts
- Prefer structs over maps for structured data

### Control Structures
```go
// Good - early return for errors
func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    // Process file...
    return nil
}

// Good - switch with fallthrough when needed
switch status {
case "pending":
    // handle pending
case "processing":
    // handle processing
default:
    return errors.New("unknown status")
}

// Bad - nested if statements
if file != nil {
    if !file.Closed() {
        // do something
    }
}
```

- Use early returns to reduce nesting
- Prefer switch over long if-else chains
- Use defer for cleanup operations
- Avoid complex nested conditions

### Comments
```go
// Good - explains why, not what
// validatePort checks if the port number is within the valid range
// to prevent security issues with privileged ports
func validatePort(port int) error {
    // Implementation...
}

// Bad - redundant comment
// This function adds two numbers
func add(a, b int) int {
    return a + b
}
```

- Write comments for exported functions/types
- Explain why, not what the code does
- Keep comments up to date
- Use complete sentences starting with capital letters

### Testing Guidelines
```go
func TestParseConfig(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        want     *Config
        wantErr  bool
    }{
        {
            name:  "valid config",
            input: `{"host": "localhost", "port": 8080}`,
            want:  &Config{Host: "localhost", Port: 8080},
            wantErr: false,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseConfig(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseConfig() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("parseConfig() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

- Use table-driven tests
- Name test functions as `TestFunctionName`
- Use `t.Run` for subtests
- Test both success and error cases
- Use meaningful test names and inputs

### Project Structure
```
httpfromtcp/
├── main.go              # Main application entry point
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── README.md            # Project documentation
├── AGENTS.md            # This file
├── internal/            # Private application code
├── pkg/                 # Library code meant to be used by others
└── cmd/                 # Command-line applications
```

### Performance Considerations
- Use `strings.Builder` for string concatenation in loops
- Prefer slices over arrays when size is unknown
- Use `sync.Pool` for expensive object reuse
- Profile with `go tool pprof` before optimizing
- Use buffered channels when appropriate

### Security Best Practices
- Validate all inputs
- Use `html.EscapeString` for HTML output
- Avoid SQL injection with prepared statements
- Use HTTPS for network communications
- Don't log sensitive information
- Use `crypto/rand` for cryptographic randomness

Remember: These guidelines ensure consistency and maintainability. When in doubt, look at the existing codebase and follow the established patterns.