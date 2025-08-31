package main

// import (
// 	"fmt"
// 	"log"

// 	codeanalyzer "github.com/onyx/onyx-tui/graph_service"
// )

// // This example demonstrates how AI coders can leverage the file_path property
// // to locate and retrieve source files for any code entity in the knowledge graph
// func main() {
// 	// Assume we have an analyzed codebase
// 	result, err := codeanalyzer.BuildGraph(codeanalyzer.BuildGraphOptions{
// 		RepoPath: "./my-project",
// 		DBPath:   "./my-project.db",
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer result.Close()

// 	// Example 1: Find the file containing a specific function
// 	fmt.Println("=== Example 1: Locate file for a function ===")
// 	query1 := `
// 		MATCH (f:Function {name: "HandleRequest"})
// 		RETURN f.name as function_name, f.file_path as file_location
// 	`
// 	res1, _ := result.QueryGraph(query1)
// 	fmt.Println(res1)

// 	// Example 2: Get all files that contain a specific class/struct
// 	fmt.Println("\n=== Example 2: Find all files with UserController ===")
// 	query2 := `
// 		MATCH (c:Class {name: "UserController"})
// 		RETURN DISTINCT c.file_path as file_path
// 	`
// 	res2, _ := result.QueryGraph(query2)
// 	fmt.Println(res2)

// 	// Example 3: List all entities in a specific file
// 	fmt.Println("\n=== Example 3: List all entities in main.go ===")
// 	query3 := `
// 		MATCH (n)
// 		WHERE n.file_path ENDS WITH 'main.go'
// 		RETURN labels(n)[0] as entity_type, n.name as name
// 		ORDER BY entity_type, name
// 	`
// 	res3, _ := result.QueryGraph(query3)
// 	fmt.Println(res3)

// 	// Example 4: Find cross-file function calls
// 	fmt.Println("\n=== Example 4: Cross-file function calls ===")
// 	query4 := `
// 		MATCH (caller)-[:CALLS]->(callee)
// 		WHERE caller.file_path <> callee.file_path
// 		RETURN
// 			caller.name as calling_function,
// 			caller.file_path as from_file,
// 			callee.name as called_function,
// 			callee.file_path as to_file
// 		LIMIT 10
// 	`
// 	res4, _ := result.QueryGraph(query4)
// 	fmt.Println(res4)

// 	// Example 5: Find which files import a specific module
// 	fmt.Println("\n=== Example 5: Files importing 'fmt' package ===")
// 	query5 := `
// 		MATCH (i:Import)
// 		WHERE i.path = 'fmt' OR i.name = 'fmt'
// 		RETURN DISTINCT i.file_path as importing_file
// 		ORDER BY importing_file
// 	`
// 	res5, _ := result.QueryGraph(query5)
// 	fmt.Println(res5)

// 	// Example 6: Complex query - Find methods and their files for a specific struct
// 	fmt.Println("\n=== Example 6: Methods of Calculator struct ===")
// 	query6 := `
// 		MATCH (s:Struct {name: "Calculator"})<-[:DEFINES]-(m:Method)
// 		RETURN
// 			m.name as method_name,
// 			m.signature as signature,
// 			m.file_path as file_location
// 		ORDER BY m.name
// 	`
// 	res6, _ := result.QueryGraph(query6)
// 	fmt.Println(res6)

// 	// Example 7: File statistics - entities per file
// 	fmt.Println("\n=== Example 7: Entity count per file ===")
// 	query7 := `
// 		MATCH (n)
// 		WHERE n.file_path IS NOT NULL
// 		WITH n.file_path as file, labels(n)[0] as type
// 		RETURN
// 			file,
// 			count(*) as total_entities,
// 			collect(DISTINCT type) as entity_types
// 		ORDER BY total_entities DESC
// 		LIMIT 10
// 	`
// 	res7, _ := result.QueryGraph(query7)
// 	fmt.Println(res7)

// 	// Example 8: Find test files and their tested functions
// 	fmt.Println("\n=== Example 8: Test files and functions ===")
// 	query8 := `
// 		MATCH (f:Function)
// 		WHERE f.file_path CONTAINS '_test.go' OR f.file_path CONTAINS 'test_'
// 		RETURN
// 			f.file_path as test_file,
// 			collect(f.name) as test_functions
// 		ORDER BY test_file
// 	`
// 	res8, _ := result.QueryGraph(query8)
// 	fmt.Println(res8)

// 	// Example 9: AI Coder use case - Get context for code generation
// 	fmt.Println("\n=== Example 9: Context for AI code generation ===")
// 	// When AI needs to generate code for a function, it can get the file context
// 	targetFile := "src/handlers/user.go"
// 	query9 := fmt.Sprintf(`
// 		MATCH (n)
// 		WHERE n.file_path = '%s'
// 		RETURN
// 			labels(n)[0] as type,
// 			n.name as name,
// 			CASE
// 				WHEN n.signature IS NOT NULL THEN n.signature
// 				ELSE 'N/A'
// 			END as signature
// 		ORDER BY
// 			CASE labels(n)[0]
// 				WHEN 'Import' THEN 1
// 				WHEN 'Struct' THEN 2
// 				WHEN 'Interface' THEN 3
// 				WHEN 'Function' THEN 4
// 				WHEN 'Method' THEN 5
// 				ELSE 6
// 			END,
// 			n.name
// 	`, targetFile)
// 	res9, _ := result.QueryGraph(query9)
// 	fmt.Println("File context for:", targetFile)
// 	fmt.Println(res9)

// 	// Example 10: Find related files for refactoring
// 	fmt.Println("\n=== Example 10: Related files for refactoring ===")
// 	query10 := `
// 		MATCH (source:File {path: 'src/models/user.go'})-[:Contains]->(entity)
// 		MATCH (entity)-[rel]-(related)
// 		WHERE related.file_path <> 'src/models/user.go'
// 		RETURN DISTINCT
// 			related.file_path as related_file,
// 			collect(DISTINCT type(rel)) as relationship_types
// 		ORDER BY related_file
// 	`
// 	res10, _ := result.QueryGraph(query10)
// 	fmt.Println(res10)
// }

// // Helper function for AI coders to get file path for any entity
// func GetEntityFilePath(result *codeanalyzer.BuildGraphResult, entityName string) (string, error) {
// 	query := fmt.Sprintf(`
// 		MATCH (n {name: "%s"})
// 		RETURN n.file_path as file_path
// 		LIMIT 1
// 	`, entityName)

// 	res, err := result.QueryGraph(query)
// 	if err != nil {
// 		return "", err
// 	}

// 	return res, nil
// }

// // Helper function to get all entities from a specific file
// func GetFileEntities(result *codeanalyzer.BuildGraphResult, filePath string) (string, error) {
// 	query := fmt.Sprintf(`
// 		MATCH (n)
// 		WHERE n.file_path = '%s'
// 		RETURN
// 			labels(n)[0] as type,
// 			n.name as name,
// 			n.id as id
// 		ORDER BY type, name
// 	`, filePath)

// 	return result.QueryGraph(query)
// }
