package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"os"
// 	"strings"

// 	codeanalyzer "github.com/onyx/onyx-tui/graph_service"
// )

// // TestCoverageExampleMain demonstrates comprehensive test coverage analysis
// // using the KuzuDB embedded graph with complete test detection and metrics
// func main() {
// 	fmt.Println("🧪 Comprehensive Test Coverage Analysis Example")
// 	fmt.Println("===============================================")

// 	// Create a temporary directory for the database
// 	tempDir, err := os.MkdirTemp("", "test_coverage_*")
// 	if err != nil {
// 		log.Fatalf("Failed to create temp dir: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	fmt.Printf("📁 Using database directory: %s\n\n", tempDir)

// 	// Analyze the current project (which includes test files)
// 	fmt.Println("🔍 Analyzing codebase with test detection...")
// 	result, err := codeanalyzer.BuildGraph(codeanalyzer.BuildGraphOptions{
// 		RepoPath:    ".", // Analyze current directory
// 		DBPath:      tempDir,
// 		CleanupDB:   false,
// 		LoadEnvFile: true,
// 	})
// 	if err != nil {
// 		log.Fatalf("❌ Failed to build graph: %v", err)
// 	}
// 	defer result.Close()

// 	fmt.Printf("✅ Analysis complete!\n")
// 	fmt.Printf("   📊 Files: %d, Functions: %d, Classes: %d\n\n",
// 		result.Stats.FilesCount, result.Stats.FunctionsCount, result.Stats.ClassesCount)

// 	// Demonstrate comprehensive coverage analysis
// 	demonstrateCoverageAnalysis(result)

// 	// Demonstrate entity-specific coverage
// 	demonstrateEntityCoverage(result)

// 	// Demonstrate uncovered entity detection
// 	demonstrateUncoveredDetection(result)

// 	// Demonstrate test relationship queries
// 	demonstrateTestQueries(result)

// 	fmt.Println("\n🎉 Test coverage analysis demonstration complete!")
// }

// // demonstrateCoverageAnalysis shows overall coverage metrics
// func demonstrateCoverageAnalysis(result *codeanalyzer.BuildGraphResult) {
// 	fmt.Println("📈 Overall Coverage Metrics")
// 	fmt.Println("==========================")

// 	metrics, err := result.GetCoverageMetrics()
// 	if err != nil {
// 		fmt.Printf("❌ Error getting coverage metrics: %v\n", err)
// 		return
// 	}

// 	// Display comprehensive metrics
// 	fmt.Printf("🎯 Coverage Summary:\n")
// 	fmt.Printf("   Total Entities: %d\n", metrics.TotalEntities)
// 	fmt.Printf("   Tested Entities: %d\n", metrics.TestedEntities)
// 	fmt.Printf("   Untested Entities: %d\n", metrics.UntestedEntities)
// 	fmt.Printf("   Coverage Percentage: %.1f%%\n", metrics.CoveragePercentage)
// 	fmt.Printf("   Total Tests: %d\n", metrics.TestCount)
// 	fmt.Printf("   Total Assertions: %d\n", metrics.TotalAssertions)

// 	// Test framework breakdown
// 	fmt.Printf("\n🛠️  Test Frameworks:\n")
// 	for framework, count := range metrics.TestFrameworks {
// 		fmt.Printf("   %s: %d tests\n", framework, count)
// 	}

// 	// Test type breakdown
// 	fmt.Printf("\n📝 Test Types:\n")
// 	for testType, count := range metrics.TestTypes {
// 		fmt.Printf("   %s: %d tests\n", testType, count)
// 	}

// 	// Coverage by entity type
// 	fmt.Printf("\n🏗️  Coverage by Entity Type:\n")
// 	for entityType, coverage := range metrics.CoverageByType {
// 		fmt.Printf("   %s: %.1f%%\n", entityType, coverage)
// 	}

// 	// Coverage quality score
// 	if qualityScore, ok := metrics.Details["coverage_quality_score"].(float64); ok {
// 		fmt.Printf("\n⭐ Coverage Quality Score: %.2f/1.0\n", qualityScore)
// 	}

// 	// Test to production ratio
// 	if testRatio, ok := metrics.Details["test_to_production_ratio"].(float64); ok {
// 		fmt.Printf("📊 Test to Production Ratio: %.2f\n", testRatio)
// 	}

// 	// Average assertions per test
// 	if avgAssertions, ok := metrics.Details["average_assertions_per_test"].(float64); ok {
// 		fmt.Printf("🔍 Average Assertions per Test: %.1f\n", avgAssertions)
// 	}

// 	fmt.Println()
// }

// // demonstrateEntityCoverage shows detailed coverage for specific entities
// func demonstrateEntityCoverage(result *codeanalyzer.BuildGraphResult) {
// 	fmt.Println("🎯 Entity-Specific Coverage Analysis")
// 	fmt.Println("====================================")

// 	// Get all entities and analyze a few examples
// 	allEntities := result.GetAllEntities()

