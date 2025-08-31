# Go Code Graph - Complete Documentation

## Table of Contents

1. [Overview](#overview)
2. [Key Features](#key-features)
3. [Architecture](#architecture)
4. [Installation](#installation)
5. [Quick Start](#quick-start)
6. [Core Concepts](#core-concepts)
7. [Test Coverage Analysis](#test-coverage-analysis)
8. [Usage Guide](#usage-guide)
9. [Chat Agent](#chat-agent)
10. [API Reference](#api-reference)
11. [Language Support](#language-support)
12. [Live Analysis](#live-analysis)
13. [AI Agent Integration](#ai-agent-integration)
14. [Database Schema](#database-schema)
15. [Query Examples](#query-examples)
16. [Advanced Features](#advanced-features)
17. [Performance](#performance)
18. [Troubleshooting](#troubleshooting)
19. [Contributing](#contributing)

## Overview

Go Code Graph is a sophisticated code analysis tool that builds a knowledge graph representation of source code repositories using KuzuDB as an embedded graph database. It provides powerful features for understanding code structure, dependencies, and relationships across multiple programming languages.

### What It Does

- **Analyzes** source code repositories to extract entities (functions, classes, methods, variables)
- **Builds** a queryable knowledge graph with relationships (calls, inheritance, imports)
- **Provides** real-time file watching and incremental updates
- **Enables** AI agents to understand and navigate codebases
- **Supports** multiple programming languages (Go, Python, TypeScript)

### Use Cases

- **AI Coding Assistants**: Provide context and code understanding for AI agents
- **Code Analysis**: Understand complex codebases and their dependencies
- **Refactoring**: Identify impact of changes across files
- **Documentation**: Generate insights about code structure
- **Architecture Review**: Detect patterns and architectural issues

## Key Features

### 🔍 Multi-Language Support
- **Go**: Full support for packages, functions, methods, structs, interfaces
- **Python**: Classes, functions, methods, imports, inheritance
- **TypeScript**: Classes, functions, interfaces, types, modules

### 📊 Knowledge Graph
- **KuzuDB**: High-performance embedded graph database
- **Cypher Queries**: Standard graph query language
- **Rich Schema**: Comprehensive entity and relationship types

### 🧪 Test Coverage Analysis
- **Comprehensive Test Detection**: Automatically identify test functions across multiple frameworks
- **Coverage Metrics**: Calculate coverage percentages, test counts, and quality scores
- **Multi-Framework Support**: Go testing, pytest, unittest, nose, hypothesis
- **Test Relationships**: Track which tests cover which production code
- **Coverage Quality**: Assess test quality with assertion counts and confidence scoring

### 🔄 Live Analysis
- **Real-time Monitoring**: File system watching with fsnotify
- **Incremental Updates**: Only analyze changed files
- **Debounced Processing**: Handle rapid file changes efficiently

### 🤖 AI Integration
- **Event-driven API**: Real-time notifications for AI agents
- **Code Metrics**: Complexity, coupling, and quality measurements
- **Pattern Detection**: Identify architectural patterns
- **Natural Language Queries**: LLM integration for queries

### 📍 File Path Tracking
- **Complete File Paths**: Every entity includes its source file location
- **Cross-file Analysis**: Track dependencies between files
- **Navigation Support**: Direct access to source locations

## Architecture

```
go-code-graph/
├── cmd/                    # Command-line tools and examples
├── internal/
│   ├── analyzer/          # Language analyzers and core analysis
│   │   ├── go_analyzer.go
│   │   ├── python_analyzer.go
│   │   ├── typescript_analyzer.go
│   │   ├── live_analyzer.go
│   │   └── ai_agent_api.go
│   ├── db/                # Database integration
│   │   ├── kuzudb.go
│   │   └── schema.go
│   ├── entities/          # Core data structures
│   │   ├── entity.go
│   │   ├── file.go
│   │   └── relationship.go
│   ├── git/              # Git integration
│   └── llm/              # LLM integration
├── examples/             # Usage examples
└── graph_builder.go      # Main API entry point
```

## Installation

### Prerequisites

- Go 1.23.0 or later
- Git (for repository cloning)
- C++ compiler (for KuzuDB)

### Install

```bash
go get github.com/your-username/go-code-graph
```

### Dependencies

The project uses these key dependencies:
- `github.com/kuzudb/go-kuzu` - Graph database
- `github.com/tree-sitter/go-tree-sitter` - Code parsing
- `github.com/fsnotify/fsnotify` - File watching
- `github.com/sashabaranov/go-openai` - LLM integration

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    codeanalyzer "go-code-graph"
)

func main() {
    // Analyze a repository
    result, err := codeanalyzer.BuildGraph(codeanalyzer.BuildGraphOptions{
        RepoPath:    "/path/to/your/repository",
        DBPath:      "./code-graph.db",
        CleanupDB:   false,  // Keep database for reuse
        LoadEnvFile: true,   // Load .env for API keys
    })
    if err != nil {
        log.Fatal(err)
    }
    defer result.Close()

    // Print statistics
    fmt.Printf("Analyzed: %d files, %d functions, %d classes\n",
        result.Stats.FilesCount,
        result.Stats.FunctionsCount,
        result.Stats.ClassesCount)

    // Query the graph
    classes, err := result.QueryGraph("MATCH (c:Class) RETURN c.name, c.file_path")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Classes found:", classes)
}
```

### Analyze Remote Repository

```go
result, err := codeanalyzer.BuildGraph(codeanalyzer.BuildGraphOptions{
    RepoURL:     "https://github.com/user/repo.git",
    CleanupDB:   true,  // Clean up after analysis
    LoadEnvFile: true,
})
```

## Core Concepts

### Entities

Entities represent code constructs:

- **Function**: Standalone functions
- **Class**: Class definitions
- **Method**: Class or struct methods
- **Variable**: Variable declarations
- **Import**: Import statements
- **Struct**: Go structs
- **Interface**: Interface definitions
- **Type**: Type aliases
- **Enum**: Enumeration types

### Relationships

Relationships connect entities:

- **Contains**: File contains entity, class contains method
- **Calls**: Function/method calls another
- **Inherits**: Class inheritance
- **Imports**: File imports
- **Implements**: Struct implements interface
- **Embeds**: Struct embedding
- **Uses**: Entity uses type/interface
- **Defines**: Struct/interface defines method

### File Tracking

Every entity includes a `file_path` property for source location.

## Test Coverage Analysis

The system provides comprehensive test coverage analysis capabilities, automatically detecting test functions and calculating coverage metrics across your codebase.

### Test Entity Types

The following test-specific entity types are supported:

- **TestFunction**: Individual test functions/methods
- **TestCase**: Specific test cases within a test function
- **TestSuite**: Test suites/classes containing multiple tests
- **Assertion**: Individual assertions within tests
- **Mock**: Mock objects/functions used in tests
- **Fixture**: Test fixtures and test data

### Test Detection

#### Go Test Detection
- **Function Patterns**: Functions starting with `Test` (e.g., `TestUserLogin`)
- **File Patterns**: Files ending with `_test.go`
- **Benchmark Tests**: Functions starting with `Benchmark`
- **Example Tests**: Functions starting with `Example`
- **Fuzz Tests**: Functions starting with `Fuzz`

#### Python Test Detection
- **pytest**: Functions starting with `test_` (e.g., `test_user_login`)
- **unittest**: Methods in classes inheriting from `TestCase`
- **Decorators**: Functions with test decorators (`@pytest.mark`, `@unittest`, etc.)
- **File Patterns**: Files matching `test_*.py`, `*_test.py`, or containing "test"

#### Test Framework Support
- **Go**: Built-in testing framework, testify, ginkgo
- **Python**: pytest, unittest, nose, hypothesis
- **TypeScript**: Jest, Mocha, Jasmine (planned)

### Coverage Metrics

#### Overall Coverage Metrics
```go
metrics, err := result.GetCoverageMetrics()
// Returns:
// - TotalEntities: Total number of entities in codebase
// - TestedEntities: Number of entities with test coverage
// - UntestedEntities: Number of entities without tests
// - CoveragePercentage: Overall coverage percentage
// - TestCount: Total number of test functions
// - TotalAssertions: Total number of assertions
// - TestFrameworks: Breakdown by framework
// - TestTypes: Breakdown by test type (unit, integration, etc.)
// - CoverageByType: Coverage percentage by entity type
```

#### Entity-Specific Coverage
```go
coverage, err := result.GetTestCoverage(entityID)
// Returns:
// - IsCovered: Whether the entity has test coverage
// - CoverageScore: Confidence score (0.0-1.0)
// - TestCount: Number of tests covering this entity
// - AssertionCount: Number of assertions in covering tests
// - DirectTests: Tests directly testing this entity
// - IndirectTests: Tests indirectly covering this entity
// - CoverageMetrics: Additional metrics and properties
```

### Test Relationships

The system tracks various types of test relationships:

- **TESTS**: Test function tests a production function/class/method
- **COVERS**: Test covers execution of production code
- **MOCKS**: Test mocks a dependency or external service
- **USES_FIXTURE**: Test uses a test fixture or shared test data
- **SETUP_FOR**: Test setup function prepares for another test
- **TEARDOWN_FOR**: Test cleanup function cleans up after another test
- **ASSERTS_ON**: Test assertion checks a specific condition/value
- **GROUPS_TESTS**: Test suite groups multiple related tests
- **PARAMETERIZES**: Test is parameterized with multiple input sets
- **SKIPS**: Test is conditionally skipped
- **EXPECTS_FAILURE**: Test expects a specific failure condition
- **BENCHMARKS**: Test benchmarks performance of code
- **DOCUMENTS**: Test serves as executable documentation/example
- **VERIFIES_BEHAVIOR**: Test verifies specific behavior or contract

### Coverage API

#### Get Coverage Metrics
```go
// Get overall coverage statistics
metrics, err := result.GetCoverageMetrics()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Coverage: %.1f%% (%d/%d entities)\n", 
    metrics.CoveragePercentage, 
    metrics.TestedEntities, 
    metrics.TotalEntities)
```

#### Get Entity Coverage
```go
// Get coverage for specific entity
coverage, err := result.GetTestCoverage("entity-id")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Entity covered: %v (score: %.2f)\n", 
    coverage.IsCovered, coverage.CoverageScore)
```

#### Find Uncovered Entities
```go
// Find all entities without test coverage
uncovered, err := result.GetUncoveredEntities()
if err != nil {
    log.Fatal(err)
}

for _, entity := range uncovered {
    fmt.Printf("Uncovered: %s in %s\n", entity.Name, entity.FilePath)
}
```

#### Get Tests for Entity
```go
// Get all tests covering a specific entity
tests, err := result.GetTestsByTarget("entity-id")
if err != nil {
    log.Fatal(err)
}

for _, test := range tests {
    fmt.Printf("Test: %s (%s)\n", test.Name, test.GetTestFramework())
}
```

### Coverage Quality Scoring

The system calculates coverage quality scores based on multiple factors:

- **Direct vs Indirect Coverage**: Direct tests (score: 0.8-1.0), indirect tests (score: 0.3-0.7)
- **Assertion Count**: More assertions indicate more thorough testing
- **Test Framework**: Some frameworks provide higher confidence
- **Test Type**: Unit tests vs integration tests vs benchmarks
- **Mock Usage**: Tests using mocks have adjusted confidence scores

### Test Coverage Queries

#### Find All Test Functions
```cypher
MATCH (test:TestFunction)
RETURN test.name, test.test_type, test.test_framework, test.file_path
ORDER BY test.file_path, test.name
```

#### Find Coverage Relationships
```cypher
MATCH (test:TestFunction)-[:TESTS]->(target)
RETURN test.name, target.name, target.file_path
```

#### Find Uncovered Functions
```cypher
MATCH (f:Function)
WHERE NOT EXISTS((test:TestFunction)-[:TESTS]->(f))
AND NOT f.file_path CONTAINS '_test'
RETURN f.name, f.file_path
```

#### Test Statistics by Framework
```cypher
MATCH (test:TestFunction)
RETURN test.test_framework, count(test) as test_count
ORDER BY test_count DESC
```

#### Coverage by File
```cypher
MATCH (f:Function)
OPTIONAL MATCH (test:TestFunction)-[:TESTS]->(f)
WITH f.file_path as file, 
     count(DISTINCT f) as total_functions,
     count(DISTINCT test) as covered_functions
RETURN file, 
       total_functions, 
       covered_functions,
       (covered_functions * 100.0 / total_functions) as coverage_percent
ORDER BY coverage_percent DESC
```

### Example Usage

#### Basic Coverage Analysis
```go
package main

import (
    "fmt"
    "log"
    codeanalyzer "go-code-graph"
)

func main() {
    // Analyze repository with test coverage
    result, err := codeanalyzer.BuildGraph(codeanalyzer.BuildGraphOptions{
        RepoPath:    "./my-project",
        DBPath:      "./coverage-analysis.db",
        LoadEnvFile: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer result.Close()

    // Get overall coverage metrics
    metrics, err := result.GetCoverageMetrics()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("📊 Coverage Summary:\n")
    fmt.Printf("   Coverage: %.1f%%\n", metrics.CoveragePercentage)
    fmt.Printf("   Tests: %d\n", metrics.TestCount)
    fmt.Printf("   Assertions: %d\n", metrics.TotalAssertions)

    // Find uncovered entities
    uncovered, err := result.GetUncoveredEntities()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n🚨 Uncovered Entities (%d):\n", len(uncovered))
    for _, entity := range uncovered {
        fmt.Printf("   %s (%s) in %s\n", 
            entity.Name, entity.Type, entity.FilePath)
    }

    // Analyze specific entity coverage
    entities := result.GetEntitiesByName("MyFunction")
    if len(entities) > 0 {
        coverage, err := result.GetTestCoverage(entities[0].ID)
        if err == nil {
            fmt.Printf("\n🎯 MyFunction Coverage:\n")
            fmt.Printf("   Covered: %v\n", coverage.IsCovered)
            fmt.Printf("   Score: %.2f\n", coverage.CoverageScore)
            fmt.Printf("   Tests: %d\n", coverage.TestCount)
        }
    }
}
```

## Usage Guide

### Finding Entities

```go
// By name
entities := result.GetEntityByName("calculateSum")

// By file
fileEntities, err := result.GetFileEntities("src/main.go")

// Get file path for entity
filePath, err := result.GetEntityFilePath("UserController")
```

### Querying the Graph

```go
// Cypher query
result, err := result.QueryGraph(`
    MATCH (f:Function)-[:CALLS]->(g:Function)
    WHERE f.file_path <> g.file_path
    RETURN f.name, f.file_path, g.name, g.file_path
`)

// Natural language query (requires OPENAI_API_KEY)
result, err := result.QueryGraph("Show me all classes and their methods")
```

### Analyzing Relationships

```go
// Get related files
relatedFiles, err := result.GetRelatedFiles("src/models/user.go")

// Get cross-file references
refs, err := result.GetCrossFileReferences("src/api/handler.go")
for relType, files := range refs {
    fmt.Printf("%s: %v\n", relType, files)
}
```

## Chat Agent

The Go Code Graph system includes a sophisticated chat agent implementation (`cmd/chat_agent/main.go`) that allows users to have conversational interactions with their codebase through a knowledge graph interface.

### Overview

The chat agent provides:
- **Natural conversation** with your codebase using LLM integration
- **Dynamic file reading** with automatic path resolution
- **Cypher query execution** for precise code graph queries
- **Context-aware responses** that understand code structure and relationships
- **Multi-step reasoning** with query results and file content integration

### Key Features

#### 🧠 Conversational Interface
- Ask questions in natural language about your code
- Get insights about code structure, functions, classes, and relationships
- Understand dependencies and architectural patterns

#### 🔍 Dynamic Code Analysis
- The agent can read specific files on demand based on conversation context
- Automatically resolves relative file paths from the knowledge graph to absolute paths
- Provides detailed code explanations and analysis

#### 📊 Graph Query Integration
- Execute Cypher queries against the code knowledge graph
- Get structured data about code entities and relationships
- Combine graph data with file content for comprehensive analysis

### Architecture

```go
type ChatSession struct {
    client       llm.LLMClient      // LLM integration (Cerebras)
    graphDB      *graph.BuildGraphResult  // Code knowledge graph
    context      []llm.Message      // Conversation history
    systemPrompt string             // Agent instructions
    repoPath     string             // Repository root path
}

type ChatResponse struct {
    MessageToUser string `json:"message_to_user"`  // Response to display
    Cypher        string `json:"cypher"`           // Optional graph query
    ReadFile      string `json:"read_file"`        // Optional file to read
}
```

### Usage

#### Starting the Chat Agent

```bash
# Use current directory as repository
./chat_agent

# Specify repository path
./chat_agent /path/to/your/repository
```

#### Example Conversations

**Basic Code Inquiry:**
```
You: What functions are in the main package?
AI: I'll search for functions in the main package for you.
    [Executes: MATCH (f:Function) WHERE f.file_path CONTAINS 'main' RETURN f.name, f.file_path]
    
    Found 3 functions in main package:
    - main() in cmd/main.go
    - initConfig() in cmd/main.go  
    - setupLogging() in cmd/main.go
```

**File Analysis:**
```
You: Show me the implementation of the NewChatSession function
AI: Let me read the implementation of NewChatSession for you.
    [Reads: cmd/chat_agent/main.go]
    
    The NewChatSession function creates a new chat session with the codebase graph...
    [Provides detailed explanation of the function]
```

**Cross-File Dependencies:**
```
You: What files depend on the graph_service package?
AI: [Executes query to find import relationships]
    
    These files import the graph_service package:
    - cmd/chat_agent/main.go
    - examples/basic_usage/main.go
    - cmd/test/main.go
```

### Path Resolution System

One of the key innovations in the chat agent is its **automatic path resolution system**:

#### Problem Solved
- **KuzuDB stores relative paths** (e.g., `cmd/chat_agent/main.go`) for portability
- **File reading tools need absolute paths** (e.g., `/Users/user/repo/cmd/chat_agent/main.go`)
- **AI agents work with relative paths** but need to read actual files

#### Solution Implementation
```go
func (cs *ChatSession) readFileContent(filePathOrEntity string) (string, error) {
    var filePath string
    if filepath.IsAbs(filePathOrEntity) {
        // Already absolute - use as-is
        filePath = filePathOrEntity
    } else if strings.Contains(filePathOrEntity, "/") {
        // Relative path from KuzuDB - convert to absolute
        filePath = filepath.Join(cs.repoPath, filePathOrEntity)
    } else {
        // Entity name - lookup file path, then convert to absolute
        entityPath, err := cs.graphDB.GetEntityFilePath(filePathOrEntity)
        if err != nil {
            filePath = filepath.Join(cs.repoPath, filePathOrEntity)
        } else {
            filePath = filepath.Join(cs.repoPath, entityPath)
        }
    }
    // ... file reading logic
}
```

This system handles three cases:
1. **Absolute paths**: Used directly
2. **Relative paths with "/"**: Converted using repository root
3. **Entity names**: Looked up in graph, then converted to absolute

### System Prompt Design

The chat agent uses a comprehensive system prompt that provides:

#### Codebase Context
- Statistics about functions, classes, files, and relationships
- Database schema explanation with entity and relationship types
- Available query patterns and examples

#### Response Format
- **Structured JSON responses** with message, optional Cypher query, and optional file to read
- **Multi-step processing** where queries and file reads add context for final response
- **Error handling** that gracefully handles failed queries or file reads

#### Example System Prompt Elements
```
CODEBASE STATS:
- Functions: 150
- Classes: 45
- Files: 78
- Calls: 892

You can access detailed information through Cypher queries:
- Find functions: "MATCH (f:Function) RETURN f.name, f.file_path"
- Find calls: "MATCH (f:Function)-[:CALLS]->(g:Function) RETURN f.name, g.name"

RESPONSE FORMAT:
{
  "message_to_user": "Your conversational response",
  "cypher": "Optional Cypher query to gather information",
  "read_file": "Optional file path or entity name to read"
}
```

### Configuration

#### LLM Provider Configuration
```go
// Currently configured for Cerebras
parts.Provider = "cerebras"
parts.Model = "qwen-3-235b-a22b-instruct-2507"
parts.MaxTokens = 1000
parts.Temperature = 0.3
```

#### Database Configuration
```go
graphResult, err := graph.BuildGraph(graph.BuildGraphOptions{
    RepoPath:    repoPath,
    DBPath:      "./chat_agent.db",  // Persistent database
    CleanupDB:   false,              // Keep for session reuse
    LoadEnvFile: true,               // Load API keys
})
```

### Current Status

✅ **Implemented:**
- Basic conversational interface
- File path resolution system
- Cypher query integration
- Context management
- Multi-step reasoning

🔄 **Recent Fixes:**
- **Path Resolution Bug**: Fixed issue where relative paths from KuzuDB couldn't be resolved to absolute paths for file reading
- **Repository Path Storage**: Added `repoPath` field to `ChatSession` struct for proper path conversion

🚀 **Ready for Use:**
- Run the chat agent to interact with your codebase
- Ask questions about code structure and relationships
- Get detailed file analysis and explanations

### Future Enhancements

Potential improvements:
- **Web interface** for better user experience
- **Code editing capabilities** with AI suggestions
- **Multi-repository support** for analyzing dependencies
- **Export functionality** for saving conversation insights
- **Plugin system** for custom analysis tools

## API Reference

### BuildGraphOptions

```go
type BuildGraphOptions struct {
    RepoPath    string  // Local repository path
    RepoURL     string  // Git repository URL
    DBPath      string  // Database storage path
    CleanupDB   bool    // Clean up database after use
    LoadEnvFile bool    // Load .env file
}
```

### BuildGraphResult

Key methods:

#### Core Methods
- `QueryGraph(query string) (string, error)` - Execute Cypher or natural language query
- `GetEntityByName(name string) []*Entity` - Find entities by name
- `GetFile(filePath string) *File` - Get file information
- `GetEntityFilePath(entityName string) (string, error)` - Get file path for entity
- `GetFileEntities(filePath string) ([]*Entity, error)` - Get all entities in file
- `GetRelatedFiles(filePath string) ([]string, error)` - Find related files
- `GetCrossFileReferences(filePath string) (map[string][]string, error)` - Get dependencies
- `Close()` - Clean up resources

#### Test Coverage Methods
- `GetCoverageMetrics() (*CoverageMetrics, error)` - Get overall coverage statistics
- `GetTestCoverage(entityID string) (*TestCoverageResult, error)` - Get coverage for specific entity
- `GetUncoveredEntities() ([]*Entity, error)` - Find entities without test coverage
- `GetTestsByTarget(entityID string) ([]*Entity, error)` - Get tests covering an entity
- `GetEntitiesByName(name string) []*Entity` - Find entities by name
- `GetAllEntities() []*Entity` - Get all entities in the graph

#### Test Coverage Types

```go
type CoverageMetrics struct {
    TotalEntities       int                 // Total entities analyzed
    TestedEntities      int                 // Entities with test coverage
    UntestedEntities    int                 // Entities without tests
    CoveragePercentage  float64             // Overall coverage percentage
    TestCount          int                 // Total number of test functions
    TotalAssertions    int                 // Total assertions across all tests
    TestFrameworks     map[string]int      // Test count by framework
    TestTypes          map[string]int      // Test count by type
    CoverageByType     map[string]float64  // Coverage percentage by entity type
    Details            map[string]interface{} // Additional metrics
}

type TestCoverageResult struct {
    EntityID         string                 // ID of the analyzed entity
    IsCovered        bool                   // Whether entity has test coverage
    CoverageScore    float64               // Coverage confidence score (0.0-1.0)
    TestCount        int                   // Number of tests covering this entity
    AssertionCount   int                   // Total assertions in covering tests
    DirectTests      []*TestInfo           // Tests directly testing this entity
    IndirectTests    []*TestInfo           // Tests indirectly covering this entity
    CoverageMetrics  map[string]interface{} // Additional coverage data
}

type TestInfo struct {
    TestID           string    // Unique test identifier
    TestName         string    // Name of the test function
    TestFramework    string    // Testing framework used
    TestType         string    // Type of test (unit, integration, etc.)
    CoverageType     string    // Type of coverage (direct, indirect)
    ConfidenceScore  float64   // Confidence in this test relationship
    FilePath         string    // Path to test file
    AssertionCount   int       // Number of assertions in this test
}
```

## Language Support

### Go Language Features

- **Packages**: Package declarations and organization
- **Functions**: Function definitions with parameters and returns
- **Methods**: Receiver methods for structs
- **Structs**: Type definitions with fields
- **Interfaces**: Interface definitions with method signatures
- **Imports**: Import statements with aliases
- **Variables**: Package and local variables
- **Type Definitions**: Custom type definitions

### Python Language Features

- **Classes**: Class definitions with inheritance
- **Functions**: Function and method definitions
- **Methods**: Instance and class methods
- **Imports**: Import and from-import statements
- **Variables**: Global and class variables
- **Decorators**: Function and class decorators
- **Docstrings**: Documentation extraction

### TypeScript Language Features

- **Classes**: Class definitions with generics
- **Functions**: Function declarations and expressions
- **Interfaces**: Interface definitions
- **Types**: Type aliases and unions
- **Enums**: Enumeration types
- **Modules**: Module declarations
- **Imports/Exports**: ES6 module system
- **Generics**: Generic type parameters

## Live Analysis

### Setting Up Live Analysis

```go
// Create live analyzer
liveAnalyzer, err := analyzer.NewLiveAnalyzer(database, analyzer.DefaultWatchOptions())
if err != nil {
    log.Fatal(err)
}

// Set callbacks
liveAnalyzer.SetCallbacks(
    func(filePath string, changeType analyzer.FileChangeType) {
        fmt.Printf("File changed: %s (%v)\n", filePath, changeType)
    },
    func(stats *analyzer.UpdateStats) {
        fmt.Printf("Graph updated: %d entities added\n", stats.EntitiesAdded)
    },
    func(err error) {
        log.Printf("Error: %v\n", err)
    },
)

// Start watching
err = liveAnalyzer.StartWatching("./src")
```

### Watch Options

```go
watchOptions := &analyzer.WatchOptions{
    WatchedExtensions: []string{".go", ".py", ".ts"},
    IgnorePatterns:    []string{".git", "node_modules"},
    DebounceInterval:  200 * time.Millisecond,
    MaxDepth:          10,
    EnableCrossLang:   true,
}
```

## AI Agent Integration

### AI Agent API

```go
// Create AI API
aiAPI := analyzer.NewAIAgentAPI(liveAnalyzer)

// Get code metrics
metrics, err := aiAPI.AnalyzeCodeQuality("main.go")
fmt.Printf("Complexity: %d\n", metrics.CyclomaticComplexity)

// Get recommendations
recommendations, err := aiAPI.GetRecommendations("complex_file.go")
for _, rec := range recommendations {
    fmt.Printf("%s: %s\n", rec.Type, rec.Description)
}

// Detect patterns
patterns, err := aiAPI.DetectArchitecturalPatterns()
for _, pattern := range patterns {
    fmt.Printf("Found: %s (confidence: %.2f)\n", 
        pattern.Pattern, pattern.Confidence)
}
```

### Event Handling

```go
// Monitor events
go func() {
    for event := range aiAPI.GetEvents() {
        switch event.Type {
        case analyzer.EventFileChanged:
            fmt.Printf("File changed: %s\n", event.FilePath)
        case analyzer.EventGraphUpdated:
            fmt.Printf("Graph updated at %v\n", event.Timestamp)
        case analyzer.EventComplexityChanged:
            fmt.Printf("Complexity changed in %s\n", event.FilePath)
        }
    }
}()
```

## Database Schema

### Node Types

| Node Type | Properties |
|-----------|------------|
| File | path, name, language |
| Function | id, name, signature, body, file_path |
| Class | id, name, signature, file_path |
| Method | id, name, signature, body, receiver_type, file_path |
| Struct | id, name, type_definition, file_path |
| Interface | id, name, type_definition, file_path |
| Import | id, name, path, alias, file_path |
| Variable | id, name, type, value, file_path |

#### Test Node Types

| Node Type | Properties |
|-----------|------------|
| TestFunction | id, name, signature, body, file_path, test_type, test_target, assertion_count, test_framework |
| TestCase | id, name, description, test_suite_id, test_data, expected_result, file_path |
| TestSuite | id, name, description, file_path, test_count, setup_method, teardown_method |
| Assertion | id, assertion_type, expected_value, actual_value, test_function_id, file_path |
| Mock | id, name, mock_type, target_entity, mock_framework, file_path |
| Fixture | id, name, fixture_type, setup_code, cleanup_code, file_path |

### Relationship Types

#### Core Relationships
| Relationship | From → To | Description |
|--------------|-----------|-------------|
| Contains | File → Entity | File contains entity |
| CALLS | Function → Function | Function calls |
| IMPORTS | File → File | File imports |
| INHERITS | Class → Class | Class inheritance |
| EMBEDS | Struct → Struct | Struct embedding |
| IMPLEMENTS | Struct → Interface | Interface implementation |
| DEFINES | Struct/Interface → Method | Method definition |
| USES | Function → Type | Type usage |

#### Test Relationships
| Relationship | From → To | Description |
|--------------|-----------|-------------|
| TESTS | TestFunction → Entity | Test function tests a production entity |
| COVERS | TestFunction → Entity | Test covers execution of production code |
| MOCKS | TestFunction → Entity | Test mocks a dependency |
| USES_FIXTURE | TestFunction → Fixture | Test uses a test fixture |
| SETUP_FOR | TestFunction → TestFunction | Test setup function prepares for another test |
| TEARDOWN_FOR | TestFunction → TestFunction | Test cleanup function cleans up after test |
| ASSERTS_ON | Assertion → Entity | Test assertion checks a condition |
| GROUPS_TESTS | TestSuite → TestFunction | Test suite groups multiple tests |
| PARAMETERIZES | TestFunction → TestCase | Test is parameterized with test cases |
| SKIPS | TestFunction → Entity | Test is conditionally skipped |
| EXPECTS_FAILURE | TestFunction → Entity | Test expects a specific failure |
| BENCHMARKS | TestFunction → Entity | Test benchmarks performance |
| DOCUMENTS | TestFunction → Entity | Test serves as executable documentation |
| VERIFIES_BEHAVIOR | TestFunction → Entity | Test verifies specific behavior |

## Query Examples

### Find Functions by Complexity

```cypher
MATCH (f:Function)
WHERE size(f.body) > 1000
RETURN f.name, f.file_path, size(f.body) as complexity
ORDER BY complexity DESC
```

### Find Circular Dependencies

```cypher
MATCH path = (a)-[:CALLS*]->(a)
RETURN path
LIMIT 10
```

### Find Unused Functions

```cypher
MATCH (f:Function)
WHERE NOT (()-[:CALLS]->(f))
AND f.name <> 'main'
RETURN f.name, f.file_path
```

### Cross-File Dependencies

```cypher
MATCH (caller)-[:CALLS]->(callee)
WHERE caller.file_path <> callee.file_path
RETURN DISTINCT 
    caller.file_path as from_file,
    callee.file_path as to_file,
    count(*) as call_count
ORDER BY call_count DESC
```

### Test Coverage Analysis

#### Find All Test Functions
```cypher
MATCH (test:TestFunction)
RETURN test.name, test.test_type, test.test_framework, test.file_path
ORDER BY test.file_path, test.name
```

#### Find Functions with Test Coverage
```cypher
MATCH (test:TestFunction)-[:TESTS]->(func:Function)
RETURN func.name, func.file_path, count(test) as test_count
ORDER BY test_count DESC
```

#### Find Uncovered Functions
```cypher
MATCH (f:Function)
WHERE NOT EXISTS((test:TestFunction)-[:TESTS]->(f))
AND NOT f.file_path CONTAINS '_test'
AND NOT f.file_path CONTAINS 'test_'
RETURN f.name, f.file_path
```

#### Coverage Statistics by Framework
```cypher
MATCH (test:TestFunction)
RETURN test.test_framework, count(test) as test_count,
       avg(test.assertion_count) as avg_assertions
ORDER BY test_count DESC
```

#### Find Tests with Most Assertions
```cypher
MATCH (test:TestFunction)
WHERE test.assertion_count > 0
RETURN test.name, test.assertion_count, test.file_path
ORDER BY test.assertion_count DESC
LIMIT 10
```

#### Coverage by File
```cypher
MATCH (f:Function)
OPTIONAL MATCH (test:TestFunction)-[:TESTS]->(f)
WITH f.file_path as file, 
     count(DISTINCT f) as total_functions,
     count(DISTINCT test) as covered_functions
WHERE total_functions > 0
RETURN file, 
       total_functions, 
       covered_functions,
       (covered_functions * 100.0 / total_functions) as coverage_percent
ORDER BY coverage_percent DESC
```

#### Find Mocked Dependencies
```cypher
MATCH (test:TestFunction)-[:MOCKS]->(entity)
RETURN test.name, entity.name, entity.file_path
ORDER BY test.name
```

#### Find Integration Tests
```cypher
MATCH (test:TestFunction)
WHERE test.test_type = "integration"
RETURN test.name, test.file_path, test.test_framework
```

## Advanced Features

### Natural Language Queries

Set up OpenAI integration:

```bash
export OPENAI_API_KEY=your_api_key
```

Use natural language:

```go
result, err := result.QueryGraph("Which functions are never called?")
result, err := result.QueryGraph("Show me the most complex classes")
result, err := result.QueryGraph("Find circular dependencies")
```

### Cross-Language Analysis

The system can analyze relationships across different languages:

```go
// Find Python functions called from Go
query := `
    MATCH (go:Function)-[:CALLS]->(py:Function)
    WHERE go.file_path ENDS WITH '.go'
    AND py.file_path ENDS WITH '.py'
    RETURN go.name, py.name
`
```

### Custom Entity Properties

Entities support custom properties:

```go
entity.SetProperty("complexity", 15)
entity.SetProperty("test_coverage", 0.85)
entity.SetProperty("last_modified", time.Now())
```

## Performance

### Optimization Tips

1. **Use temporary databases** for one-time analysis
2. **Enable cleanup** for large repositories
3. **Limit watch depth** for deep directory structures
4. **Increase debounce interval** for slower systems
5. **Use specific file extensions** in watch options

### Benchmarks

- Repository with 1,000 files: ~5-10 seconds
- Incremental update: ~50-200ms per file
- Query execution: ~1-10ms for simple queries
- Memory usage: ~100-500MB for large repositories

## Troubleshooting

### Common Issues

**Database Connection Errors**
```bash
# Ensure KuzuDB dependencies are installed
go mod tidy
```

**Memory Issues**
```go
// Process large repos with cleanup
opts := BuildGraphOptions{
    RepoPath:  "/large/repo",
    CleanupDB: true,
}
```

**Parse Errors**
- Check file encoding (UTF-8 required)
- Verify syntax validity
- Update tree-sitter grammars

**Missing Entities**
- Check supported file extensions
- Verify language analyzer support
- Check for parsing errors in logs

**Chat Agent Issues**

*Path Resolution Errors:*
```
Error: file not found: cmd/chat_agent/main.go
```
- **Solution**: Ensure the `repoPath` is correctly set in `ChatSession`
- **Check**: Verify file paths in KuzuDB are relative to repository root
- **Fix**: Update to use `filepath.Join(cs.repoPath, relativePath)`

*LLM Connection Issues:*
- **Check API keys**: Ensure Cerebras API key is set in environment
- **Verify model**: Confirm model name `qwen-3-235b-a22b-instruct-2507` is available
- **Network**: Check internet connectivity for LLM API calls

*Database Connection Failures:*
```
Error: failed to build graph: ...
```
- **Solution**: Ensure KuzuDB dependencies are properly installed
- **Check**: Verify database path `./chat_agent.db` is writable
- **Fix**: Use absolute path or ensure directory exists

### Debug Mode

Enable detailed logging:

```go
os.Setenv("DEBUG", "true")
```

## Contributing

### Adding Language Support

1. Create analyzer in `internal/analyzer/`
2. Implement `AnalyzeFile` method
3. Register in GraphBuilder
4. Add tests and examples

### Development Setup

```bash
# Clone repository
git clone https://github.com/your-username/go-code-graph
cd go-code-graph

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build ./cmd/...
```

### Code Style

- Follow Go conventions
- Add documentation for exported functions
- Include examples in comments
- Write comprehensive tests

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- [KuzuDB](https://kuzudb.com/) - Embedded graph database
- [Tree-sitter](https://tree-sitter.github.io/) - Incremental parsing
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system notifications
- [OpenAI](https://openai.com/) - Natural language processing