package main

// import (
// 	"fmt"
// 	"log"
// 	"os"

// 	codeanalyzer "graph"
// )

// func main() {
// 	// Check if repository path is provided
// 	if len(os.Args) < 2 {
// 		fmt.Println("Usage: go run main.go <repository_path>")
// 		os.Exit(1)
// 	}
// 	repoPath := os.Args[1]

// 	fmt.Printf("Analyzing repository: %s\n", repoPath)

// 	// Build graph from the repository
// 	result, err := codeanalyzer.BuildGraph(codeanalyzer.BuildGraphOptions{
// 		RepoPath:    repoPath,
// 		LoadEnvFile: true, // Will attempt to load .env file
// 	})
// 	if err != nil {
// 		log.Fatalf("Failed to build graph: %v", err)
// 	}

// 	// Print statistics
// 	fmt.Printf("\nAnalysis complete!\n")
// 	fmt.Printf("Database path: %s\n", result.DBPath)
// 	fmt.Printf("Statistics:\n")
// 	fmt.Printf("- Functions: %d\n", result.Stats.FunctionsCount)
// 	fmt.Printf("- Classes: %d\n", result.Stats.ClassesCount)
// 	fmt.Printf("- Calls: %d\n", result.Stats.CallsCount)

// 	// If OPENAI_API_KEY is set, try to query the graph
// 	if os.Getenv("OPENAI_API_KEY") != "" {
// 		fmt.Println("\nQuerying the graph...")
// 		answer, err := codeanalyzer.QueryGraph(result.Database, "Which function has the most calls?")
// 		if err != nil {
// 			log.Printf("Failed to query graph: %v", err)
// 		} else {
// 			fmt.Printf("Query result:\n%s\n", answer)
// 		}
// 	} else {
// 		fmt.Println("\nOPENAI_API_KEY not set. Skipping graph query.")
// 	}

// 	fmt.Println("\nDone! The database is stored at:", result.DBPath)
// 	fmt.Println("You can keep this database for future queries or delete it when no longer needed.")
// }
