package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

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

		handleConnection(conn)
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

	hearders := make(map[string]string)

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
		hearders[key] = value
	}

	fmt.Println("Headers:", hearders)
}
