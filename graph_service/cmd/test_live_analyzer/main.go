package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/onyx/onyx-tui/graph_service/internal/analyzer"
	"github.com/onyx/onyx-tui/graph_service/internal/db"
)

func main() {
	fmt.Println("=== Live Graph Analysis Demo for AI Coding Agents ===")

	// Initialize database
	dbPath := filepath.Join(".", "live_analysis.db")
	database, err := db.NewKuzuDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	// Initialize schema
	err = database.CreateSchema()
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// Set up watch options
	watchOptions := analyzer.DefaultWatchOptions()
	watchOptions.DebounceInterval = 200 * time.Millisecond // Faster for AI agents
	watchOptions.WatchedExtensions = []string{".go", ".py"}
	watchOptions.IgnorePatterns = append(watchOptions.IgnorePatterns, "*.db", "test_live_analyzer")

	// Create live analyzer
	liveAnalyzer, err := analyzer.NewLiveAnalyzer(database, watchOptions)
	if err != nil {
		log.Fatalf("Failed to create live analyzer: %v", err)
	}
	defer liveAnalyzer.StopWatching()

	// Set up AI agent integration callbacks
	setupAIAgentCallbacks(liveAnalyzer)

	// Start watching the current directory and examples
	watchPaths := []string{
		"./examples",
		"./internal",
	}

	for _, path := range watchPaths {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Starting to watch: %s\n", path)
			err = liveAnalyzer.StartWatching(path)
			if err != nil {
				log.Printf("Warning: Failed to watch %s: %v", path, err)
			}
		}
	}

	// Demonstrate AI agent integration
	fmt.Println("\n=== AI Agent Integration Demo ===")

	// Simulate AI agent creating a new file
	fmt.Println("\n1. Simulating AI agent creating a new Go file...")
	testFilePath := "./examples/ai_generated_code.go"
	createTestFile(testFilePath)

	// Wait for the file to be processed
	time.Sleep(1 * time.Second)

	// Simulate AI agent modifying the file
	fmt.Println("\n2. Simulating AI agent modifying the file...")
	modifyTestFile(testFilePath)

	// Wait for the modification to be processed
	time.Sleep(1 * time.Second)

	// Manually trigger an update (this is what the AI agent would call)
	fmt.Println("\n3. AI agent manually triggering file analysis...")
	err = liveAnalyzer.UpdateFile(testFilePath)
	if err != nil {
		log.Printf("Error manually updating file: %v", err)
	}

	// Show current stats
	fmt.Println("\n4. Current live analyzer statistics:")
	stats := liveAnalyzer.GetStats()
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Show file states
	fmt.Println("\n5. Tracked file states:")
	fileStates := liveAnalyzer.GetAllFileStates()
	for filePath, state := range fileStates {
		fmt.Printf("  %s: %d entities, analyzed: %v\n",
			filePath, len(state.Entities), state.Analyzed)
	}

	// Set up graceful shutdown
	fmt.Println("\n=== Live Analysis Running ===")
	fmt.Println("The analyzer is now watching for file changes...")
	fmt.Println("Try editing files in ./examples or ./internal directories")
	fmt.Println("Press Ctrl+C to stop")

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case <-interrupt:
		fmt.Println("\nReceived interrupt signal, shutting down...")
	case <-time.After(30 * time.Second):
		fmt.Println("\nDemo completed after 30 seconds")
	}

	// Clean up test file
	os.Remove(testFilePath)

	fmt.Println("Live analysis demo completed!")
}

// setupAIAgentCallbacks configures callbacks that an AI coding agent would use
func setupAIAgentCallbacks(liveAnalyzer *analyzer.LiveAnalyzer) {
	liveAnalyzer.SetCallbacks(
		// onFileChanged - called when a file is modified
		func(filePath string, changeType analyzer.FileChangeType) {
			fmt.Printf("🤖 AI Agent notified: File %s was %v\n", filePath, changeType)

			// This is where an AI agent could:
			// 1. Update its internal model of the codebase
			// 2. Trigger related analyses
			// 3. Suggest improvements or detect issues
			// 4. Update documentation or tests
		},

		// onGraphUpdated - called when the graph is updated
		func(stats *analyzer.UpdateStats) {
			fmt.Printf("📊 Graph updated: %d files, +%d entities, -%d entities (in %v)\n",
				stats.FilesUpdated, stats.EntitiesAdded, stats.EntitiesRemoved, stats.ProcessingTime)

			// This is where an AI agent could:
			// 1. Query the updated graph for new insights
			// 2. Detect architectural changes
			// 3. Update code recommendations
			// 4. Trigger cross-language analysis
		},

		// onError - called when errors occur
		func(err error) {
			fmt.Printf("❌ Live analysis error: %v\n", err)

			// This is where an AI agent could:
			// 1. Log errors for debugging
			// 2. Attempt recovery strategies
			// 3. Notify users of issues
		},
	)
}

