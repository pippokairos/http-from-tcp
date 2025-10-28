package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

var CRLF = []byte("\r\n")

var ErrInvalidHeader = errors.New("invalid header format")

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 || len(parts[0]) == 0 || parts[0][len(parts[0])-1] == ' ' {
		return "", "", ErrInvalidHeader
	}

	if !isToken(parts[0]) {
		return "", "", ErrInvalidHeader
	}

	name := bytes.ToLower((parts[0]))
	value := bytes.Trim(parts[1], " ")

	return string(name), string(value), nil
}

func isToken(b []byte) bool {
	allowedChars := "!#$%&'*+-.^_`|~"

	for _, c := range b {
		if (c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') {
			continue
		}

		if strings.ContainsRune(allowedChars, rune(c)) {
			continue
		}

		return false
	}

	return true
}

func (h Headers) Parse(b []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(b[read:], CRLF)
		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			break
		}

		fieldLine := bytes.Trim(b[read:read+idx], " ")
		name, value, err := parseHeader(fieldLine)
		if err != nil {
			return 0, false, err
		}

		if v, ok := h[name]; ok {
			h[name] = fmt.Sprintf("%s, %s", v, value)
		} else {
			h[name] = value
		}

		read += idx + len(CRLF)
	}

	return read, done, nil
}
