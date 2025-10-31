package main

import (
	"fmt"
	"net"

	"httpfromtcp/internal/request"
)

func main() {
	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		req, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println("Error reading request:", err)
			conn.Close()
			continue
		}

		printRequest(req)
		conn.Close()
	}
}

func printRequest(req *request.Request) {
	fmt.Printf("Request line:\n")
	fmt.Printf("- Method: %s\n", req.RequestLine.Method)
	fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

	fmt.Printf("Headers:\n")
	if len(req.Headers) > 0 {
		for name, value := range req.Headers {
			fmt.Printf("- %s: %s\n", name, value)
		}
	}

	if len(req.Body) > 0 {
		fmt.Printf("Body:\n%s\n", string(req.Body))
	}
}
