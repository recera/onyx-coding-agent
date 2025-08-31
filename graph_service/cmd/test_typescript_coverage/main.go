package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/onyx/onyx-tui/graph_service/internal/analyzer"
	"github.com/onyx/onyx-tui/graph_service/internal/db"
	"github.com/onyx/onyx-tui/graph_service/internal/entities"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("TypeScript Test Coverage Analysis Demo")
	fmt.Println("========================================\n")

	// Create a temporary database
	tempDir, err := os.MkdirTemp("", "typescript_test_coverage_*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test_coverage.db")
	fmt.Printf("📁 Database: %s\n\n", dbPath)

	// Initialize database
	database, err := db.NewKuzuDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	// Create schema
	err = database.CreateSchema()
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// Create TypeScript analyzer
	tsAnalyzer := analyzer.NewTypeScriptAnalyzer()

	// Analyze test file
	fmt.Println("🔍 Analyzing TypeScript test file...")
	testFilePath := "internal/graph_service/examples/typescript_test_coverage_demo.spec.ts"
	testContent, err := os.ReadFile(testFilePath)
	if err != nil {
		log.Fatalf("Failed to read test file: %v", err)
	}

	testFile, testRelationships, err := tsAnalyzer.AnalyzeFile(testFilePath, testContent)
	if err != nil {
		log.Fatalf("Failed to analyze test file: %v", err)
	}

	// Analyze production file
	fmt.Println("🔍 Analyzing TypeScript production file...")
	prodFilePath := "internal/graph_service/examples/services/UserService.ts"
	prodContent, err := os.ReadFile(prodFilePath)
	if err != nil {
		// If production file doesn't exist, just analyze test file
		fmt.Println("⚠️  Production file not found, analyzing test file only")
	} else {
		prodFile, prodRelationships, err := tsAnalyzer.AnalyzeFile(prodFilePath, prodContent)
		if err != nil {
			log.Printf("Warning: Failed to analyze production file: %v", err)
		} else {
			// Store production entities
			for _, entity := range prodFile.Entities {
				err = database.StoreEntity(entity)
				if err != nil {
					log.Printf("Failed to store entity %s: %v", entity.Name, err)
				}
			}
			
			// Store production relationships
			for _, rel := range prodRelationships {
				err = database.StoreRelationship(rel)
				if err != nil {
					log.Printf("Failed to store relationship: %v", err)
				}
			}
		}
	}

	// Store test entities in database
	for _, entity := range testFile.Entities {
		err = database.StoreEntity(entity)
		if err != nil {
			log.Printf("Failed to store entity %s: %v", entity.Name, err)
		}
	}

	// Store test relationships
	for _, rel := range testRelationships {
		err = database.StoreRelationship(rel)
		if err != nil {
			log.Printf("Failed to store relationship: %v", err)
		}
	}

	// Display test coverage analysis
	fmt.Println("\n📊 Test Coverage Analysis Results")
	fmt.Println("==================================")

	// Count different types of test entities
	var testSuites, testCases, assertions, mocks int
	var testFramework string
	
	for _, entity := range testFile.Entities {
		switch entity.Type {
		case entities.EntityTypeTestSuite:
			testSuites++
			if fw := entity.GetTestFramework(); fw != "" && testFramework == "" {
				testFramework = fw
			}
		case entities.EntityTypeTestFunction:
			testCases++
		case entities.EntityTypeAssertion:
			assertions++
		case entities.EntityTypeMock:
			mocks++
		}
	}

	fmt.Printf("\n🧪 Test Framework: %s\n", testFramework)
	fmt.Printf("📦 Test Suites: %d\n", testSuites)
	fmt.Printf("✅ Test Cases: %d\n", testCases)
	fmt.Printf("🎯 Assertions: %d\n", assertions)
	fmt.Printf("🎭 Mocks/Spies: %d\n", mocks)

	// Display test details
	fmt.Println("\n📝 Test Suites Detected:")
	fmt.Println("------------------------")
	for _, entity := range testFile.Entities {
		if entity.Type == entities.EntityTypeTestSuite {
			fmt.Printf("  📦 %s\n", entity.Name)
		}
	}

	// Display test cases with details
	fmt.Println("\n🧪 Test Cases Detected:")
	fmt.Println("----------------------")
	testCount := 0
	for _, entity := range testFile.Entities {
		if entity.Type == entities.EntityTypeTestFunction {
			testCount++
			testType := entity.GetTestType()
			assertCount := entity.GetAssertionCount()
			target := entity.GetTestTarget()
			
			fmt.Printf("  %d. %s\n", testCount, entity.Name)
			fmt.Printf("     Type: %s | Assertions: %d", testType, assertCount)
			if target != "" {
				fmt.Printf(" | Target: %s", target)
			}
			fmt.Println()
		}
	}

	// Display test relationships
	fmt.Println("\n🔗 Test Relationships:")
	fmt.Println("---------------------")
	relCount := 0
	for _, rel := range testRelationships {
		if rel.Type == "TESTS" || rel.Type == "COVERS" || rel.Type == "TESTS_COMPONENT" {
			relCount++
			fmt.Printf("  %s → %s (%s)\n", rel.SourceID[:8], rel.TargetID[:8], rel.Type)
			if relCount >= 10 {
				fmt.Printf("  ... and %d more\n", len(testRelationships)-10)
				break
			}
		}
	}

	// Run Cypher queries to demonstrate test coverage
	fmt.Println("\n🔍 Cypher Query Examples:")
	fmt.Println("========================")

	// Query 1: Count test functions
	query1 := `MATCH (t:TestFunction) RETURN COUNT(t) as test_count`
	result1, err := database.ExecuteQuery(query1)
	if err != nil {
		log.Printf("Query 1 failed: %v", err)
	} else {
		fmt.Printf("\n1. Total Test Functions:\n   %s\n", strings.TrimSpace(result1))
	}

	// Query 2: Test suites with test counts
	query2 := `
		MATCH (s:TestSuite)
		RETURN s.name as suite_name, s.test_count as test_count
		ORDER BY s.test_count DESC
		LIMIT 5
	`
	result2, err := database.ExecuteQuery(query2)
	if err != nil {
		log.Printf("Query 2 failed: %v", err)
	} else if result2 != "" {
		fmt.Printf("\n2. Test Suites:\n")
		lines := strings.Split(strings.TrimSpace(result2), "\n")
		for _, line := range lines {
			fmt.Printf("   %s\n", line)
		}
	}

	// Query 3: Test framework statistics
	query3 := `
		MATCH (t:TestFunction)
		WHERE t.test_framework IS NOT NULL
		RETURN t.test_framework as framework, COUNT(t) as count
	`
	result3, err := database.ExecuteQuery(query3)
	if err != nil {
		log.Printf("Query 3 failed: %v", err)
	} else if result3 != "" {
		fmt.Printf("\n3. Test Framework Usage:\n")
		lines := strings.Split(strings.TrimSpace(result3), "\n")
		for _, line := range lines {
			fmt.Printf("   %s\n", line)
		}
	}

	// Query 4: Test types distribution
	query4 := `
		MATCH (t:TestFunction)
		WHERE t.test_type IS NOT NULL
		RETURN t.test_type as type, COUNT(t) as count
		ORDER BY count DESC
	`
	result4, err := database.ExecuteQuery(query4)
	if err != nil {
		log.Printf("Query 4 failed: %v", err)
	} else if result4 != "" {
		fmt.Printf("\n4. Test Types Distribution:\n")
		lines := strings.Split(strings.TrimSpace(result4), "\n")
		for _, line := range lines {
			fmt.Printf("   %s\n", line)
		}
	}

	// Display coverage summary
	fmt.Println("\n📈 Coverage Summary:")
	fmt.Println("===================")
	
	// Calculate simple coverage metrics
	testedFunctions := 0
	
	// This would be more accurate with actual production code analysis
	for _, rel := range testRelationships {
		if rel.Type == "TESTS" || rel.Type == "COVERS" {
			testedFunctions++
		}
	}
	
	fmt.Printf("✅ Test Cases: %d\n", testCases)
	fmt.Printf("🎯 Total Assertions: %d\n", assertions)
	if testCases > 0 {
		avgAssertions := float64(assertions) / float64(testCases)
		fmt.Printf("📊 Average Assertions per Test: %.1f\n", avgAssertions)
	}
	fmt.Printf("🎭 Mock Usage: %d\n", mocks)
	
	// Test quality indicators
	fmt.Println("\n⭐ Test Quality Indicators:")
	if testFramework != "" && testFramework != "unknown" {
		fmt.Println("  ✅ Framework detected:", testFramework)
	}
	if testSuites > 0 {
		fmt.Println("  ✅ Tests are well organized in suites")
	}
	if assertions > testCases {
		fmt.Println("  ✅ Multiple assertions per test (thorough testing)")
	}
	if mocks > 0 {
		fmt.Println("  ✅ Uses mocking for isolation")
	}

	fmt.Println("\n✨ TypeScript test coverage analysis complete!")
}