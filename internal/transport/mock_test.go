package transport

import (
	"context"
	"testing"
	"time"
)

func TestMockTransport_Fetch_Success(t *testing.T) {
	testData := []byte(`{"status": "success"}`)
	mockTransport := NewMockTransportFromBytes(testData, 10, 20, 1.0)

	req := Request{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	ctx := context.Background()
	resp, err := mockTransport.Fetch(ctx, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if string(resp.Body) != string(testData) {
		t.Errorf("Expected body %s, got %s", string(testData), string(resp.Body))
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Headers["Content-Type"])
	}
}

func TestMockTransport_Fetch_SimulatedDelay(t *testing.T) {
	testData := []byte(`{"status": "success"}`)
	// Set delay range of 50-60ms
	mockTransport := NewMockTransportFromBytes(testData, 50, 60, 1.0)

	req := Request{}
	ctx := context.Background()

	start := time.Now()
	_, err := mockTransport.Fetch(ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have taken at least 50ms
	if elapsed < 45*time.Millisecond {
		t.Errorf("Expected at least 45ms delay, got %v", elapsed)
	}
}

func TestMockTransport_Fetch_SimulatedFailure(t *testing.T) {
	testData := []byte(`{"status": "success"}`)
	// 0% success rate - always fail
	mockTransport := NewMockTransportFromBytes(testData, 10, 20, 0.0)

	req := Request{}
	ctx := context.Background()

	_, err := mockTransport.Fetch(ctx, req)

	if err == nil {
		t.Error("Expected error for simulated failure, got nil")
	}
}

func TestMockTransport_Fetch_ContextCancellation(t *testing.T) {
	testData := []byte(`{"status": "success"}`)
	// Long delay
	mockTransport := NewMockTransportFromBytes(testData, 500, 600, 1.0)

	req := Request{}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := mockTransport.Fetch(ctx, req)

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestMockTransport_Fetch_DataLoaderError(t *testing.T) {
	// Create a file loader for non-existent file
	loader := NewFileDataLoader("/nonexistent/file.json")
	mockTransport := NewMockTransport(loader, 10, 20, 1.0)

	req := Request{}
	ctx := context.Background()

	_, err := mockTransport.Fetch(ctx, req)

	if err == nil {
		t.Error("Expected error for data loader failure, got nil")
	}
}

func TestNewMockTransportFromFile(t *testing.T) {
	// This will fail since the file doesn't exist, but we test the function exists
	mockTransport := NewMockTransportFromFile("../../spec/garuda_indonesia_search_response.json", 10, 20, 1.0)

	if mockTransport == nil {
		t.Error("Expected mockTransport to be created")
	}

	// Test with valid file
	req := Request{}
	ctx := context.Background()
	resp, err := mockTransport.Fetch(ctx, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp == nil || len(resp.Body) == 0 {
		t.Error("Expected non-empty response body")
	}
}

func TestMockTransport_GetTimeout(t *testing.T) {
	mockTransport := NewMockTransportFromBytes([]byte(`{}`), 10, 20, 1.0)

	timeout := mockTransport.GetTimeout()
	if timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", timeout)
	}
}

func TestMockTransport_Fetch_MultipleCalls(t *testing.T) {
	testData := []byte(`{"status": "success"}`)
	mockTransport := NewMockTransportFromBytes(testData, 10, 20, 1.0)

	req := Request{}
	ctx := context.Background()

	// Make multiple calls
	for i := 0; i < 5; i++ {
		resp, err := mockTransport.Fetch(ctx, req)
		if err != nil {
			t.Errorf("Call %d: Expected no error, got: %v", i+1, err)
			continue
		}
		if string(resp.Body) != string(testData) {
			t.Errorf("Call %d: Body mismatch", i+1)
		}
	}
}

func TestMockTransport_Fetch_RandomFailures(t *testing.T) {
	testData := []byte(`{"status": "success"}`)
	// 50% success rate
	mockTransport := NewMockTransportFromBytes(testData, 10, 20, 0.5)

	req := Request{}
	ctx := context.Background()

	successCount := 0
	failureCount := 0

	// Run multiple times to test randomness
	for i := 0; i < 20; i++ {
		_, err := mockTransport.Fetch(ctx, req)
		if err != nil {
			failureCount++
		} else {
			successCount++
		}
	}

	// With 50% success rate over 20 attempts, we should see both successes and failures
	// This is probabilistic, so we just check we got some variety
	t.Logf("Successes: %d, Failures: %d", successCount, failureCount)

	if successCount == 0 {
		t.Error("Expected at least some successes")
	}
	if failureCount == 0 {
		t.Error("Expected at least some failures")
	}
}

func TestMockTransport_Variants(t *testing.T) {
	testData := []byte(`{"test": "data"}`)

	// Test NewMockTransportFromBytes
	transport1 := NewMockTransportFromBytes(testData, 10, 20, 1.0)
	if transport1 == nil {
		t.Error("NewMockTransportFromBytes returned nil")
	}

	// Test NewMockTransport with bytes loader
	loader := NewBytesDataLoader(testData)
	transport2 := NewMockTransport(loader, 10, 20, 1.0)
	if transport2 == nil {
		t.Error("NewMockTransport returned nil")
	}
}
