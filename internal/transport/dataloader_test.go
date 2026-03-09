package transport

import (
	"os"
	"testing"
)

func TestFileDataLoader_Load_Success(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test data
	testData := []byte(`{"test": "data"}`)
	if _, err := tmpFile.Write(testData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test loading
	loader := NewFileDataLoader(tmpFile.Name())
	data, err := loader.Load()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(data))
	}
}

func TestFileDataLoader_Load_FileNotFound(t *testing.T) {
	loader := NewFileDataLoader("/nonexistent/path/file.json")
	_, err := loader.Load()

	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestFileDataLoader_LoadFromFile_Success(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testData := []byte(`{"flights": []}`)
	if _, err := tmpFile.Write(testData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	loader := NewFileDataLoader(tmpFile.Name())
	data, err := loader.Load()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Data mismatch")
	}
}

func TestBytesDataLoader_Load(t *testing.T) {
	testData := []byte(`{"test": "data"}`)
	loader := NewBytesDataLoader(testData)

	data, err := loader.Load()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(data))
	}
}

func TestBytesDataLoader_Load_Empty(t *testing.T) {
	loader := NewBytesDataLoader([]byte{})

	data, err := loader.Load()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty data, got %d bytes", len(data))
	}
}

func TestBytesDataLoader_Load_LargeData(t *testing.T) {
	// Create a large byte slice (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	loader := NewBytesDataLoader(largeData)
	data, err := loader.Load()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(data) != len(largeData) {
		t.Errorf("Expected %d bytes, got %d", len(largeData), len(data))
	}

	// Verify data integrity
	for i, b := range data {
		if b != byte(i%256) {
			t.Errorf("Data mismatch at index %d", i)
			break
		}
	}
}

func TestFileDataLoader_Load_RealSpecFile(t *testing.T) {
	// Test loading actual spec files
	specFiles := []string{
		"../../spec/garuda_indonesia_search_response.json",
		"../../spec/airasia_search_response.json",
		"../../spec/lion_air_search_response.json",
		"../../spec/batik_air_search_response.json",
	}

	for _, file := range specFiles {
		t.Run(file, func(t *testing.T) {
			loader := NewFileDataLoader(file)
			data, err := loader.Load()

			if err != nil {
				t.Errorf("Failed to load %s: %v", file, err)
				return
			}

			if len(data) == 0 {
				t.Errorf("Expected non-empty data for %s", file)
			}

			// Verify it's valid JSON by checking first/last characters
			if data[0] != '{' && data[0] != '[' {
				t.Errorf("File %s doesn't appear to be valid JSON", file)
			}
		})
	}
}

func TestFileDataLoader_MultipleLoads(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testData := []byte(`{"count": 1}`)
	if _, err := tmpFile.Write(testData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	loader := NewFileDataLoader(tmpFile.Name())

	// Load multiple times
	for i := 0; i < 3; i++ {
		data, err := loader.Load()
		if err != nil {
			t.Errorf("Load %d: Expected no error, got: %v", i+1, err)
			continue
		}
		if string(data) != string(testData) {
			t.Errorf("Load %d: Data mismatch", i+1)
		}
	}
}

func TestBytesDataLoader_MultipleLoads(t *testing.T) {
	testData := []byte(`{"test": "data"}`)
	loader := NewBytesDataLoader(testData)

	// Load multiple times - should return same data
	for i := 0; i < 3; i++ {
		data, err := loader.Load()
		if err != nil {
			t.Errorf("Load %d: Expected no error, got: %v", i+1, err)
			continue
		}
		if string(data) != string(testData) {
			t.Errorf("Load %d: Data mismatch", i+1)
		}
	}
}
