# Go Code Graph 

A sophisticated code analysis framework that builds knowledge graphs from source code repositories using KuzuDB. Designed for AI coding agents, code analysis tools, and architectural understanding.

## 🚀 Key Features

- **🔍 Multi-Language Support**: Go, Python, TypeScript with Tree-sitter parsing
- **📊 Knowledge Graph**: KuzuDB embedded database with Cypher queries  
- **🧪 Test Coverage Analysis**: Comprehensive test detection and coverage metrics
- **🔄 Live Analysis**: Real-time file watching and incremental updates
- **🤖 AI Integration**: Built for AI coding agents with metrics and events
- **📍 File Tracking**: Every entity includes its source file location
- **🎯 Pattern Detection**: Identify architectural patterns and anti-patterns
- **⚡ High Performance**: Incremental analysis, debounced processing

## 📚 Documentation

For complete documentation, see [DOCUMENTATION.md](DOCUMENTATION.md)

### Quick Links
- [Installation & Setup](DOCUMENTATION.md#installation)
- [Quick Start Guide](DOCUMENTATION.md#quick-start)
- [Test Coverage Analysis](DOCUMENTATION.md#test-coverage-analysis)
- [API Reference](DOCUMENTATION.md#api-reference)
- [Query Examples](DOCUMENTATION.md#query-examples)
- [AI Agent Integration](DOCUMENTATION.md#ai-agent-integration)
- [Live Analysis](DOCUMENTATION.md#live-analysis)

## 📦 Quick Start

```bash
go get github.com/your-username/graph_service
```

```go
package main

import (
    "fmt"
    "log"
    codeanalyzer "graph_service"
)

func main() {
    // Analyze a repository
    result, err := codeanalyzer.BuildGraph(codeanalyzer.BuildGraphOptions{
        RepoPath:    "./my-project",
        DBPath:      "./analysis.db", 
        LoadEnvFile: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer result.Close()

    // Query the graph
    fmt.Println("Functions in main.go:")
    functions, _ := result.QueryGraph(`
        MATCH (f:Function) 
        WHERE f.file_path ENDS WITH 'main.go'
        RETURN f.name, f.signature
    `)
    fmt.Println(functions)
    
    // AI Agent: Get file for entity
    filePath, _ := result.GetEntityFilePath("HandleRequest")
    fmt.Printf("HandleRequest is in: %s\n", filePath)
    
    // Test Coverage: Get coverage metrics
    metrics, _ := result.GetCoverageMetrics()
    fmt.Printf("Test Coverage: %.1f%% (%d tests)\n", 
        metrics.CoveragePercentage, metrics.TestCount)
}
```

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    AI Coding Agent                      │
│                 (Uses file paths to                     │
│                  retrieve source)                       │
└────────────────────┬────────────────────────────────────┘
                     │ 
┌────────────────────▼───────────────────────────────────┐
│                 AI Agent API                           │
│  • Code metrics  • Events  • Recommendations           │
└────────────────────┬───────────────────────────────────┘
                     │
┌────────────────────▼───────────────────────────────────┐
│               Live Analyzer                            │
│  • File watching  • Incremental updates                │
└────────────────────┬───────────────────────────────────┘
                     │
┌────────────────────▼───────────────────────────────────┐
│            Language Analyzers                          │
│  • Go  • Python  • TypeScript  • Tree-sitter           │
└────────────────────┬───────────────────────────────────┘
                     │
┌────────────────────▼───────────────────────────────────┐
│              KuzuDB Graph                              │
│  • Entities  • Relationships  • File paths             │
└────────────────────────────────────────────────────────┘
```

## 🎯 Perfect for AI Coders

Every entity in the graph includes its source file path:

```cypher
// Find file containing a function
MATCH (f:Function {name: "processData"})
RETURN f.file_path

// Get all entities from a file  
MATCH (n)
WHERE n.file_path = "src/handlers/api.go"
RETURN n.name, labels(n)[0] as type

// Find cross-file dependencies
MATCH (caller)-[:CALLS]->(callee)
WHERE caller.file_path <> callee.file_path
RETURN DISTINCT caller.file_path, callee.file_path
```

## 🧪 Test Coverage Example

Run the comprehensive test coverage analysis example:

```bash
cd cmd/test_coverage_example
go run main.go
```

This example demonstrates:
- 📊 Overall coverage metrics and statistics
- 🎯 Entity-specific coverage analysis with confidence scoring
- 🚨 Detection of uncovered entities
- 🔍 Advanced test coverage queries
- 📈 Coverage quality assessment

## 🚀 Advanced Features

- **Natural Language Queries**: Ask questions in plain English
- **Architectural Pattern Detection**: Find design patterns  
- **Code Quality Metrics**: Complexity, coupling, cohesion
- **Change Impact Analysis**: Understand ripple effects
- **Cross-Language Analysis**: Track polyglot dependencies

## 📈 Performance

- Analyze 1,000 files in ~5-10 seconds
- Incremental updates in ~50-200ms per file
- Sub-millisecond query performance
- Memory efficient for large codebases

## 🤝 Contributing

See [DOCUMENTATION.md](DOCUMENTATION.md#contributing) for contribution guidelines.

## 📄 License

MIT License

---

**Built with ❤️ for AI coding agents and developers**
