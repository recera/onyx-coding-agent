package main

import (
	"fmt"
	"log"
	"path/filepath"

	codeanalyzer "github.com/onyx/onyx-tui/graph_service"
)

func main() {
	// Path to the code-graph-backend-main repository
	repoPath := "../../../code-graph-backend-main"

	fmt.Printf("Analyzing code-graph-backend-main repository...\n")
	fmt.Printf("Repository path: %s\n", repoPath)

	// Build the graph
	result, err := codeanalyzer.BuildGraph(codeanalyzer.BuildGraphOptions{
		RepoPath:    repoPath,
		CleanupDB:   false, // Don't cleanup so we can inspect the DB
		LoadEnvFile: true,
	})
	if err != nil {
		log.Fatalf("Failed to build graph: %v", err)
	}
	defer result.Close()

	// Display basic statistics
	fmt.Printf("\n=== Analysis Statistics ===\n")
	fmt.Printf("Files processed: %d\n", result.Stats.FilesCount)
	fmt.Printf("Functions found: %d\n", result.Stats.FunctionsCount)
	fmt.Printf("Classes found: %d\n", result.Stats.ClassesCount)
	fmt.Printf("Methods found: %d\n", result.Stats.MethodsCount)
	fmt.Printf("Total relationships: %d\n", result.Stats.CallsCount)
	fmt.Printf("Errors encountered: %d\n", result.Stats.ErrorsCount)

	// Get detailed analysis results
	analysis := result.GetAnalysisResult()

	fmt.Printf("\n=== Entity Analysis ===\n")
	fmt.Printf("Total entities extracted: %d\n", len(analysis.Entities))

	// Demonstrate KuzuDB queries with corrected property names
	fmt.Printf("\n=== KuzuDB Graph Queries ===\n")

	// Query 1: Schema inspection
	fmt.Println("1. Database schema:")
	schemaResult, err := result.QueryGraph("CALL show_tables() RETURN *;")
	if err != nil {
		fmt.Printf("   Schema query error: %v\n", err)
	} else {
		fmt.Printf("%s\n", schemaResult)
	}

	// Query 2: All classes with correct property names
	fmt.Println("\n2. All classes in the codebase:")
	classResult, err := result.QueryGraph("MATCH (c:Class) RETURN c.name ORDER BY c.name LIMIT 10")
	if err != nil {
		fmt.Printf("   Query error: %v\n", err)
	} else {
		fmt.Printf("%s\n", classResult)
	}

	// Query 3: Functions with their signatures
	fmt.Println("\n3. Functions with signatures:")
	funcResult, err := result.QueryGraph("MATCH (f:Function) RETURN f.name, f.signature ORDER BY f.name LIMIT 8")
	if err != nil {
		fmt.Printf("   Query error: %v\n", err)
	} else {
		fmt.Printf("%s\n", funcResult)
	}

	// Query 4: Class hierarchy (inheritance)
	fmt.Println("\n4. Class inheritance relationships:")
	inheritResult, err := result.QueryGraph("MATCH (child:Class)-[:INHERITS]->(parent:Class) RETURN child.name + ' inherits from ' + parent.name as inheritance")
	if err != nil {
		fmt.Printf("   Query error: %v\n", err)
	} else {
		fmt.Printf("%s\n", inheritResult)
	}

	// Query 5: Function calls
	fmt.Println("\n5. Function call relationships:")
	callResult, err := result.QueryGraph("MATCH (caller)-[:CALLS]->(callee) RETURN caller.name + ' calls ' + callee.name as call_relationship LIMIT 10")
	if err != nil {
		fmt.Printf("   Query error: %v\n", err)
	} else {
		fmt.Printf("%s\n", callResult)
	}

	// Query 6: Files and their imports
	fmt.Println("\n6. Import relationships:")
	importResult, err := result.QueryGraph("MATCH (file:File)-[:IMPORTS]->(imported) RETURN file.name + ' imports ' + imported.name as import_relationship LIMIT 10")
	if err != nil {
		fmt.Printf("   Query error: %v\n", err)
	} else {
		fmt.Printf("%s\n", importResult)
	}

	// Query 7: Entity types and counts
	fmt.Println("\n7. Entity types and their counts:")
	typeResult, err := result.QueryGraph("MATCH (n) RETURN label(n) as entity_type, count(*) as count ORDER BY count DESC")
	if err != nil {
		fmt.Printf("   Query error: %v\n", err)
	} else {
		fmt.Printf("%s\n", typeResult)
	}

	// Query 8: Classes and their methods
	fmt.Println("\n8. Classes and their methods:")
	methodResult, err := result.QueryGraph("MATCH (c:Class)-[:CONTAINS]->(m:Function) RETURN c.name as class, m.name as method LIMIT 15")
	if err != nil {
		fmt.Printf("   Query error: %v\n", err)
	} else {
		fmt.Printf("%s\n", methodResult)
	}

	// Show some specific interesting entities
	fmt.Printf("\n--- Key Classes and Their Structure ---\n")
	classCount := 0
	for _, entity := range analysis.Entities {
		if entity.Type == "Class" && classCount < 5 {
			fmt.Printf("Class: %s (in %s)\n", entity.Name, filepath.Base(entity.FilePath))
			if entity.DocString != "" {
				docStr := entity.DocString
				if len(docStr) > 150 {
					docStr = docStr[:150] + "..."
				}
				fmt.Printf("  Description: %s\n", docStr)
			}
			// Show methods for this class
			methodCount := 0
			for _, child := range entity.Children {
				if (child.Type == "Method" || child.Type == "Function") && methodCount < 5 {
					fmt.Printf("  - %s%s\n", child.Name, child.Signature)
					methodCount++
				}
			}
			fmt.Println()
			classCount++
		}
	}

	fmt.Println("\n🎉 Analysis of code-graph-backend-main completed!")
	fmt.Printf("Graph database created with %d entities and %d relationships\n",
		len(analysis.Entities), len(analysis.Relationships))
	fmt.Printf("Database preserved for inspection at: %s\n", result.DBPath)
}
