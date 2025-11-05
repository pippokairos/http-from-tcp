package server

import (
	"fmt"
	"log"
	"net"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

type Server struct {
	handler Handler
	closed  bool
}

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (handlerError HandlerError) write(w *response.Writer) {
	w.WriteHeaders(response.GetDefaultHeaders(len(handlerError.Message)))
	w.Writer.Write([]byte(handlerError.Message))
}

func Serve(port int, handler Handler) (*Server, error) {
	s := &Server{
		handler: handler,
		closed:  false,
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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
		conn, err := l.Accept()
		if err != nil {
			if s.closed {
				return
			}
			log.Println("Error accepting connection:", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		handlerError := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		handlerError.write(responseWriter)
		return
	}

	s.handler(responseWriter, req)
}
