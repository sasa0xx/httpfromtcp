package response

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int
type WriterState int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

const (
	StateInit WriterState = iota
	StateStatusLine
	StateHeaders
	StateBody
)

type Writer struct {
	buf     *bytes.Buffer
	body    *bytes.Buffer
	state   WriterState
	headers headers.Headers
}

var codeNames = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusBadRequest:          "Bad Request",
	StatusInternalServerError: "Internal Server Error",
}

func (c StatusCode) String() string {
	if val, ok := codeNames[c]; ok {
		return val
	}
	return "Unknown code"
}

func NewWriter() *Writer {
	return &Writer{
		buf:     &bytes.Buffer{},
		body:    &bytes.Buffer{},
		state:   StateInit,
		headers: GetDefaultHeaders(0),
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	w.state = StateStatusLine
	_, err := fmt.Fprintf(w.buf, "HTTP/1.1 %d %s\r\n", statusCode, statusCode.String())
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state == StateBody {
		return fmt.Errorf("You can only write headers once, when you write the status line.")
	}
	w.state = StateHeaders
	w.headers = headers

	return nil
}

func (w *Writer) WriteBody(p []byte) error {
	if w.state != StateHeaders && w.state != StateBody {
		return fmt.Errorf("You need to write headers first.")
	}
	w.state = StateBody

	_, err := w.body.Write(p)
	return err
}

func (w *Writer) Bytes() []byte {
	body := w.body.Bytes()
	w.headers.Set("content-length", fmt.Sprintf("%d", len(body)))
	for k, v := range w.headers {
		fmt.Fprintf(w.buf, "%s: %s\r\n", k, v)
	}
	w.buf.Write([]byte("\r\n"))
	w.buf.Write(body)
	return w.buf.Bytes()
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, statusCode.String())
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("content-length", fmt.Sprintf("%d", contentLen))
	h.Set("connection", "close")
	h.Set("content-type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
