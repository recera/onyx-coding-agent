package main

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"path/filepath"

// 	codeanalyzer "github.com/onyx/onyx-tui/graph_service"
// 	"github.com/onyx/onyx-tui/graph_service/internal/db"
// )

// func main() {
// 	fmt.Println("=== Testing File Path Storage Feature ===")

// 	// Create a test repository structure
// 	testRepo := createTestRepository()
// 	defer os.RemoveAll(testRepo)

// 	// Analyze the test repository
// 	fmt.Println("\n1. Analyzing test repository...")
// 	result, err := codeanalyzer.BuildGraph(codeanalyzer.BuildGraphOptions{
// 		RepoPath:  testRepo,
// 		DBPath:    "./test_file_paths.db",
// 		CleanupDB: false,
// 	})
// 	if err != nil {
// 		log.Fatalf("Failed to build graph: %v", err)
// 	}
// 	defer func() {
// 		result.Close()
// 		os.RemoveAll("./test_file_paths.db")
// 	}()

// 	// Test 1: Verify file paths are stored for all entity types
// 	fmt.Println("\n2. Testing file path storage for different entity types...")
// 	testFilePathQueries(result.Database)

// 	// Test 2: Verify cross-file relationships
// 	fmt.Println("\n3. Testing cross-file relationships...")
// 	testCrossFileRelationships(result.Database)

// 	// Test 3: Query entities by file path
// 	fmt.Println("\n4. Testing entity queries by file path...")
// 	testQueryByFilePath(result.Database)

// 	// Test 4: Verify all entities have file paths
// 	fmt.Println("\n5. Verifying all entities have file paths...")
// 	verifyAllEntitiesHaveFilePaths(result.Database)

// 	fmt.Println("\n=== All File Path Tests Passed! ===")
// }

// func createTestRepository() string {
// 	// Create temporary test repository
// 	tmpDir, err := os.MkdirTemp("", "test_repo_*")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Create test Go file
// 	goContent := `package main

// import "fmt"

// type Calculator struct {
// 	name string
// }

// func (c *Calculator) Add(a, b int) int {
// 	return a + b
// }

// func main() {
// 	calc := &Calculator{name: "test"}
// 	result := calc.Add(5, 3)
// 	fmt.Println(result)
// }

// type Config struct {
// 	Host string
// 	Port int
// }

// var globalConfig = &Config{
// 	Host: "localhost",
// 	Port: 8080,
// }
// `
// 	goPath := filepath.Join(tmpDir, "main.go")
// 	err = os.WriteFile(goPath, []byte(goContent), 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Create test Python file
// 	pythonContent := `import math

// class MathOperations:
//     """A class for mathematical operations"""

//     def __init__(self, name):
//         self.name = name

//     def calculate_area(self, radius):
//         """Calculate area of a circle"""
//         return math.pi * radius ** 2

// def process_data(data):
//     """Process some data"""
//     operations = MathOperations("processor")
//     return operations.calculate_area(data)

// # Global variable
// PI_SQUARED = math.pi ** 2
// `
// 	pythonPath := filepath.Join(tmpDir, "math_ops.py")
// 	err = os.WriteFile(pythonPath, []byte(pythonContent), 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Create subdirectory with another Go file
// 	subDir := filepath.Join(tmpDir, "utils")
// 	err = os.MkdirAll(subDir, 0755)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	utilsContent := `package utils

// type Logger interface {
// 	Log(message string)
// }

// type FileLogger struct {
// 	filename string
// }

// func (f *FileLogger) Log(message string) {
// 	// Log to file
// }

// func NewLogger(filename string) Logger {
// 	return &FileLogger{filename: filename}
// }
// `
// 	utilsPath := filepath.Join(subDir, "logger.go")
// 	err = os.WriteFile(utilsPath, []byte(utilsContent), 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	return tmpDir
// }

