package main

import (
	"fmt"
	"log"
	"os"

	"github.com/onyx/onyx-tui/graph_service"
)

func main() {
	fmt.Println("=== Go Code Graph Server ===")

	// Check for required arguments
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <repository-path-or-url>\n", os.Args[0])
		fmt.Println("Example:")
		fmt.Printf("  %s /path/to/local/repo\n", os.Args[0])
		fmt.Printf("  %s https://github.com/user/repo.git\n", os.Args[0])
		os.Exit(1)
	}

	repoPathOrURL := os.Args[1]
	fmt.Printf("Analyzing: %s\n", repoPathOrURL)

	// Determine if it's a local path or URL
	var options graph.BuildGraphOptions
	if isLocalPath(repoPathOrURL) {
		options = graph.BuildGraphOptions{
			RepoPath:    repoPathOrURL,
			LoadEnvFile: true, // Load .env for LLM features
		}
	} else {
		options = graph.BuildGraphOptions{
			RepoURL:     repoPathOrURL,
			LoadEnvFile: true, // Load .env for LLM features
		}
	}

	// Build the graph
	fmt.Println("\n=== Building Code Graph ===")
	result, err := graph.BuildGraph(options)
	if err != nil {
		log.Fatalf("Failed to build graph: %v", err)
	}
	defer result.Close()

	// Print results
	fmt.Printf("✅ Analysis Complete!\n")
	fmt.Printf("Files processed: %d\n", result.Stats.FilesCount)
	fmt.Printf("Functions found: %d\n", result.Stats.FunctionsCount)
	fmt.Printf("Classes found: %d\n", result.Stats.ClassesCount)
	fmt.Printf("Methods found: %d\n", result.Stats.MethodsCount)
	fmt.Printf("Relationships found: %d\n", result.Stats.CallsCount)
	fmt.Printf("Errors encountered: %d\n", result.Stats.ErrorsCount)
	fmt.Printf("Database path: %s\n", result.DBPath)

	// Example queries
	fmt.Println("\n=== Sample Queries ===")

	// Query 1: List all classes
	fmt.Println("1. All classes in the codebase:")
	classQuery := "MATCH (c:Class) RETURN c.name ORDER BY c.name LIMIT 10"
	classResult, err := result.QueryGraph(classQuery)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Classes found:\n%s\n", classResult)
	}

	// Query 2: Find inheritance relationships
	fmt.Println("2. Inheritance relationships:")
	inheritQuery := "MATCH (child)-[r:INHERITS]->(parent) RETURN child.name, parent.name LIMIT 5"
	inheritResult, err := result.QueryGraph(inheritQuery)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Inheritance found:\n%s\n", inheritResult)
	}

	// Show analysis details
	analysis := result.GetAnalysisResult()
	fmt.Printf("\n=== Detailed Analysis ===\n")
	fmt.Printf("Total entities: %d\n", len(analysis.Entities))
	fmt.Printf("Total relationships: %d\n", len(analysis.Relationships))

	// Show inheritance relationships from analysis
	inheritanceCount := 0
	for _, rel := range analysis.Relationships {
		if rel.Type == "INHERITS" {
			inheritanceCount++
		}
	}
	fmt.Printf("Inheritance relationships: %d\n", inheritanceCount)

	fmt.Println("\n🎉 Server completed successfully!")
	fmt.Printf("Database available at: %s\n", result.DBPath)
	fmt.Println("You can now query the graph using the database path.")
}

// isLocalPath checks if the given string is a local file path
func isLocalPath(path string) bool {
	// Simple heuristic: if it doesn't start with http:// or https://, treat as local path
	return !(len(path) > 7 && (path[:7] == "http://" || path[:8] == "https://"))
}
