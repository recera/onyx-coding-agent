package main

import (
	"fmt"
	"io"
	"os"
)

// Writer interface defines writing capability
type Writer interface {
	Write(data []byte) (int, error)
}

// Reader interface defines reading capability
type Reader interface {
	Read(data []byte) (int, error)
}

// ReadWriter combines Reader and Writer interfaces
type ReadWriter interface {
	Reader
	Writer
}

// Logger interface for logging functionality
type Logger interface {
	Log(message string)
	Level() string
}

// Base struct that will be embedded
type Base struct {
	ID   int
	Name string
}

// GetID returns the ID of the base
func (b *Base) GetID() int {
	return b.ID
}

// SetName sets the name of the base
func (b *Base) SetName(name string) {
	b.Name = name
}

// FileHandler embeds Base and implements ReadWriter
type FileHandler struct {
	Base   // Embedded struct
	file   *os.File
	logger Logger
}

// NewFileHandler creates a new file handler
func NewFileHandler(filename string, logger Logger) (*FileHandler, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &FileHandler{
		Base:   Base{ID: 1, Name: filename},
		file:   file,
		logger: logger,
	}, nil
}

// Read implements the Reader interface
func (fh *FileHandler) Read(data []byte) (int, error) {
	fh.logger.Log("Reading from file: " + fh.Name)
	return fh.file.Read(data)
}

// Write implements the Writer interface
func (fh *FileHandler) Write(data []byte) (int, error) {
	fh.logger.Log("Writing to file: " + fh.Name)
	return fh.file.Write(data)
}

// Close closes the file
func (fh *FileHandler) Close() error {
	fh.logger.Log("Closing file: " + fh.Name)
	return fh.file.Close()
}

// SimpleLogger implements the Logger interface
type SimpleLogger struct {
	level string
}

// NewSimpleLogger creates a new simple logger
func NewSimpleLogger(level string) *SimpleLogger {
	return &SimpleLogger{level: level}
}

// Log implements the Logger interface
func (sl *SimpleLogger) Log(message string) {
	fmt.Printf("[%s] %s\n", sl.level, message)
}

// Level implements the Logger interface
func (sl *SimpleLogger) Level() string {
	return sl.level
}

// ProcessFile demonstrates interface usage and method calls
func ProcessFile(rw ReadWriter, logger Logger) error {
	logger.Log("Starting file processing")

	data := make([]byte, 1024)
	n, err := rw.Read(data)
	if err != nil && err != io.EOF {
		return err
	}

	if n > 0 {
		processed := processData(data[:n])
		_, err = rw.Write(processed)
		if err != nil {
			return err
		}
	}

	logger.Log("File processing completed")
	return nil
}

// processData processes the input data
func processData(data []byte) []byte {
	// Simple processing: convert to uppercase
	result := make([]byte, len(data))
	for i, b := range data {
		if b >= 'a' && b <= 'z' {
			result[i] = b - 32
		} else {
			result[i] = b
		}
	}
	return result
}

// Factory function demonstrating type instantiation
func CreateHandlerWithLogger(filename string) (*FileHandler, error) {
	logger := NewSimpleLogger("INFO")
	return NewFileHandler(filename, logger)
}

func main() {
	fmt.Println("Advanced Go Example - Interface Implementation & Struct Embedding")

	// Create a logger
	logger := NewSimpleLogger("DEBUG")

	// Create a file handler
	handler, err := NewFileHandler("test.txt", logger)
	if err != nil {
		logger.Log("Failed to create file handler: " + err.Error())
		return
	}
	defer handler.Close()

	// Use the handler as a ReadWriter interface
	err = ProcessFile(handler, logger)
	if err != nil {
		logger.Log("Error processing file: " + err.Error())
		return
	}

	// Demonstrate embedded struct access
	fmt.Printf("Handler ID: %d, Name: %s\n", handler.GetID(), handler.Name)

	// Type assertion example
	if rw, ok := interface{}(handler).(ReadWriter); ok {
		logger.Log("Handler implements ReadWriter interface")
		_ = rw // Use the type-asserted value
	}

	logger.Log("Program completed successfully")
}
