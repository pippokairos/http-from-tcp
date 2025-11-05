package response

import (
	"errors"
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

type WriterState string

const (
	StateStatusLine WriterState = "statusLine"
	StateHeaders    WriterState = "headers"
	StateBody       WriterState = "body"
	StateDone       WriterState = "done"
)

var ErrInvalidState = errors.New("invalid writer state")

type Writer struct {
	Writer io.Writer
	State  WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer: w,
		State:  StateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != StateStatusLine {
		return ErrInvalidState
	}

	var statusLine string
	switch statusCode {
	case StatusOK:
		statusLine = "HTTP/1.1 200 OK\n"
	case StatusBadRequest:
		statusLine = "HTTP/1.1 400 Bad Request\n"
	case StatusInternalServerError:
		statusLine = "HTTP/1.1 500 Internal Server Error\n"
	default:
		statusLine = fmt.Sprintf("HTTP/1.1 %d \n", statusCode)
	}

	_, err := w.Writer.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	w.State = StateHeaders

	return nil
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.State != StateHeaders {
		return ErrInvalidState
	}

	for name, value := range h {
		_, err := fmt.Fprintf(w.Writer, "%s: %s\r\n", name, value)
		if err != nil {
			return err
		}
	}

	_, err := w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.State = StateBody

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.State != StateBody {
		return 0, ErrInvalidState
	}

	n, err := w.Writer.Write(p)
	if err != nil {
		return 0, err
	}

	w.State = StateDone

	return n, nil
}

func GetDefaultHeaders(contentLength int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = strconv.Itoa(contentLength)
	h["Connection"] = "close"
	h["Content-Type"] = "text/html"

	return h
}
