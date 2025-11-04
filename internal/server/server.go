package server

import (
	"fmt"
	"net"

	"httpfromtcp/internal/response"
)

type Server struct {
	port   int
	closed bool
}

func Serve(port int) (*Server, error) {
	s := &Server{
		port:   port,
		closed: false,
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return nil, err
	}

	go s.listen(l)

	return s, nil
}

func (s *Server) Close() error {
	s.closed = true

	return nil
}

func (s *Server) listen(l net.Listener) {
	for {
		if s.closed {
			return
		}

		conn, err := l.Accept()
		if err != nil {
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		return
	}

	body := "Hello World!"

	err = response.WriteHeaders(conn, response.GetDefaultHeaders(len(body)))
	if err != nil {
		return
	}

	conn.Write([]byte(body))
}
