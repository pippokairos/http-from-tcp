package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const (
	port          = 42069
	httpbinPrefix = "/httpbin/"
)

func toString(bytes []byte) string {
	var s string
	for _, b := range bytes {
		s += fmt.Sprintf("%02x", b)
	}

	return s
}

func respond400() string {
	return `
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
}

func respond500() string {
	return `
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
}

func respond200() string {
	return `
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		var body string
		var status response.StatusCode

		endpoint, found := strings.CutPrefix(req.RequestLine.RequestTarget, httpbinPrefix)
		if found {
			r, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", endpoint))
			if err != nil || r.StatusCode != http.StatusOK {
				log.Println("Error fetching from httpbin")
				status = response.StatusInternalServerError
				body = respond500()
			} else {
				w.WriteStatusLine(response.StatusOK)
				w.WriteHeaders(response.GetChunkedHeaders())

				var fullBody []byte
				for {
					b := make([]byte, 1024)
					n, err := r.Body.Read(b)
					if err != nil {
						break
					}

					fullBody = append(fullBody, b[:n]...)
					w.WriteChunkedBody(b)
				}
				w.WriteChunkedBodyDone()

				trailers := headers.NewHeaders()
				sha := sha256.Sum256(fullBody)
				trailers["X-Content-SHA256"] = fmt.Sprintf("%x", toString(sha[:]))
				trailers["X-Content-Length"] = fmt.Sprintf("%d", len(fullBody))
				err := w.WriteTrailers(trailers)
				if err != nil {
					log.Println("Error writing trailers:", err)
				}

				return
			}
		} else if req.RequestLine.RequestTarget == "/video" {
			data, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				log.Println("Error reading video file:", err)
			}

			w.WriteStatusLine(response.StatusOK)
			h := response.GetDefaultHeaders(len(data))
			h["Content-Type"] = "video/mp4"
			w.WriteHeaders(h)
			w.WriteBody(data)
			return
		} else if req.RequestLine.RequestTarget == "/yourproblem" {
			status = response.StatusBadRequest
			body = respond400()
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			status = response.StatusInternalServerError
			body = respond500()
		} else {
			status = response.StatusOK
			body = respond200()
		}

		w.WriteStatusLine(status)
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody([]byte(body))
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()

	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Server gracefully stopped")
}
