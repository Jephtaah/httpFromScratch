package main_test

import (
	"net"
	"sync"
	"testing"
	"time"

	"httpFromScratch/client"
	"httpFromScratch/server"
)

const testAddr = "127.0.0.1:8081"

var (
	startOnce  sync.Once
	testClient *client.Client
)

func startTestServer(t *testing.T) {
	t.Helper()

	startOnce.Do(func() {
		go func() {
			if err := server.Start(testAddr); err != nil {
				t.Errorf("server.Start returned error: %v", err)
			}
		}()
		waitForServer(t)
	})

	testClient = client.New("http://" + testAddr)
}

func waitForServer(t *testing.T) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err := net.Dial("tcp", testAddr)
		if err == nil {
			conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("server never became reachable on %s: %v", testAddr, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestHealthEndpoint(t *testing.T) {
	startTestServer(t)

	status, body, err := testClient.Get("/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	if status != 200 {
		t.Errorf("expected status 200, got %d", status)
	}
	if body != "OK" {
		t.Errorf("expected body %q, got %q", "OK", body)
	}
}

func TestEchoEndpoint(t *testing.T) {
	startTestServer(t)

	status, body, err := testClient.Post("/echo", "text/plain", "test message")
	if err != nil {
		t.Fatalf("POST /echo failed: %v", err)
	}
	if status != 200 {
		t.Errorf("expected status 200, got %d", status)
	}
	if body != "test message" {
		t.Errorf("expected body %q, got %q", "test message", body)
	}
}

func TestNotFoundRoute(t *testing.T) {
	startTestServer(t)

	status, _, err := testClient.Get("/doesnotexist")
	if err != nil {
		t.Fatalf("GET /doesnotexist failed: %v", err)
	}
	if status != 404 {
		t.Errorf("expected status 404, got %d", status)
	}
}