// createTestFile simulates an AI agent creating a new Go file
func createTestFile(filePath string) {
	content := `package main

import "fmt"

// AIGeneratedFunction demonstrates AI-generated code
func AIGeneratedFunction() {
	fmt.Println("This function was created by an AI agent")
}

// DataProcessor is an AI-generated struct
type DataProcessor struct {
	Name    string
	Active  bool
}

// NewDataProcessor creates a new data processor
func NewDataProcessor(name string) *DataProcessor {
	return &DataProcessor{
		Name:   name,
		Active: true,
	}
}

// Process processes data (AI-generated method)
func (dp *DataProcessor) Process(data string) string {
	if !dp.Active {
		return ""
	}
	return fmt.Sprintf("Processed by %s: %s", dp.Name, data)
}
`

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	os.MkdirAll(dir, 0755)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Printf("Failed to create test file: %v", err)
	}
}

// modifyTestFile simulates an AI agent modifying an existing file
func modifyTestFile(filePath string) {
	additionalContent := `

// AIEnhancedFunction is an enhancement added by AI
func AIEnhancedFunction(input string) string {
	processor := NewDataProcessor("AI-Enhanced")
	return processor.Process(input)
}

// ConfigManager manages configuration (AI-added)
type ConfigManager struct {
	settings map[string]interface{}
}

// NewConfigManager creates a new config manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		settings: make(map[string]interface{}),
	}
}

// SetSetting sets a configuration setting
func (cm *ConfigManager) SetSetting(key string, value interface{}) {
	cm.settings[key] = value
}

// GetSetting gets a configuration setting
func (cm *ConfigManager) GetSetting(key string) interface{} {
	return cm.settings[key]
}
`

	// Read existing content
	existing, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read test file: %v", err)
		return
	}

	// Append new content
	newContent := string(existing) + additionalContent

	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		log.Printf("Failed to modify test file: %v", err)
	}
}

// aiAgentWorkflow simulates a complete AI agent workflow
func aiAgentWorkflow(ctx context.Context, liveAnalyzer *analyzer.LiveAnalyzer) {
	fmt.Println("\n=== AI Agent Workflow Simulation ===")

	// Step 1: AI agent analyzes current codebase state
	fmt.Println("1. AI Agent: Analyzing current codebase state...")
	stats := liveAnalyzer.GetStats()
	fmt.Printf("   Current state: %d files tracked, %d entities\n",
		stats["tracked_files"], stats["total_entities"])

	// Step 2: AI agent identifies files that need attention
	fmt.Println("2. AI Agent: Identifying files that need attention...")
	fileStates := liveAnalyzer.GetAllFileStates()
	for filePath, state := range fileStates {
		if len(state.Entities) > 10 {
			fmt.Printf("   High complexity file: %s (%d entities)\n",
				filePath, len(state.Entities))
		}
	}

	// Step 3: AI agent creates a new utility file
	fmt.Println("3. AI Agent: Creating new utility file...")
	utilityPath := "./examples/ai_utility.go"
	createUtilityFile(utilityPath)

	// Step 4: Wait for analysis and check results
	time.Sleep(500 * time.Millisecond)
	fmt.Println("4. AI Agent: Checking analysis results...")

	newState := liveAnalyzer.GetFileState(utilityPath)
	if newState != nil {
		fmt.Printf("   New file analyzed: %d entities found\n", len(newState.Entities))
	}

	// Step 5: AI agent makes cross-file connections
	fmt.Println("5. AI Agent: Making cross-file connections...")
	// In a real implementation, the AI would analyze relationships
	// and suggest imports, function calls, or refactoring opportunities

	// Clean up
	os.Remove(utilityPath)
}

// createUtilityFile creates a utility file that an AI might generate
func createUtilityFile(filePath string) {
	content := `package main

import (
	"strings"
	"time"
)

// StringUtils provides string manipulation utilities
type StringUtils struct{}

// NewStringUtils creates a new string utilities instance
func NewStringUtils() *StringUtils {
	return &StringUtils{}
}

// Capitalize capitalizes the first letter of a string
func (su *StringUtils) Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Reverse reverses a string
func (su *StringUtils) Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// TimeUtils provides time manipulation utilities
type TimeUtils struct{}

// NewTimeUtils creates a new time utilities instance
func NewTimeUtils() *TimeUtils {
	return &TimeUtils{}
}

// FormatTimestamp formats a timestamp for display
func (tu *TimeUtils) FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// IsWeekend checks if a time is on a weekend
func (tu *TimeUtils) IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}
`

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	os.MkdirAll(dir, 0755)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Printf("Failed to create utility file: %v", err)
	}
}