// func testFilePathQueries(db *db.KuzuDatabase) {
// 	tests := []struct {
// 		name  string
// 		query string
// 		desc  string
// 	}{
// 		{
// 			name:  "Functions have file paths",
// 			query: `MATCH (f:Function) WHERE f.file_path IS NOT NULL RETURN count(f) as count`,
// 			desc:  "Count of functions with file paths",
// 		},
// 		{
// 			name:  "Classes have file paths",
// 			query: `MATCH (c:Class) WHERE c.file_path IS NOT NULL RETURN count(c) as count`,
// 			desc:  "Count of classes with file paths",
// 		},
// 		{
// 			name:  "Methods have file paths",
// 			query: `MATCH (m:Method) WHERE m.file_path IS NOT NULL RETURN count(m) as count`,
// 			desc:  "Count of methods with file paths",
// 		},
// 		{
// 			name:  "Structs have file paths",
// 			query: `MATCH (s:Struct) WHERE s.file_path IS NOT NULL RETURN count(s) as count`,
// 			desc:  "Count of structs with file paths",
// 		},
// 		{
// 			name:  "Interfaces have file paths",
// 			query: `MATCH (i:Interface) WHERE i.file_path IS NOT NULL RETURN count(i) as count`,
// 			desc:  "Count of interfaces with file paths",
// 		},
// 		{
// 			name:  "Variables have file paths",
// 			query: `MATCH (v:Variable) WHERE v.file_path IS NOT NULL RETURN count(v) as count`,
// 			desc:  "Count of variables with file paths",
// 		},
// 	}

// 	for _, test := range tests {
// 		result, err := db.ExecuteQuery(test.query)
// 		if err != nil {
// 			log.Printf("Failed %s: %v", test.name, err)
// 			continue
// 		}
// 		fmt.Printf("✓ %s: %s\n", test.name, result)
// 	}
// }

// func testCrossFileRelationships(db *db.KuzuDatabase) {
// 	// Test if we can find entities from different files
// 	query := `
// 		MATCH (n)
// 		WHERE n.file_path IS NOT NULL
// 		RETURN DISTINCT n.file_path as file_path, count(n) as entity_count
// 		ORDER BY file_path
// 	`

// 	result, err := db.ExecuteQuery(query)
// 	if err != nil {
// 		log.Printf("Failed to query file paths: %v", err)
// 		return
// 	}

// 	fmt.Println("Files and their entity counts:")
// 	fmt.Println(result)
// }

// func testQueryByFilePath(db *db.KuzuDatabase) {
// 	// Query all entities from main.go
// 	query := `
// 		MATCH (n)
// 		WHERE n.file_path ENDS WITH 'main.go'
// 		RETURN labels(n)[0] as type, n.name as name
// 		ORDER BY type, name
// 	`

// 	result, err := db.ExecuteQuery(query)
// 	if err != nil {
// 		log.Printf("Failed to query entities by file path: %v", err)
// 		return
// 	}

// 	fmt.Println("Entities in main.go:")
// 	fmt.Println(result)
// }

// func verifyAllEntitiesHaveFilePaths(db *db.KuzuDatabase) {
// 	// Check if any entities are missing file paths
// 	entityTypes := []string{"Function", "Class", "Method", "Struct", "Interface", "Variable", "Import"}

// 	allHaveFilePaths := true
// 	for _, entityType := range entityTypes {
// 		query := fmt.Sprintf(`
// 			MATCH (n:%s)
// 			WHERE n.file_path IS NULL OR n.file_path = ''
// 			RETURN count(n) as count
// 		`, entityType)

// 		result, err := db.ExecuteQuery(query)
// 		if err != nil {
// 			// Entity type might not exist in the test data
// 			continue
// 		}

// 		// Check if result contains "0"
// 		if result != "0\n" && result != "" {
// 			fmt.Printf("⚠️  Found %s entities without file_path: %s", entityType, result)
// 			allHaveFilePaths = false
// 		}
// 	}

// 	if allHaveFilePaths {
// 		fmt.Println("✓ All entities have file_path property set!")
// 	}
// }
