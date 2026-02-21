package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func respond400() []byte {
	return []byte(`<html>
			  <head>
				<title>400 Bad Request</title>
			  </head>
			  <body>
				<h1>Bad Request</h1>
				<p>Your request honestly kinda sucked.</p>
			  </body>
			</html>`)
}

func respond500() []byte {
	return []byte(`<html>
			  <head>
				<title>500 Internal Server Error</title>
			  </head>
			  <body>
				<h1>Internal Server Error</h1>
				<p>Okay, you know what? This one is on me.</p>
			  </body>
			</html>`)
}

func respond200() []byte {
	return []byte(`<html>
			  <head>
				<title>200 OK</title>
			  </head>
			  <body>
				<h1>Success!</h1>
				<p>Your request was an absolute banger.</p>
			  </body>
			</html>`)
}

func main() {
	const port = 42069

	handler := func(w *response.Writer, req *request.Request) {
		if req.RequestLine.RequestTarget == "/yourproblem" {
			w.WriteStatusLine(response.StatusBadRequest)
			headers := headers.NewHeaders()
			headers.Set("Content-Type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody(respond400())

		} else if req.RequestLine.RequestTarget == "/myproblem" {
			w.WriteStatusLine(response.StatusInternalServerError)
			headers := headers.NewHeaders()
			headers.Set("Content-Type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody(respond500())
		} else if req.RequestLine.RequestTarget == "/video" {
			videoData, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				h := headers.NewHeaders()
				h.Set("Content-Type", "text/html")
				w.WriteHeaders(h)
				w.WriteBody(respond500())
			} else {
				w.WriteStatusLine(response.StatusOK)
				h := headers.NewHeaders()
				h.Set("Content-Type", "video/mp4")
				w.WriteHeaders(h)
				w.WriteBody(videoData)
			}

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/html") {
			res, err := http.Get("https://httpbin.org/html")
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				h := headers.NewHeaders()
				h.Set("Content-Type", "text/html")
				w.WriteHeaders(h)
				w.WriteBody(respond500())
			} else {
				defer res.Body.Close()
				w.WriteStatusLine(response.StatusOK)
				h := headers.NewHeaders()
				w.DeleteHeader("content-length")
				h.Set("transfer-encoding", "chunked")
				h.Set("trailer", "X-Content-SHA256, X-Content-Length")
				h.Set("content-type", res.Header.Get("Content-Type"))
				w.WriteHeaders(h)

				var fullBody []byte
				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if n > 0 {
						fullBody = append(fullBody, data[:n]...)
						w.WriteChunkedBody(data[:n])
					}
					if err != nil {
						break
					}
				}
				w.WriteChunkedBodyDone()

				hash := sha256.Sum256(fullBody)
				trailers := headers.NewHeaders()
				trailers.Set("X-Content-SHA256", hex.EncodeToString(hash[:]))
				trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteTrailers(trailers)
			}

		} else {
			w.WriteStatusLine(response.StatusOK)
			headers := headers.NewHeaders()
			headers.Set("Content-Type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody(respond200())

		}
	}

	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
