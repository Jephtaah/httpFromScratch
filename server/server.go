package server

import (
	"bufio"
	"fmt"
	"net"
)

func Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("starting listener on %s: %w", addr, err)
	}
	defer listener.Close()

	fmt.Println("Listening on", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		req, err := readRequest(reader)
		if err != nil {
			fmt.Println("Error reading request:", err)
			return
		}

		statusCode, responseBody := dispatch(req)

		if err := writeResponse(conn, req, &Response{StatusCode: statusCode, Body: responseBody}); err != nil {
			fmt.Println("Error writing response:", err)
			return
		}

		if req.wantsClose() {
			return
		}
	}
}
