# HTTP/1.1 Server from Scratch

Ever wonder what actually happens when you type a URL into your browser? This project is my attempt to answer that question by building an HTTP server from the ground up, using nothing but raw TCP sockets in Go.

No `net/http`. No frameworks. Just bytes on a wire.

Built as part of [Boot.dev's "Build an HTTP Server" course](https://boot.dev), though I've made a bunch of changes and fixes along the way.

## What This Does

At its core, this is an HTTP/1.1 server that:

- Parses incoming HTTP requests from raw TCP streams
- Generates proper HTTP responses
- Supports chunked transfer encoding
- Can proxy requests to other servers (with trailers!)

The fun part is that it all happens incrementally. The parser doesn't wait for the full request to arrive - it processes data as it comes in, which is how real servers handle slow or unreliable connections.

## How It's Built

### Request Parsing

The request parser uses a state machine:

```
StateInit → StateHeaders → StateBody → StateDone
```

Each state consumes whatever data is available and hands off to the next. This means we can parse requests byte-by-byte if needed, which is useful for handling those weird edge cases that show up in real network traffic.

### Response Writing

Responses are built using a `Writer` that tracks its own state. You write the status line, then headers, then body - and it handles the formatting for you. It also supports:

- Chunked responses (for streaming data)
- Trailers (headers that come *after* the body, which is pretty cool)

### The Proxy Example

There's a working proxy at `/httpbin/html` that fetches content from httpbin.org and streams it back with:

- Chunked encoding (since we don't know the content length upfront)
- SHA256 hash and content length as trailers

So the client gets the headers they need *after* the body finishes streaming. Try it:

```bash
curl --raw http://localhost:42069/httpbin/html
```

You'll see the chunk sizes in hex, the body, then the trailers at the end.

## Getting Started

### Prerequisites

Go 1.21 or higher.

### Running the Server

```bash
go run ./cmd/httpserver/main.go
```

The server listens on port 42069.

### Trying It Out

```bash
# Basic request
curl http://localhost:42069/

# See a 400 error
curl http://localhost:42069/yourproblem

# See a 500 error
curl http://localhost:42069/myproblem

# Proxy with chunked encoding and trailers
curl --raw http://localhost:42069/httpbin/html
```

### Running Tests

```bash
go test ./...
```

Tests cover the usual stuff plus some trickier cases like byte-by-byte reading (simulating slow networks) and malformed requests.

## Project Structure

```
.
├── cmd/
│   └── httpserver/      # Main server with routes
├── internal/
│   ├── request/         # Request parsing (state machine)
│   ├── response/        # Response writing + chunking
│   ├── headers/         # Header parsing logic
│   └── server/          # TCP server boilerplate
└── README.md
```

## Things I Found Along The Way

Working through this, I ran into some differences between the course material and what the RFCs actually say:

- **Header whitespace**: The course allowed leading whitespace in headers, but RFC 9112 explicitly forbids this (it's leftover from HTTP/1.0 line folding)
- **Field name validation**: No whitespace allowed before the colon in header names
- **EOF handling**: Had a bug where the final chunk plus EOF wasn't being processed correctly

All reported back to Boot.dev, so hopefully the course is better for the next person.

## What's Not Here

This is still an educational project, so some things are intentionally missing:

- HTTP/2 or HTTP/3 (that's a whole other adventure)
- Concurrent connections (single-threaded for simplicity)
- Proper routing (just a few hardcoded paths)
- Chunked *request* bodies (only responses)

If you need any of those, you're probably better off with Go's standard library or a real framework.

## What I Learned

- How text-based protocols are structured and why they look the way they do
- State machines are actually useful (who knew)
- Reading RFCs is painful but necessary
- The gap between "it works in my tests" and "it works on the real internet" is bigger than I thought
- `io.Reader` is a beautiful interface

## Acknowledgments

- [Boot.dev](https://boot.dev) for the course that got me started
- RFCs [9110](https://www.rfc-editor.org/rfc/rfc9110.html) and [9112](https://www.rfc-editor.org/rfc/rfc9112.html) for the nitty-gritty details
- The Go team for making networking actually pleasant

## License

MIT - do whatever you want with it.