// 	count := 0
// 	for _, entity := range allEntities {
// 		// Skip test entities and analyze only production code
// 		if entity.IsTest() || entity.IsTestFile() {
// 			continue
// 		}

// 		// Limit to first 3 entities for demonstration
// 		if count >= 3 {
// 			break
// 		}

// 		fmt.Printf("\n📍 Entity: %s (Type: %s)\n", entity.Name, entity.Type)
// 		fmt.Printf("   📁 File: %s\n", entity.FilePath)

// 		coverage, err := result.GetTestCoverage(entity.ID)
// 		if err != nil {
// 			fmt.Printf("   ❌ Error analyzing coverage: %v\n", err)
// 			continue
// 		}

// 		fmt.Printf("   🎯 Coverage Status: %v\n", coverage.IsCovered)
// 		fmt.Printf("   📊 Coverage Score: %.2f\n", coverage.CoverageScore)
// 		fmt.Printf("   🧪 Test Count: %d\n", coverage.TestCount)
// 		fmt.Printf("   🔍 Assertion Count: %d\n", coverage.AssertionCount)

// 		// Direct tests
// 		if len(coverage.DirectTests) > 0 {
// 			fmt.Printf("   ✅ Direct Tests (%d):\n", len(coverage.DirectTests))
// 			for _, test := range coverage.DirectTests {
// 				fmt.Printf("      - %s (%s, confidence: %.2f)\n",
// 					test.TestName, test.TestFramework, test.ConfidenceScore)
// 			}
// 		}

// 		// Indirect tests
// 		if len(coverage.IndirectTests) > 0 {
// 			fmt.Printf("   🔗 Indirect Tests (%d):\n", len(coverage.IndirectTests))
// 			for _, test := range coverage.IndirectTests {
// 				fmt.Printf("      - %s (%s, %s coverage)\n",
// 					test.TestName, test.TestFramework, test.CoverageType)
// 			}
// 		}

// 		// Coverage metrics
// 		fmt.Printf("   📈 Coverage Metrics:\n")
// 		for key, value := range coverage.CoverageMetrics {
// 			fmt.Printf("      %s: %v\n", key, value)
// 		}

// 		count++
// 	}

// 	fmt.Println()
// }

// // demonstrateUncoveredDetection finds and reports entities without tests
// func demonstrateUncoveredDetection(result *codeanalyzer.BuildGraphResult) {
// 	fmt.Println("🚨 Uncovered Entity Detection")
// 	fmt.Println("=============================")

// 	uncovered, err := result.GetUncoveredEntities()
// 	if err != nil {
// 		fmt.Printf("❌ Error finding uncovered entities: %v\n", err)
// 		return
// 	}

// 	fmt.Printf("Found %d uncovered entities:\n\n", len(uncovered))

// 	// Group by file for better organization
// 	fileGroups := make(map[string][]*codeanalyzer.Entity)
// 	for _, entity := range uncovered {
// 		fileGroups[entity.FilePath] = append(fileGroups[entity.FilePath], entity)
// 	}

// 	for filePath, entities := range fileGroups {
// 		fmt.Printf("📁 %s:\n", filePath)
// 		for _, entity := range entities {
// 			fmt.Printf("   ❌ %s (%s)\n", entity.Name, entity.Type)
// 		}
// 		fmt.Println()
// 	}

// 	if len(uncovered) == 0 {
// 		fmt.Println("🎉 All entities have test coverage!")
// 	}

// 	fmt.Println()
// }

// // demonstrateTestQueries shows advanced querying capabilities
// func demonstrateTestQueries(result *codeanalyzer.BuildGraphResult) {
// 	fmt.Println("🔍 Advanced Test Coverage Queries")
// 	fmt.Println("=================================")

// 	// Query 1: Find all test functions
// 	fmt.Println("📝 All Test Functions:")
// 	testQuery := `
// 		MATCH (test:TestFunction)
// 		RETURN test.name, test.test_type, test.test_framework, test.file_path
// 		ORDER BY test.file_path, test.name
// 		LIMIT 10
// 	`

// 	testResult, err := result.QueryGraph(testQuery)
// 	if err != nil {
// 		fmt.Printf("❌ Error querying tests: %v\n", err)
// 	} else if testResult != "" {
// 		lines := splitLines(testResult)
// 		for _, line := range lines {
// 			parts := splitFields(line)
// 			if len(parts) >= 4 {
// 				fmt.Printf("   🧪 %s (%s, %s) in %s\n",
// 					cleanField(parts[0]), cleanField(parts[1]),
// 					cleanField(parts[2]), cleanField(parts[3]))
// 			}
// 		}
// 	} else {
// 		fmt.Println("   No test functions found in the database.")
// 	}

// 	fmt.Println()

// 	// Query 2: Find coverage relationships
// 	fmt.Println("🔗 Test Coverage Relationships:")
// 	coverageQuery := `
// 		MATCH (test)-[:TESTS]->(target)
// 		RETURN test.name, target.name, target.file_path
// 		LIMIT 10
// 	`

