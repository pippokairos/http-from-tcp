package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"httpfromtcp/internal/headers"
)

const bufferSize = 4096

var SEPARATOR = []byte("\r\n")

var (
	ErrRequestTooLarge        = errors.New("request exceeds maximum size")
	ErrRequestIncomplete      = errors.New("incomplete request")
	ErrInvalidRequestLine     = errors.New("invalid request line")
	ErrUnsupportedHTTPVersion = errors.New("unsupported HTTP version")
	ErrBodyTooLarge           = errors.New("body exceeds Content-Length")
	ErrInvalidParserState     = errors.New("invalid parser state")
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       ParserState
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type ParserState string

const (
	StateInit    ParserState = "initialized"
	StateHeaders ParserState = "headers"
	StateBody    ParserState = "body"
	StateDone    ParserState = "done"
)

func newRequest() *Request {
	return &Request{
		Headers: headers.NewHeaders(),
		State:   StateInit,
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()

	b := make([]byte, bufferSize)
	bufLen := 0
	for !req.done() {
		if bufLen >= len(b) {
			return nil, ErrRequestTooLarge
		}

		read, err := reader.Read(b[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += read

		parsed, err := req.parse(b[:bufLen])
		if err != nil {
			return nil, err
		}

		if parsed > 0 {
			copy(b, b[parsed:bufLen])
			bufLen -= parsed
		}
	}

	return req, nil
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	firstLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(firstLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrInvalidRequestLine
	}

	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, 0, ErrInvalidRequestLine
		}
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" {
		return nil, 0, ErrInvalidRequestLine
	}
	httpVersion := string(httpParts[1])
	if httpVersion != "1.1" {
		return nil, 0, ErrUnsupportedHTTPVersion
	}

	return &RequestLine{
		Method:        string(method),
		RequestTarget: string(parts[1]),
		HttpVersion:   httpVersion,
	}, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

	for {
		switch r.State {
		case StateInit:
			reqLine, n, err := parseRequestLine(data)
			if err != nil {
				return 0, err
			}

			if n == 0 {
				return 0, nil
			}

			r.RequestLine = *reqLine
			read += n
			r.State = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				return 0, err
			}

			if n == 0 {
				return read, nil
			}

			read += n

			if done {
				if r.hasBody() {
					r.State = StateBody
				} else {
					r.State = StateDone
				}
			}

		case StateBody:
			cl := r.Headers.Get(headers.ContentLengthHeader)
			if cl == "" {
				return read, nil
			}

			contentLength, err := strconv.Atoi(cl)
			if err != nil {
				return 0, err
			}

			if contentLength == 0 {
				r.State = StateDone
				return read, nil
			}

			if len(data[read:]) == 0 {
				return read, nil
			}

			remaining := contentLength - len(r.Body)
			n := min(remaining, len(data[read:]))
			r.Body = append(r.Body, data[read:read+n]...)
			read += n

			if len(r.Body) == contentLength {
				r.State = StateDone
			}

		case StateDone:
			return read, nil

		default:
			return 0, ErrInvalidParserState
		}
	}
}

func (r *Request) done() bool {
	return r.State == StateDone
}

func (r *Request) hasBody() bool {
	cl := r.Headers.Get(headers.ContentLengthHeader)
	if cl == "" {
		return false
	}

	contentLength, err := strconv.Atoi(cl)
	if err != nil || contentLength == 0 {
		return false
	}

	return true
}
