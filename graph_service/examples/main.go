package main

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"path/filepath"

// 	codeanalyzer "github.com/onyx/onyx-tui/graph_service"
// )

// func main() {
// 	// Create a test Python repository
// 	tempDir, err := os.MkdirTemp("", "example_repo_*")
// 	if err != nil {
// 		log.Fatalf("Failed to create temp dir: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	// Create a complex Python example
// 	pythonCode := `"""
// Example module demonstrating the code analyzer capabilities.
// """

// import math
// import os

// class Calculator:
//     """A simple calculator with basic operations."""

//     def __init__(self, precision=2):
//         """Initialize calculator with specified precision."""
//         self.precision = precision
//         self.history = []

//     def add(self, a, b):
//         """Add two numbers."""
//         result = a + b
//         self.record_operation("add", a, b, result)
//         return round(result, self.precision)

//     def multiply(self, a, b):
//         """Multiply two numbers."""
//         result = a * b
//         self.record_operation("multiply", a, b, result)
//         return round(result, self.precision)

//     def power(self, base, exponent):
//         """Calculate power using math.pow."""
//         result = math.pow(base, exponent)
//         self.record_operation("power", base, exponent, result)
//         return round(result, self.precision)

//     def record_operation(self, op, a, b, result):
//         """Record operation in history."""
//         self.history.append({
//             "operation": op,
//             "inputs": [a, b],
//             "result": result
//         })

// class ScientificCalculator(Calculator):
//     """Extended calculator with scientific functions."""

//     def __init__(self, precision=4):
//         super().__init__(precision)
//         self.mode = "radians"

//     def sin(self, x):
//         """Calculate sine."""
//         result = math.sin(x)
//         self.record_operation("sin", x, 0, result)
//         return round(result, self.precision)

//     def cos(self, x):
//         """Calculate cosine."""
//         result = math.cos(x)
//         self.record_operation("cos", x, 0, result)
//         return round(result, self.precision)

// def calculate_area(radius):
//     """Calculate circle area using a calculator."""
//     calc = Calculator()
//     pi = 3.14159
//     radius_squared = calc.multiply(radius, radius)
//     area = calc.multiply(pi, radius_squared)
//     return area

// def main():
//     """Demonstrate calculator usage."""
//     print("Basic Calculator Demo")

//     # Basic calculator
//     calc = Calculator(precision=3)
//     result1 = calc.add(10, 5)
//     result2 = calc.multiply(result1, 2)

//     print(f"10 + 5 = {result1}")
//     print(f"{result1} * 2 = {result2}")

//     # Scientific calculator
//     sci_calc = ScientificCalculator()
//     angle = 1.57  # roughly π/2
//     sin_result = sci_calc.sin(angle)
//     cos_result = sci_calc.cos(angle)

//     print(f"sin({angle}) = {sin_result}")
//     print(f"cos({angle}) = {cos_result}")

//     # Calculate area
//     circle_area = calculate_area(5)
//     print(f"Circle area (radius=5): {circle_area}")

// if __name__ == "__main__":
//     main()
// `

// 	// Write the Python file
// 	testFile := filepath.Join(tempDir, "calculator.py")
// 	err = os.WriteFile(testFile, []byte(pythonCode), 0644)
// 	if err != nil {
// 		log.Fatalf("Failed to write test file: %v", err)
// 	}

// 	fmt.Printf("Created test repository in: %s\n", tempDir)
// 	fmt.Println("Analyzing repository with new entity system...")

// 	// Build the graph with new API
// 	result, err := codeanalyzer.BuildGraph(codeanalyzer.BuildGraphOptions{
// 		RepoPath:    tempDir,
// 		CleanupDB:   true,
// 		LoadEnvFile: true,
// 	})
// 	if err != nil {
// 		log.Fatalf("Failed to build graph: %v", err)
// 	}
// 	defer result.Close()

// 	// Display basic statistics
// 	fmt.Printf("\n=== Analysis Statistics ===\n")
// 	fmt.Printf("Files processed: %d\n", result.Stats.FilesCount)
// 	fmt.Printf("Functions found: %d\n", result.Stats.FunctionsCount)
// 	fmt.Printf("Classes found: %d\n", result.Stats.ClassesCount)
// 	fmt.Printf("Methods found: %d\n", result.Stats.MethodsCount)
// 	fmt.Printf("Total relationships: %d\n", result.Stats.CallsCount)
// 	fmt.Printf("Errors encountered: %d\n", result.Stats.ErrorsCount)

// 	// Get detailed analysis results using new API
// 	analysis := result.GetAnalysisResult()

// 	fmt.Printf("\n=== Detailed Entity Analysis ===\n")
// 	fmt.Printf("Total entities extracted: %d\n", len(analysis.Entities))

// 	// Show all classes and their methods
// 	fmt.Printf("\n--- Classes and Methods ---\n")
// 	for _, entity := range analysis.Entities {
// 		if entity.Type == "Class" {
// 			fmt.Printf("Class: %s\n", entity.Name)
// 			if entity.DocString != "" {
// 				fmt.Printf("  Description: %s\n", entity.DocString[:min(50, len(entity.DocString))]+"...")
// 			}

// 			// Find methods for this class
// 			for _, child := range entity.Children {
// 				if child.Type == "Method" || child.Type == "Function" {
// 					fmt.Printf("  - %s%s\n", child.Name, child.Signature)
// 				}
// 			}
// 			fmt.Println()
// 		}
// 	}

// 	// Show function call relationships
// 	fmt.Printf("--- Function Call Relationships ---\n")
// 	callCount := 0
// 	for _, rel := range analysis.Relationships {
// 		if rel.Type == "CALLS" && callCount < 10 { // Show first 10
// 			fmt.Printf("  %s\n", rel.String())
// 			callCount++
// 		}
// 	}

// 	// Demonstrate entity lookup by name
// 	fmt.Printf("\n--- Entity Lookup Examples ---\n")
// 	calculators := result.GetEntityByName("Calculator")
// 	for _, calc := range calculators {
// 		fmt.Printf("Found entity '%s' of type %s in file %s\n",
// 			calc.Name, calc.Type, filepath.Base(calc.FilePath))
// 	}

// 	// Show file-level analysis
// 	fmt.Printf("\n--- File Analysis ---\n")
// 	for filePath, file := range result.GetAllFiles() {
// 		fmt.Printf("File: %s\n", filepath.Base(filePath))
// 		stats := file.GetStats()
// 		fmt.Printf("  Functions: %d, Classes: %d, Methods: %d\n",
// 			stats.Functions, stats.Classes, stats.Methods)
// 	}

// 	// Demonstrate database querying
// 	fmt.Printf("\n=== Database Queries ===\n")

// 	// Query for all classes
// 	fmt.Println("Querying for all classes...")
// 	classResult, err := result.QueryGraph("MATCH (c:Class) RETURN c.name ORDER BY c.name")
// 	if err != nil {
// 		fmt.Printf("Query error: %v\n", err)
// 	} else {
// 		fmt.Printf("Classes found:\n%s\n", classResult)
// 	}

// 	// Query for all functions
// 	fmt.Println("Querying for functions with their signatures...")
// 	funcResult, err := result.QueryGraph("MATCH (f:Function) RETURN f.name, f.signature LIMIT 5")
// 	if err != nil {
// 		fmt.Printf("Query error: %v\n", err)
// 	} else {
// 		fmt.Printf("Functions found:\n%s\n", funcResult)
// 	}

// 	fmt.Println("\n🎉 New API demonstration completed successfully!")
// }

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }
