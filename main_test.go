package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	go startServer(":8081")
	time.Sleep(200 * time.Millisecond) 

	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	if string(body) != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", string(body))
	}
}

func TestEchoEndpoint(t *testing.T) {
	resp, err := http.Post("http://localhost:8081/echo", "text/plain", strings.NewReader("test message"))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	if string(body) != "test message" {
		t.Errorf("Expected echoed body 'test message', got '%s'", string(body))
	}
}

func TestNotFoundRoute(t *testing.T) {
	resp, err := http.Get("http://localhost:8081/doesnotexist")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}
