package server

import (
	"fmt"
	"net"
)

var statusTexts = map[int]string{
	200: "OK",
	404: "Not Found",
	500: "Internal Server Error",
}

type Response struct {
	StatusCode int
	Body       string
}

func writeResponse(conn net.Conn, req *Request, resp *Response) error {
	statusText, ok := statusTexts[resp.StatusCode]
	if !ok {
		statusText = "Unknown"
	}

	connection := "keep-alive"
	if req.wantsClose() {
		connection = "close"
	}

	response := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\nContent-Length: %d\r\nContent-Type: text/plain\r\nConnection: %s\r\n\r\n%s",
		resp.StatusCode,
		statusText,
		len(resp.Body),
		connection,
		resp.Body,
	)

	_, err := conn.Write([]byte(response))
	return err
}
