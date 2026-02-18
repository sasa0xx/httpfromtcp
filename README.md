# HTTP/1.1 Server from Scratch

A ground-up implementation of an HTTP/1.1 request parser built directly on top of TCP, written in Go. This project demonstrates how HTTP actually works under the hood by parsing the protocol without using Go's `net/http` package.

Built as part of [Boot.dev's "Build an HTTP Server" course](https://boot.dev), with additional RFC compliance improvements and bug fixes.

## Features

- **Streaming Request Parser** - Efficiently handles incomplete data with a state machine architecture
- **RFC 9110/9112 Compliant** - Strictly follows HTTP/1.1 specifications
- **Request Line Parsing** - Validates HTTP methods, paths, and protocol versions
- **Header Parsing** - Case-insensitive header names with proper whitespace handling
- **Body Parsing** - Content-Length based body reading
- **Comprehensive Testing** - Edge cases covered with configurable chunk sizes

## What Makes This Different

Most developers use HTTP libraries without understanding the underlying protocol. This project:
- Parses raw TCP byte streams into HTTP requests
- Implements proper state machine for incremental parsing
- Handles real-world edge cases (incomplete reads, malformed requests)
- Works with actual HTTP clients like `curl`

## Technical Highlights

### Streaming Parser
The parser doesn't require the entire request to be in memory. It processes data as it arrives over TCP:
```go
// Handles requests even when received 1-2 bytes at a time
reader := &chunkReader{
    data:            "GET /path HTTP/1.1\r\nHost: localhost\r\n\r\n",
    numBytesPerRead: 2,  // Simulates slow network
}
```

### State Machine Architecture
```
StateInit → StateHeaders → StateBody → StateDone
```

Each state consumes available data and transitions when ready, allowing efficient parsing of incomplete requests.

### RFC Compliance Notes

During development, I found several discrepancies between the course material and the actual RFC specifications:

- **Header whitespace**: The course allowed leading whitespace in headers, but RFC 9112 forbids this (it's obsolete line folding from HTTP/1.0)
- **Field name validation**: Stricter enforcement of no whitespace before colons
- **EOF handling**: Fixed bug where final data chunk + EOF wasn't being parsed

All reported to Boot.dev for course improvement.

## Installation & Usage

### Prerequisites
- Go 1.21 or higher

### Running the Server
```bash
# Clone the repository
git clone https://github.com/yourusername/http-from-tcp
cd http-from-tcp

# Run the server
go run ./cmd/tcplistener
```

The server listens on `localhost:42069` by default.

### Testing with curl
```bash
# Simple GET request
curl http://localhost:42069/test

# POST request with body
curl -X POST http://localhost:42069/api/data \
  -H "Content-Type: application/json" \
  -d '{"hello":"world"}'
```

### Running Tests
```bash
go test ./...
```

Tests include edge cases like:
- Byte-by-byte reading (simulating slow networks)
- Malformed requests
- Multiple headers with the same name
- Binary body data
- Invalid Content-Length values

## Project Structure
```
.
├── cmd/
│   └── tcplistener/     # Main server executable
├── internal/
│   ├── request/         # Request parser with state machine
│   └── headers/         # Header parsing logic
└── README.md
```

## How It Works

### 1. TCP Connection
```go
listener, err := net.Listen("tcp", ":42069")
conn, err := listener.Accept()
```

### 2. Streaming Parse
```go
req, err := request.RequestFromReader(conn)
// Incrementally reads and parses:
// - Request line (method, path, version)
// - Headers (field-name: field-value pairs)
// - Body (if Content-Length present)
```

### 3. Output
```go
fmt.Printf("Method: %s\n", req.RequestLine.Method)
fmt.Printf("Path: %s\n", req.RequestLine.RequestTarget)
fmt.Printf("Headers: %v\n", req.Headers)
fmt.Printf("Body: %s\n", req.Body)
```

## What I Learned

- **Protocol Design**: How text-based protocols like HTTP are structured
- **Streaming Parsers**: Building state machines to handle incomplete data
- **RFC Reading**: Interpreting technical specifications (RFCs 9110, 9112)
- **Edge Cases**: Real-world networking challenges (partial reads, EOF handling)
- **Go Interfaces**: Leveraging `io.Reader` for flexible input sources
- **Testing**: Writing comprehensive tests with custom mock readers

## Current Limitations

- Only supports HTTP/1.1 (not HTTP/2 or HTTP/3)
- No chunked transfer encoding support
- No response generation (parser only)
- Single-threaded (no concurrent connection handling yet)
- Assumes `Content-Length` for body size (no other methods)

These are intentional - the goal was to understand HTTP fundamentals, not build a production server.

## Future Improvements

- [ ] HTTP response generation
- [ ] Request routing by path/method
- [ ] Concurrent connection handling with goroutines
- [ ] Chunked transfer encoding support
- [ ] Static file serving
- [ ] Middleware architecture

## Acknowledgments

- Built following [Boot.dev's HTTP Server course](https://boot.dev)
- RFCs: [9110 (HTTP Semantics)](https://www.rfc-editor.org/rfc/rfc9110.html), [9112 (HTTP/1.1)](https://www.rfc-editor.org/rfc/rfc9112.html)
- Thanks to the Go community for excellent networking primitives

## License

MIT License - feel free to learn from and build upon this code.

---

**Note**: This is an educational project. For production use, always use battle-tested libraries like Go's `net/http` package.
