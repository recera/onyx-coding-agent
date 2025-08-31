package main

import (
	"fmt"
	"log"
	"os"

	"github.com/onyx/onyx-tui/graph_service/internal/db"
)

func main() {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kuzu_test_*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Testing KuzuDB in: %s\n", tempDir)

	// Initialize KuzuDB
	kdb, err := db.NewKuzuDatabase(tempDir)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer kdb.Close()

	// Create schema
	fmt.Println("Creating schema...")
	err = kdb.CreateSchema()
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// Test adding a file node
	fmt.Println("Adding file node...")
	err = kdb.AddFileNode("test/file.py", "file.py", "python")
	if err != nil {
		log.Fatalf("Failed to add file node: %v", err)
	}

	// Test adding a function node
	fmt.Println("Adding function node...")
	err = kdb.AddFunctionNode("test/file.py:10-30", "test_function", "def test_function():", "def test_function():\n    pass", "test/file.py")
	if err != nil {
		log.Fatalf("Failed to add function node: %v", err)
	}

	// Test adding a class node
	fmt.Println("Adding class node...")
	err = kdb.AddClassNode("test/file.py:40-60", "TestClass", "class TestClass:", "test/file.py")
	if err != nil {
		log.Fatalf("Failed to add class node: %v", err)
	}

	// Test adding relationships
	fmt.Println("Adding Contains relationship...")
	err = kdb.AddContainsRel("test/file.py", "test/file.py:10-30", "Function")
	if err != nil {
		log.Fatalf("Failed to add Contains relationship: %v", err)
	}

	err = kdb.AddContainsRel("test/file.py", "test/file.py:40-60", "Class")
	if err != nil {
		log.Fatalf("Failed to add Contains relationship: %v", err)
	}

	// Test querying
	fmt.Println("Querying database...")
	result, err := kdb.ExecuteQuery("MATCH (f:File) RETURN f.name, f.language LIMIT 5")
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	fmt.Printf("Query result:\n%s\n", result)

	// Get schema
	fmt.Println("Getting schema...")
	schema, err := kdb.GetSchema()
	if err != nil {
		log.Fatalf("Failed to get schema: %v", err)
	}

	fmt.Printf("Schema:\n%s\n", schema)

	fmt.Println("✅ All tests passed!")
}
