package main

import (
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	const port = 42069

	handler := func(w *response.Writer, req *request.Request) {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			w.WriteStatusLine(response.StatusBadRequest)
			headers := headers.NewHeaders()
			headers.Set("Content-Type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody([]byte(`<html>
								  <head>
									<title>400 Bad Request</title>
								  </head>
								  <body>
									<h1>Bad Request</h1>
									<p>Your request honestly kinda sucked.</p>
								  </body>
								</html>`))
		case "/myproblem":
			w.WriteStatusLine(response.StatusInternalServerError)
			headers := headers.NewHeaders()
			headers.Set("Content-Type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody([]byte(`<html>
								  <head>
									<title>500 Internal Server Error</title>
								  </head>
								  <body>
									<h1>Internal Server Error</h1>
									<p>Okay, you know what? This one is on me.</p>
								  </body>
								</html>`))
		default:
			w.WriteStatusLine(response.StatusOK)
			headers := headers.NewHeaders()
			headers.Set("Content-Type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody([]byte(`<html>
								  <head>
									<title>200 OK</title>
								  </head>
								  <body>
									<h1>Success!</h1>
									<p>Your request was an absolute banger.</p>
								  </body>
								</html>`))
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
