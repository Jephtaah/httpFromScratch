package server

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    []byte
}

func readRequest(reader *bufio.Reader) (*Request, error) {
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("reading request line: %w", err)
	}

	parts := strings.Split(strings.TrimRight(requestLine, "\r\n"), " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed request line %q", requestLine)
	}

	req := &Request{
		Method:  parts[0],
		Path:    parts[1],
		Version: parts[2],
		Headers: make(map[string]string),
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("reading header line: %w", err)
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) != 2 {
			return nil, fmt.Errorf("malformed header line %q", line)
		}

		key := textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(headerParts[0]))
		value := strings.TrimSpace(headerParts[1])
		req.Headers[key] = value
	}

	if contentLengthStr, ok := req.Headers["Content-Length"]; ok {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length %q: %w", contentLengthStr, err)
		}
		if contentLength < 0 {
			return nil, fmt.Errorf("negative Content-Length %d", contentLength)
		}

		req.Body = make([]byte, contentLength)
		if _, err := io.ReadFull(reader, req.Body); err != nil {
			return nil, fmt.Errorf("reading body: %w", err)
		}
	}

	return req, nil
}

func (r *Request) wantsClose() bool {
	return strings.EqualFold(r.Headers["Connection"], "close")
}
