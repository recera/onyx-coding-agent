package main

import (
	"fmt"
	"log"
	"os"

	"github.com/onyx/onyx-tui/graph_service/internal/analyzer"
	"github.com/onyx/onyx-tui/graph_service/internal/db"
)

func main() {
	fmt.Println("=== Phase 3: Advanced Cross-Language Analysis Demo ===")

	// Initialize database
	dbPath := "phase3_test.db"
	defer os.Remove(dbPath)

	kuzuDB, err := db.NewKuzuDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer kuzuDB.Close()

	// Initialize schema
	err = kuzuDB.CreateSchema()
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// Test 1: Advanced Go Features
	testAdvancedGoFeatures(kuzuDB)

	// Test 2: Cross-Language Analysis
	testCrossLanguageAnalysis(kuzuDB)

	fmt.Println("\n🎉 Phase 3 Demo Complete!")
}

func testAdvancedGoFeatures(kuzuDB *db.KuzuDatabase) {
	fmt.Println("\n=== Advanced Go Features Test ===")

	// Create sample code with advanced Go features
	advancedGoCode := `package main

import (
	"context"
	"sync"
)

// Generic type with constraints
type Cache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{items: make(map[K]V)}
}

// Worker pool with channels and goroutines
type WorkerPool struct {
	jobs    chan Job
	results chan Result
	workers int
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	for {
		select {
		case job := <-wp.jobs:
			result := processJob(job)
			wp.results <- result
		case <-ctx.Done():
			return
		}
	}
}

// Pipeline pattern
func pipeline(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for x := range input {
			output <- x * 2
		}
	}()
	return output
}
`

	// Analyze with advanced Go analyzer
	advancedAnalyzer := analyzer.NewAdvancedGoAnalyzer()
	file, relationships, err := advancedAnalyzer.AnalyzeFile("advanced_test.go", []byte(advancedGoCode))
	if err != nil {
		log.Printf("Failed to analyze advanced Go code: %v", err)
		return
	}

	// Get advanced results
	results := advancedAnalyzer.GetAdvancedAnalysisResults()

	fmt.Printf("✅ Advanced Go Analysis Results:\n")
	fmt.Printf("   • Entities found: %d\n", len(file.GetAllEntities()))
	fmt.Printf("   • Relationships found: %d\n", len(relationships))
	fmt.Printf("   • Channels detected: %d\n", len(results.Channels))
	fmt.Printf("   • Goroutines detected: %d\n", len(results.Goroutines))
	fmt.Printf("   • Generics detected: %d\n", len(results.Generics))
	fmt.Printf("   • Concurrency patterns: %d\n", len(results.ConcurrencyPatterns))

	// Store entities in database
	for _, entity := range file.GetAllEntities() {
		err := kuzuDB.StoreEntity(entity)
		if err != nil {
			log.Printf("Failed to store entity %s: %v", entity.Name, err)
		}
	}
	for _, rel := range relationships {
		err := kuzuDB.StoreRelationship(rel)
		if err != nil {
			log.Printf("Failed to store relationship: %v", err)
		}
	}

	// Print detected patterns
	for _, pattern := range results.ConcurrencyPatterns {
		fmt.Printf("   🔄 Pattern: %s - %s\n", pattern.Type, pattern.Description)
	}
}

func testCrossLanguageAnalysis(kuzuDB *db.KuzuDatabase) {
	fmt.Println("\n=== Cross-Language Analysis Test ===")

	// Test cross-language analyzer
	crossAnalyzer := analyzer.NewCrossLanguageAnalyzer()

	// Test with the examples directory
	projectPath := "examples"
	analysis, err := crossAnalyzer.AnalyzeProject(projectPath)
	if err != nil {
		log.Printf("Cross-language analysis failed: %v", err)
		return
	}

	fmt.Printf("✅ Cross-Language Results:\n")
	fmt.Printf("   • Total files: %d\n", analysis.Stats.TotalFiles)
	fmt.Printf("   • Python files: %d\n", analysis.Stats.PythonFiles)
	fmt.Printf("   • Go files: %d\n", analysis.Stats.GoFiles)
	fmt.Printf("   • HTTP endpoints: %d\n", analysis.Stats.HTTPEndpoints)
	fmt.Printf("   • API calls: %d\n", analysis.Stats.APICalls)
	fmt.Printf("   • Cross-references: %d\n", analysis.Stats.CrossReferences)

	// Print endpoints and API calls
	for key, endpoint := range analysis.HTTPEndpoints {
		fmt.Printf("   🌐 Endpoint: %s (%s)\n", key, endpoint.Language)
	}

	for key, call := range analysis.APICalls {
		fmt.Printf("   📞 API Call: %s (%s)\n", key, call.Language)
	}

	// Test database queries
	testDatabaseQueries(kuzuDB)
}

func testDatabaseQueries(kuzuDB *db.KuzuDatabase) {
	fmt.Println("\n=== Database Query Test ===")

	queries := []struct {
		name  string
		query string
	}{
		{"Total Entities", "MATCH (e:Entity) RETURN COUNT(e)"},
		{"Go Functions", "MATCH (e:Entity) WHERE e.type = 'Function' RETURN COUNT(e)"},
		{"Advanced Features", "MATCH (e:Entity)-[r:Relationship]->() WHERE r.advanced_feature IS NOT NULL RETURN COUNT(r)"},
	}

	for _, q := range queries {
		result, err := kuzuDB.ExecuteQuery(q.query)
		if err != nil {
			fmt.Printf("   ❌ %s: Query failed\n", q.name)
		} else {
			// Parse result string for count queries
			if len(result) > 0 {
				fmt.Printf("   ✅ %s: Found results\n", q.name)
			} else {
				fmt.Printf("   ✅ %s: No results\n", q.name)
			}
		}
	}
}
