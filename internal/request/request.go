package request

import (
	"bytes"
	"errors"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"unicode"
	"unicode/utf8"
)

var CRFL []byte = ([]byte)("\r\n")
var ERROR_MALFORMED_HTTP_REQUEST error = errors.New("Malformed Http Request Error")
var ERROR_MALFORMED_REQUEST_LINE error = errors.New("Malformed Request Line Error")
var ERROR_READING_IN_DONE_STATE error = errors.New("Trying to read in a done state")
var ERROR_UNDIFINIED_STATE error = errors.New("Undifiened state")

type ParserState int

const ( // Best replacement for an "enum"
	StateInit ParserState = iota
	StateHeaders
	StateBody
	StateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	State       ParserState
	Headers     headers.Headers
	Body        []byte
}

func newRequest() *Request {
	return &Request{
		State:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

func isAllUpper(str []byte) bool {
	if len(str) == 0 {
		return false
	}

	for i := 0; i < len(str); {
		r, size := utf8.DecodeRune(str[i:])
		if !unicode.IsUpper(r) {
			return false
		}
		i += size
	}
	return true
}

func ParseRequestLine(data []byte) (*RequestLine, int, error) {
	splits := bytes.SplitN(data, CRFL, 2)
	if len(splits) != 2 {
		return nil, 0, nil // returns no error if it hadn't been completed yet.
	}
	rline := bytes.TrimSpace(splits[0])
	parts := bytes.Split(rline, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}
	if !isAllUpper(parts[0]) {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}
	if string(parts[2]) != "HTTP/1.1" {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	return &RequestLine{
		HttpVersion:   "1.1",
		RequestTarget: string(parts[1]),
		Method:        string(parts[0]),
	}, len(splits[0]) + len(CRFL), nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case StateInit:
		rq, read, err := ParseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if read == 0 {
			return 0, nil
		}
		r.State = StateHeaders
		r.RequestLine = *rq
		return read, nil

	case StateHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.State = StateBody
		}

		return n, nil

	case StateBody:
		l := r.Headers.Get("content-length")
		if l == "" {
			r.State = StateDone
			return 0, nil
		}
		length, err := strconv.Atoi(l)
		if err != nil {
			return 0, err
		}
		if length < 0 {
			return 0, errors.New("invalid content-length")
		}

		remaining := min(length-len(r.Body), len(data))
		r.Body = append(r.Body, data[:remaining]...)
		if len(r.Body) == length {
			r.State = StateDone
		}
		return remaining, nil

	case StateDone:
		return 0, ERROR_READING_IN_DONE_STATE
	}

	return 0, ERROR_UNDIFINIED_STATE
}

func (r *Request) parse(data []byte) (int, error) {
	consumed := 0
	for r.State != StateDone {
		n, err := r.parseSingle(data[consumed:])
		consumed += n
		if err != nil {
			return 0, err // I really feel like returning consumed instead of 0 makes more sense,
			// but let's follow the course I guess.
		}
		if n == 0 {
			break
		}
	}

	return consumed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	rq := newRequest()
	buf := make([]byte, 8)
	readToIndex := 0

	for rq.State != StateDone {
		if readToIndex >= len(buf) {
			nbuf := make([]byte, len(buf)*2)
			copy(nbuf, buf)
			buf = nbuf
		}

		n, err := reader.Read(buf[readToIndex:])
		readToIndex += n
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, err
			}
			if n == 0 {
				rq.State = StateDone
				break
			}
		}

		read, err := rq.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[read:readToIndex])
		readToIndex -= read
	}
	return rq, nil
}
