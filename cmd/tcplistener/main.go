package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("Error", "Error", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error", "Error", err)
		}

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("Error", "Error", err)
		}
		fmt.Printf("Request Line:\n- Method: %s\n- Target: %s\n- Version:%s\n",
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Printf("Body:%s\n", req.Body)

		// TODO: We will respond soon.
		conn.Close()
	}

}
