package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: Valid single header with extra whitespaces
	headers = NewHeaders()
	data = []byte("   Host:    localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 35, n)
	assert.True(t, done)

	// Test: Valid 2 headers with existing headers
	headers = Headers{
		"user-agent": "curl/7.81.0",
	}
	data = []byte("Host: localhost:42069\r\nAccept: */*\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "curl/7.81.0", headers["user-agent"])
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "*/*", headers["accept"])
	assert.Equal(t, 38, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character in the header
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Two values with the same name in headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nAccept-Language: en\r\nAccept-Language: nl\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "en, nl", headers["accept-language"])
	assert.Equal(t, 67, n)
	assert.True(t, done)

	// Test: Multiple values with the same name in headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nAccept-Language: en\r\nAccept-Language: nl\r\nAccept-Language: it\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "en, nl, it", headers["accept-language"])
	assert.Equal(t, 88, n)
	assert.True(t, done)
}
