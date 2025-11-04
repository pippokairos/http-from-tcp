package response

import (
	"fmt"
	"io"
	"strconv"

	"httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var statusLine string
	switch statusCode {
	case StatusOK:
		statusLine = "HTTP/1.1 200 OK"
	case StatusBadRequest:
		statusLine = "HTTP/1.1 400 Bad Request"
	case StatusInternalServerError:
		statusLine = "HTTP/1.1 500 Internal Server Error"
	default:
		statusLine = fmt.Sprintf("HTTP/1.1 %d ", statusCode)
	}

	_, err := fmt.Fprintf(w, "%s\r\n", statusLine)

	return err
}

func GetDefaultHeaders(contentLength int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = strconv.Itoa(contentLength)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for name, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", name, value)
		if err != nil {
			return err
		}
	}

	_, err := w.Write([]byte("\r\n"))

	return err
}
