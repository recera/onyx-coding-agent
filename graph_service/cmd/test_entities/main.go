package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/onyx/onyx-tui/graph_service/internal/analyzer"
	"github.com/onyx/onyx-tui/graph_service/internal/db"
)

func main() {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "entity_test_*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Testing entity system in: %s\n", tempDir)

	// Create a comprehensive test Python file
	testRepoDir := filepath.Join(tempDir, "test_repo")
	err = os.MkdirAll(testRepoDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create test repo dir: %v", err)
	}

	// Create test Python file with complex structures
	pythonCode := `"""
A comprehensive test module for the entity system.
This file contains various Python constructs to test entity extraction.
"""

import os
import sys
from collections import defaultdict

# Global variable
DEBUG_MODE = True

def utility_function(x, y):
    """A simple utility function that adds two numbers."""
    return x + y

class DataProcessor:
    """A data processing class with various methods."""
    
    def __init__(self, name):
        """Initialize the data processor."""
        self.name = name
        self.data = []
    
    def add_data(self, item):
        """Add an item to the data list."""
        self.data.append(item)
        return len(self.data)
    
    def process_data(self):
        """Process all data items."""
        processed = []
        for item in self.data:
            # Call the utility function
            result = utility_function(item, 1)
            processed.append(result)
        return processed
    
    def get_stats(self):
        """Get statistics about the data."""
        if not self.data:
            return {"count": 0, "sum": 0}
        
        total = sum(self.data)
        count = len(self.data)
        return {
            "count": count,
            "sum": total,
            "average": total / count
        }

class AdvancedProcessor(DataProcessor):
    """An advanced processor that inherits from DataProcessor."""
    
    def __init__(self, name, mode="standard"):
        super().__init__(name)
        self.mode = mode
    
    def advanced_process(self):
        """Advanced processing with multiple steps."""
        # Call parent method
        basic_result = self.process_data()
        
        # Apply advanced processing
        advanced_result = []
        for value in basic_result:
            if self.mode == "double":
                advanced_result.append(value * 2)
            else:
                advanced_result.append(value)
        
        return advanced_result

def main():
    """Main function to demonstrate the classes."""
    # Create a basic processor
    processor = DataProcessor("basic")
    processor.add_data(10)
    processor.add_data(20)
    
    # Process the data
    results = processor.process_data()
    stats = processor.get_stats()
    
    print(f"Basic results: {results}")
    print(f"Stats: {stats}")
    
    # Create an advanced processor
    advanced = AdvancedProcessor("advanced", "double")
    advanced.add_data(5)
    advanced.add_data(15)
    
    # Process with advanced method
    advanced_results = advanced.advanced_process()
    print(f"Advanced results: {advanced_results}")

if __name__ == "__main__":
    main()
`

	testFilePath := filepath.Join(testRepoDir, "test_module.py")
	err = os.WriteFile(testFilePath, []byte(pythonCode), 0644)
	if err != nil {
		log.Fatalf("Failed to write test file: %v", err)
	}

	// Create database
	dbPath := filepath.Join(tempDir, "test_db")
	kdb, err := db.NewKuzuDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer kdb.Close()

	// Create schema
	err = kdb.CreateSchema()
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// Create graph builder
	builder := analyzer.NewGraphBuilder(kdb)

	// Build the graph
	fmt.Println("Building graph from test repository...")
	stats, err := builder.BuildGraph(testRepoDir)
	if err != nil {
		log.Fatalf("Failed to build graph: %v", err)
	}

	// Print statistics
	fmt.Printf("\n=== Graph Building Statistics ===\n")
	fmt.Printf("Files processed: %d\n", stats.FilesProcessed)
	fmt.Printf("Functions found: %d\n", stats.FunctionsFound)
	fmt.Printf("Classes found: %d\n", stats.ClassesFound)
	fmt.Printf("Methods found: %d\n", stats.MethodsFound)
	fmt.Printf("Relationships found: %d\n", stats.UnresolvedRelationshipsFound)
	fmt.Printf("Errors encountered: %d\n", stats.ErrorsEncountered)

	// Analyze the extracted entities
	fmt.Printf("\n=== Entity Analysis ===\n")

	// Get all entities
	allEntities := builder.GetAllEntities()
	fmt.Printf("Total entities extracted: %d\n", len(allEntities))

	// Print some example entities
	fmt.Printf("\n--- Sample Entities ---\n")
	count := 0
	for _, entity := range allEntities {
		if count >= 5 { // Show first 5 entities
			break
		}
		fmt.Printf("Entity: %s (Type: %s)\n", entity.Name, entity.Type)
		if entity.Signature != "" {
			fmt.Printf("  Signature: %s\n", entity.Signature)
		}
		if entity.DocString != "" {
			fmt.Printf("  DocString: %.100s...\n", entity.DocString)
		}
		fmt.Printf("  File: %s\n", entity.FilePath)
		if entity.Parent != nil {
			fmt.Printf("  Parent: %s\n", entity.Parent.Name)
		}
		fmt.Println()
		count++
	}

	// Analyze relationships
	fmt.Printf("\n--- Sample Relationships ---\n")
	allRels := builder.GetAllRelationships()
	for i, rel := range allRels {
		if i >= 10 { // Show first 10 relationships
			break
		}
		fmt.Printf("Relationship: %s\n", rel.String())
	}

	// Test some queries
	fmt.Printf("\n=== Database Queries ===\n")

	// Query for all functions
	fmt.Println("Querying for all functions...")
	result, err := builder.QueryGraph("MATCH (f:Function) RETURN f.name, f.signature LIMIT 5")
	if err != nil {
		fmt.Printf("Error querying functions: %v\n", err)
	} else {
		fmt.Printf("Functions result:\n%s\n", result)
	}

	// Query for all classes
	fmt.Println("Querying for all classes...")
	result, err = builder.QueryGraph("MATCH (c:Class) RETURN c.name LIMIT 5")
	if err != nil {
		fmt.Printf("Error querying classes: %v\n", err)
	} else {
		fmt.Printf("Classes result:\n%s\n", result)
	}

	fmt.Println("\n🎉 Entity system test completed successfully!")
}
