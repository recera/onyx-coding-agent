# Live Graph Analysis for AI Coding Agents

This document explains the live analysis and real-time graph update functionality designed specifically for AI coding agent integration.

## Overview

The live analysis system provides real-time file watching and incremental graph updates, allowing AI coding agents to:

1. **Monitor code changes in real-time** as they write or edit code
2. **Receive immediate feedback** on graph updates and code quality metrics
3. **Query the updated graph** for architectural insights and recommendations
4. **Track relationships** between files and entities across the codebase

## Architecture

### Core Components

1. **LiveAnalyzer** (`internal/analyzer/live_analyzer.go`)
   - Real-time file watching using `fsnotify`
   - Debounced change processing to handle rapid edits
   - Incremental graph updates without full re-analysis
   - Thread-safe file state tracking

2. **AIAgentAPI** (`internal/analyzer/ai_agent_api.go`)
   - High-level interface for AI coding agents
   - Code quality metrics and recommendations
   - Architectural pattern detection
   - Event-based notifications

3. **File State Management**
   - Tracks current state of each analyzed file
   - Maintains entity and relationship mappings
   - Provides diff-based updates

## Key Features

### Real-Time File Watching

```go
// Start watching directories for changes
err := liveAnalyzer.StartWatching("./src")

// Set up AI agent callbacks
liveAnalyzer.SetCallbacks(
    func(filePath string, changeType FileChangeType) {
        // Handle file changes
    },
    func(stats *UpdateStats) {
        // Handle graph updates
    },
    func(err error) {
        // Handle errors
    },
)
```

### Manual Updates for AI Agents

```go
// AI agent can manually trigger file analysis
err := liveAnalyzer.UpdateFile("path/to/file.go")
```

### Code Quality Analysis

```go
// Get code metrics for a file
metrics, err := aiAPI.AnalyzeCodeQuality("path/to/file.go")
// Returns: complexity, function count, coupling scores, etc.

// Get AI-generated recommendations
recommendations, err := aiAPI.GetRecommendations("path/to/file.go")
// Returns: refactoring suggestions, complexity warnings, etc.
```

### Architectural Insights

```go
// Detect architectural patterns
insights, err := aiAPI.DetectArchitecturalPatterns()
// Returns: Factory patterns, Interface segregation, Microservices, etc.

// Analyze change impact
impact, err := aiAPI.GetChangeImpact("path/to/file.go")
// Returns: affected files, risk level, recommendations
```

## Event-Driven Architecture

The system uses an event-driven approach where AI agents can subscribe to various events:

### Event Types

- **FileChanged**: File was added, modified, or deleted
- **GraphUpdated**: Graph entities/relationships were updated
- **EntityAdded/Removed**: Specific entities were added or removed
- **ComplexityChanged**: Code complexity metrics changed
- **ErrorOccurred**: Analysis errors occurred

### Event Consumption

```go
// Get event channel
events := aiAPI.GetEvents()

// Process events
for event := range events {
    switch event.Type {
    case analyzer.EventFileChanged:
        // Handle file change
    case analyzer.EventGraphUpdated:
        // Handle graph update
    }
}
```

## AI Agent Integration Patterns

### 1. Real-Time Code Analysis

```go
// As AI agent writes code, get immediate feedback
func onAICodeGeneration(filePath string, generatedCode string) {
    // Update the file
    ioutil.WriteFile(filePath, []byte(generatedCode), 0644)
    
    // Trigger analysis
    aiAPI.UpdateFileWithAIGeneration(filePath, generatedCode)
    
    // Get immediate feedback
    metrics, _ := aiAPI.AnalyzeCodeQuality(filePath)
    recommendations, _ := aiAPI.GetRecommendations(filePath)
}
```

### 2. Contextual Code Suggestions

```go
// AI agent analyzes context before making suggestions
func getContextualSuggestions(currentFile string) {
    // Find related files
    relatedFiles, _ := aiAPI.GetRelatedFiles(currentFile)
    
    // Analyze impact of potential changes
    impact, _ := aiAPI.GetChangeImpact(currentFile)
    
    // Get codebase overview for context
    overview := aiAPI.GetCodebaseOverview()
}
```

### 3. Architectural Guidance

```go
// AI agent provides architectural guidance
func provideArchitecturalGuidance() {
    // Detect current patterns
    patterns, _ := aiAPI.DetectArchitecturalPatterns()
    
    // Analyze complexity distribution
    overview := aiAPI.GetCodebaseOverview()
    complexity := overview["complexity_distribution"]
    
    // Generate architectural recommendations
}
```

