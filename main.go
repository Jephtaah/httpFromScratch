package main

import (
	"fmt"
	"bufio"
	"net"
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
	buf := make([]byte, 1024)

	n, err := reader.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	fmt.Printf("Recieved %d bytes: \n%s\n", n, string(buf[:n]))
}