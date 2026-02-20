package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"net"
	"sync/atomic"
)

type ServerState = int

const (
	StateInit ServerState = iota
	StateClosed
)

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	code    response.StatusCode
	message string
}

func NewHandlerError(code response.StatusCode, message string) *HandlerError {
	return &HandlerError{
		code:    code,
		message: message,
	}
}

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func (h *HandlerError) Write(w io.Writer) error {
	response.WriteStatusLine(w, h.code)
	response.WriteHeaders(w, response.GetDefaultHeaders(len(h.message)))
	_, err := fmt.Fprintf(w, "\r\n%s", h.message)
	return err
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		e := NewHandlerError(response.StatusBadRequest, err.Error()) // the error text we defined in request package
		e.Write(conn)
		return
	}

	writer := response.NewWriter()
	s.handler(writer, req)
	p := writer.Bytes()
	conn.Write(p)
}