## Performance Characteristics

### Debouncing
- **Default debounce interval**: 500ms (customizable to 200ms for AI agents)
- **Prevents analysis storms** during rapid file editing
- **Batches multiple changes** to the same file

### Incremental Updates
- **File-level granularity**: Only re-analyzes changed files
- **Entity tracking**: Maintains state of entities between updates
- **Relationship preservation**: Preserves cross-file relationships

### Memory Management
- **File state caching**: Keeps analyzed file states in memory
- **Configurable depth**: Limits directory watching depth
- **Ignore patterns**: Skips irrelevant files (.git, node_modules, etc.)

## Configuration Options

```go
watchOptions := &analyzer.WatchOptions{
    WatchedExtensions: []string{".go", ".py", ".js", ".ts"},
    IgnorePatterns:    []string{".git", "node_modules", "__pycache__"},
    DebounceInterval:  200 * time.Millisecond, // Fast for AI agents
    MaxDepth:          10,                     // Directory depth limit
    EnableCrossLang:   true,                   // Cross-language analysis
}
```

## Usage Example

```go
// Initialize database and analyzer
database, _ := db.NewKuzuDatabase("graph.db")
liveAnalyzer, _ := analyzer.NewLiveAnalyzer(database, watchOptions)
aiAPI := analyzer.NewAIAgentAPI(liveAnalyzer)

// Start watching
liveAnalyzer.StartWatching("./src")

// AI agent workflow
go func() {
    for event := range aiAPI.GetEvents() {
        switch event.Type {
        case analyzer.EventFileChanged:
            // Analyze the changed file
            metrics, _ := aiAPI.AnalyzeCodeQuality(event.FilePath)
            
            // Generate recommendations
            recommendations, _ := aiAPI.GetRecommendations(event.FilePath)
            
            // Update AI agent's understanding of codebase
            updateAIModel(event.FilePath, metrics, recommendations)
        }
    }
}()

// AI agent can manually update files
aiAPI.UpdateFileWithAIGeneration("new_file.go", generatedCode)
```

## Comparison with Original Repository

### Original Git-Based System
The original Python repository used a **git-based incremental update system**:

- **Backlog system**: Tracked all graph modification queries
- **Commit-based changes**: Processed file changes between git commits
- **Transition storage**: Stored queries to move between commit states
- **File deletion/recreation**: Removed and re-added entire files

### New Live Analysis System
Our Go implementation provides **real-time live updates**:

- **File watching**: Real-time file system monitoring
- **Debounced processing**: Handles rapid successive changes
- **Incremental analysis**: Updates only changed parts of the graph
- **AI-friendly API**: Designed specifically for AI agent integration

### Key Improvements for AI Agents

1. **Real-time feedback**: Immediate analysis without waiting for git commits
2. **Manual triggers**: AI agents can force analysis of specific files
3. **Rich events**: Detailed event system for AI agent notifications
4. **Quality metrics**: Built-in code quality analysis and recommendations
5. **Architectural insights**: Automatic pattern detection and suggestions

## Future Enhancements

1. **Language Server Protocol (LSP) integration**: For even more detailed analysis
2. **Conflict resolution**: Handle concurrent edits from multiple AI agents
3. **Undo/redo tracking**: Track and revert AI-generated changes
4. **Performance monitoring**: Track analysis performance and optimize
5. **Plugin system**: Allow custom analyzers for specific frameworks
6. **Distributed analysis**: Scale across multiple instances for large codebases

## Testing

Run the live analysis demo:

```bash
cd go-code-graph
go run ./cmd/test_live_analyzer/
```

The demo will:
1. Set up file watching on example directories
2. Simulate AI agent creating and modifying files
3. Show real-time graph updates and statistics
4. Demonstrate callback integration for AI agents

## Integration Checklist

For AI coding agents integrating with this system:

- [ ] Set up LiveAnalyzer with appropriate watch options
- [ ] Configure event callbacks for real-time notifications
- [ ] Implement file change handlers in AI agent logic
- [ ] Use manual update triggers when generating code
- [ ] Query code quality metrics for feedback
- [ ] Monitor architectural insights for guidance
- [ ] Handle errors and edge cases gracefully
- [ ] Test with rapid file changes and large codebases 