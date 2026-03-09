package transport

import (
	"fmt"
	"os"
)

// DataLoader defines the interface for loading mock data
type DataLoader interface {
	Load() ([]byte, error)
}

// FileDataLoader loads data from a file
type FileDataLoader struct {
	FilePath string
}

// NewFileDataLoader creates a new file-based data loader
func NewFileDataLoader(filePath string) *FileDataLoader {
	return &FileDataLoader{FilePath: filePath}
}

// Load reads data from the file
func (l *FileDataLoader) Load() ([]byte, error) {
	data, err := os.ReadFile(l.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mock data from file: %w", err)
	}
	return data, nil
}

// BytesDataLoader loads data from an in-memory byte slice
type BytesDataLoader struct {
	Data []byte
}

// NewBytesDataLoader creates a new byte slice-based data loader
func NewBytesDataLoader(data []byte) *BytesDataLoader {
	return &BytesDataLoader{Data: data}
}

// Load returns the in-memory data
func (l *BytesDataLoader) Load() ([]byte, error) {
	return l.Data, nil
}

// Ensure interfaces are implemented
var _ DataLoader = (*FileDataLoader)(nil)
var _ DataLoader = (*BytesDataLoader)(nil)