// 	coverageResult, err := result.QueryGraph(coverageQuery)
// 	if err != nil {
// 		fmt.Printf("❌ Error querying coverage: %v\n", err)
// 	} else if coverageResult != "" {
// 		lines := splitLines(coverageResult)
// 		for _, line := range lines {
// 			parts := splitFields(line)
// 			if len(parts) >= 3 {
// 				fmt.Printf("   🎯 %s tests %s (in %s)\n",
// 					cleanField(parts[0]), cleanField(parts[1]), cleanField(parts[2]))
// 			}
// 		}
// 	} else {
// 		fmt.Println("   No test coverage relationships found.")
// 	}

// 	fmt.Println()

// 	// Query 3: Test statistics by framework
// 	fmt.Println("📊 Test Statistics by Framework:")
// 	frameworkQuery := `
// 		MATCH (test:TestFunction)
// 		RETURN test.test_framework, count(test) as test_count
// 		ORDER BY test_count DESC
// 	`

// 	frameworkResult, err := result.QueryGraph(frameworkQuery)
// 	if err != nil {
// 		fmt.Printf("❌ Error querying frameworks: %v\n", err)
// 	} else if frameworkResult != "" {
// 		lines := splitLines(frameworkResult)
// 		for _, line := range lines {
// 			parts := splitFields(line)
// 			if len(parts) >= 2 {
// 				fmt.Printf("   🛠️  %s: %s tests\n", cleanField(parts[0]), cleanField(parts[1]))
// 			}
// 		}
// 	} else {
// 		fmt.Println("   No framework statistics available.")
// 	}

// 	fmt.Println()

// 	// Query 4: Functions without tests
// 	fmt.Println("🚨 Functions Without Direct Tests:")
// 	uncoveredQuery := `
// 		MATCH (f:Function)
// 		WHERE NOT EXISTS((test:TestFunction)-[:TESTS]->(f))
// 		AND NOT f.file_path CONTAINS '_test.go'
// 		AND NOT f.file_path CONTAINS 'test_'
// 		RETURN f.name, f.file_path
// 		LIMIT 10
// 	`

// 	uncoveredResult, err := result.QueryGraph(uncoveredQuery)
// 	if err != nil {
// 		fmt.Printf("❌ Error querying uncovered functions: %v\n", err)
// 	} else if uncoveredResult != "" {
// 		lines := splitLines(uncoveredResult)
// 		for _, line := range lines {
// 			parts := splitFields(line)
// 			if len(parts) >= 2 {
// 				fmt.Printf("   ❌ %s (in %s)\n", cleanField(parts[0]), cleanField(parts[1]))
// 			}
// 		}
// 	} else {
// 		fmt.Println("   🎉 All functions have test coverage!")
// 	}

// 	fmt.Println()
// }

// // Helper functions for parsing query results

// func splitLines(result string) []string {
// 	if result == "" {
// 		return []string{}
// 	}
// 	return strings.Split(strings.TrimSpace(result), "\n")
// }

// func splitFields(line string) []string {
// 	if line == "" {
// 		return []string{}
// 	}
// 	return strings.Split(line, "\t|\t")
// }

// func cleanField(field string) string {
// 	// Remove quotes and trim whitespace
// 	cleaned := strings.Trim(field, "\"")
// 	cleaned = strings.TrimSpace(cleaned)
// 	if cleaned == "" {
// 		return "unknown"
// 	}
// 	return cleaned
// }

// // Additional demonstration functions

// // demonstrateAPIUsage shows how to use the coverage API programmatically
// func demonstrateAPIUsage(result *codeanalyzer.BuildGraphResult) {
// 	fmt.Println("🔧 Programmatic API Usage Examples")
// 	fmt.Println("==================================")

// 	// Example 1: Get coverage for specific entity
// 	entities := result.GetEntitiesByName("main")
// 	if len(entities) > 0 {
// 		entity := entities[0]
// 		fmt.Printf("🎯 Analyzing coverage for function: %s\n", entity.Name)

// 		coverage, err := result.GetTestCoverage(entity.ID)
// 		if err != nil {
// 			fmt.Printf("❌ Error: %v\n", err)
// 		} else {
// 			// Convert to JSON for demonstration
// 			jsonData, _ := json.MarshalIndent(coverage, "", "  ")
// 			fmt.Printf("📊 Coverage data (JSON):\n%s\n", string(jsonData))
// 		}
// 	}

// 	fmt.Println()

// 	// Example 2: Get all tests for an entity
// 	if len(entities) > 0 {
// 		entity := entities[0]
// 		tests, err := result.GetTestsByTarget(entity.ID)
// 		if err != nil {
// 			fmt.Printf("❌ Error getting tests: %v\n", err)
// 		} else {
// 			fmt.Printf("🧪 Found %d tests for %s:\n", len(tests), entity.Name)
// 			for _, test := range tests {
// 				fmt.Printf("   - %s (%s)\n", test.Name, test.GetTestFramework())
// 			}
// 		}
// 	}

// 	fmt.Println()
// }
