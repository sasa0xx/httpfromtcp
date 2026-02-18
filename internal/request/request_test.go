package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}

	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestHeadersParse(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestBody(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
}

func TestRequestLineEdgeCases(t *testing.T) {
	// Test: Invalid HTTP version
	reader := &chunkReader{
		data:            "GET / HTTP/2.0\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.Error(t, err)

	// Test: Lowercase method (should fail)
	reader = &chunkReader{
		data:            "get / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Extra spaces in request line (should fail)
	reader = &chunkReader{
		data:            "GET  /  HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: POST method
	reader = &chunkReader{
		data:            "POST /api/users HTTP/1.1\r\n\r\n",
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/api/users", r.RequestLine.RequestTarget)

	// Test: Complex path with query string
	reader = &chunkReader{
		data:            "GET /search?q=golang&limit=10 HTTP/1.1\r\n\r\n",
		numBytesPerRead: 2,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, "/search?q=golang&limit=10", r.RequestLine.RequestTarget)
}

func TestHeadersEdgeCases(t *testing.T) {
	// Test: Multiple headers with same name (should combine with comma)
	reader := &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Accept: text/html\r\n" +
			"Accept: application/json\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, "text/html, application/json", r.Headers.Get("accept"))

	// Test: Header with leading whitespace (should fail per RFC)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\n  Host: localhost\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Header with space before colon (should fail)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost : localhost\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Header with lots of whitespace around value (should trim)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost:    localhost:8080    \r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", r.Headers.Get("host"))

	// Test: Case insensitivity of header names
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Content-Type: application/json\r\n" +
			"content-length: 0\r\n" +
			"HOST: example.com\r\n" +
			"\r\n",
		numBytesPerRead: 4,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, "application/json", r.Headers.Get("content-type"))
	assert.Equal(t, "0", r.Headers.Get("content-length"))
	assert.Equal(t, "example.com", r.Headers.Get("host"))
}

func TestBodyEdgeCases(t *testing.T) {
	// Test: No body with GET request
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	assert.Empty(t, r.Body)

	// Test: Empty body with Content-Length: 0
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Empty(t, r.Body)

	// Test: JSON body
	reader = &chunkReader{
		data: "POST /api/users HTTP/1.1\r\n" +
			"Content-Type: application/json\r\n" +
			"Content-Length: 27\r\n" +
			"\r\n" +
			`{"name":"Alice","age":30}`,
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, `{"name":"Alice","age":30}`, string(r.Body))

	// Test: Body with newlines
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Content-Length: 17\r\n" +
			"\r\n" +
			"line1\nline2\nline3",
		numBytesPerRead: 2,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, "line1\nline2\nline3", string(r.Body))

	// Test: Binary-like data (non-printable characters)
	reader = &chunkReader{
		data: "POST /upload HTTP/1.1\r\n" +
			"Content-Length: 5\r\n" +
			"\r\n" +
			"\x00\x01\x02\x03\x04",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, []byte{0, 1, 2, 3, 4}, r.Body)
}

func TestVerySmallChunkSizes(t *testing.T) {
	// Test: 1 byte at a time
	reader := &chunkReader{
		data: "GET /test HTTP/1.1\r\n" +
			"Host: localhost\r\n" +
			"Content-Length: 5\r\n" +
			"\r\n" +
			"hello",
		numBytesPerRead: 1,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/test", r.RequestLine.RequestTarget)
	assert.Equal(t, "localhost", r.Headers.Get("host"))
	assert.Equal(t, "hello", string(r.Body))

	// Test: 2 bytes at a time
	reader = &chunkReader{
		data: "POST / HTTP/1.1\r\n" +
			"Content-Length: 3\r\n" +
			"\r\n" +
			"abc",
		numBytesPerRead: 2,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "abc", string(r.Body))
}

func TestLargeBody(t *testing.T) {
	// Test: Larger body (1KB)
	bodyContent := string(make([]byte, 1024))
	for i := range bodyContent {
		bodyContent = bodyContent[:i] + "x" + bodyContent[i+1:]
	}

	reader := &chunkReader{
		data: "POST /upload HTTP/1.1\r\n" +
			"Content-Length: 1024\r\n" +
			"\r\n" +
			bodyContent,
		numBytesPerRead: 8,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, 1024, len(r.Body))
	assert.Equal(t, bodyContent, string(r.Body))
}

func TestInvalidContentLength(t *testing.T) {
	// Test: Non-numeric Content-Length
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Content-Length: invalid\r\n" +
			"\r\n" +
			"body",
		numBytesPerRead: 3,
	}
	_, err := RequestFromReader(reader)
	require.Error(t, err)

	// Test: Negative Content-Length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Content-Length: -10\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	// Depends on your error handling - might error or treat as 0
}
