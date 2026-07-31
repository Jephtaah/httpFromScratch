// Package client is a tiny net/http wrapper used only by the test suite to
// exercise the from-scratch server from the outside.
package client

import (
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 5 * time.Second

// Client wraps an http.Client bound to a base URL.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a Client that talks to the server at baseURL,
// e.g. "http://127.0.0.1:8081".
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}
}

// Get issues a GET to path and returns the status code and response body.
func (c *Client) Get(path string) (int, string, error) {
	resp, err := c.http.Get(c.baseURL + path)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, string(body), nil
}

// Post issues a POST to path with the given body and content type.
func (c *Client) Post(path, contentType, body string) (int, string, error) {
	resp, err := c.http.Post(c.baseURL+path, contentType, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, string(responseBody), nil
}
