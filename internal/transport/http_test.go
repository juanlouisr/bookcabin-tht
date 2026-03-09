package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPTransport_Fetch_Success(t *testing.T) {
	// Create a test server
	expectedResponse := map[string]string{"status": "success"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept application/json, got %s", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization Bearer test-api-key, got %s", r.Header.Get("Authorization"))
		}

		// Write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	// Create HTTP transport
	transport := NewHTTPTransport(server.URL, "test-api-key", 10*time.Second)

	req := Request{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	resp, err := transport.Fetch(ctx, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Headers["Content-Type"])
	}

	// Verify response body
	var result map[string]string
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", result["status"])
	}
}

func TestHTTPTransport_Fetch_ServerError(t *testing.T) {
	// Create a test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	transport := NewHTTPTransport(server.URL, "", 10*time.Second)

	req := Request{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	ctx := context.Background()
	resp, err := transport.Fetch(ctx, req)

	// The transport returns the response even on error status codes
	// It's up to the caller to check the status code
	if err != nil {
		t.Fatalf("Expected no error (response returned), got: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestHTTPTransport_Fetch_Timeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create transport with short timeout
	transport := NewHTTPTransport(server.URL, "", 10*time.Millisecond)

	req := Request{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	ctx := context.Background()
	resp, err := transport.Fetch(ctx, req)

	if err == nil {
		t.Error("Expected timeout error")
	}
	if resp != nil {
		t.Error("Expected nil response for timeout")
	}
}

func TestHTTPTransport_GetTimeout(t *testing.T) {
	transport := NewHTTPTransport("http://example.com", "", 5*time.Second)
	if transport.GetTimeout() != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", transport.GetTimeout())
	}
}

func TestHTTPTransport_DefaultTimeout(t *testing.T) {
	transport := NewHTTPTransport("http://example.com", "", 0)
	if transport.GetTimeout() != 10*time.Second {
		t.Errorf("Expected default timeout 10s, got %v", transport.GetTimeout())
	}
}
