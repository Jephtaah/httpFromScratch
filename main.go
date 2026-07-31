package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type Handler func(method, path string, headers map[string]string, body []byte) (int, string)

var routes = map[string]Handler{
	"GET /health": func(method, path string, headers map[string]string, body []byte) (int, string) {
		return 200, "OK"
	},
	"POST /echo": func(method, path string, headers map[string]string, body []byte) (int, string) {
		return 200, string(body)
	},
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting listener:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on port 8080")

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

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request line:", err)
		return
	}

	requestLine = strings.TrimRight(requestLine, "\r\n")
	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		fmt.Println("Malformed request line:", requestLine)
		return
	}

	method := parts[0]
	path := parts[1]
	httpVersion := parts[2]

	fmt.Printf("Method: %s\nPath: %s\nVersion: %s\n", method, path, httpVersion)

	headers := make(map[string]string)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading header line:", err)
			return
		}

		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			break
		}

		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) != 2 {
			fmt.Println("Malformed header line:", line)
			continue
		}

		key := strings.TrimSpace(headerParts[0])
		value := strings.TrimSpace(headerParts[1])
		headers[key] = value
	}

	fmt.Println("Headers:", headers)

	var body []byte

	if contentLengthStr, ok := headers["Content-Length"]; ok {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err == nil {
			body = make([]byte, contentLength)
			io.ReadFull(reader, body)
		}
	}

	routeKey := method + " " + path
	handler, found := routes[routeKey]

	var statusCode int
	var responseBody string

	if found {
		statusCode, responseBody = handler(method, path, headers, body)
	} else {
		statusCode, responseBody = 404, "Not Found"
	}

	statusText := map[int]string{
		200: "OK",
		404: "Not Found",
	}[statusCode]

	response := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\nContent-Length: %d\r\nContent-Type: text/plain\r\n\r\n%s",
		statusCode,
		statusText,
		len(responseBody),
		responseBody,
	)

	conn.Write([]byte(response))
}
