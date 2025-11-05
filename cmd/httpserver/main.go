package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

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

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			status = response.StatusBadRequest
			body = respond400()
		case "/myproblem":
			status = response.StatusInternalServerError
			body = respond500()
		default:
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
