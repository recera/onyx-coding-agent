# Tonic Integration Guide for Graph Service

## Quick Start Migration (5 minutes)

### Step 1: Copy the graph_service directory
```bash
cp -r /path/to/goru/internal/graph_service /path/to/tonic/internal/
```

### Step 2: Update imports (automated)
```bash
cd /path/to/tonic/internal/graph_service
find . -name "*.go" -exec sed -i '' 's/goru\/internal\/graph_service/tonic\/internal\/graph_service/g' {} +
```

### Step 3: Update ignore patterns
```go
// In graph_builder.go line 102, change:
".goru" → ".tonic"
```

### Step 4: Add dependencies to tonic's go.mod
```go
require (
    github.com/kuzudb/go-kuzu v0.10.0
    github.com/tree-sitter/go-tree-sitter v0.25.0
    github.com/tree-sitter/tree-sitter-go v0.23.4
    github.com/tree-sitter/tree-sitter-python v0.23.6
    github.com/tree-sitter/tree-sitter-typescript v0.23.2
    github.com/fsnotify/fsnotify v1.9.0
    github.com/joho/godotenv v1.5.1
)
```

## Tonic Agent Integration

### Basic Setup
```go
package tonic

import (
    "tonic/internal/graph_service"
    "tonic/internal/graph_service/internal/analyzer"
    "tonic/internal/graph_service/internal/db"
)

type TonicAgent struct {
    database     *db.KuzuDatabase
    liveAnalyzer *analyzer.LiveAnalyzer
    aiAPI        *analyzer.AIAgentAPI
    graphPath    string
}

func NewTonicAgent(repoPath string, dbPath string) (*TonicAgent, error) {
    // Initialize database in separate folder
    database, err := db.NewKuzuDatabase(dbPath)
    if err != nil {
        return nil, err
    }
    
    // Create schema
    if err := database.CreateSchema(); err != nil {
        return nil, err
    }
    
    // Build initial graph
    result, err := graph_service.BuildGraph(graph_service.BuildGraphOptions{
        RepoPath:    repoPath,
        DBPath:      dbPath,
        CleanupDB:   false,
    })
    if err != nil {
        return nil, err
    }
    
    // Setup live analyzer for real-time updates
    liveAnalyzer, err := analyzer.NewLiveAnalyzer(database, analyzer.DefaultWatchOptions())
    if err != nil {
        return nil, err
    }
    
    // Create AI API
    aiAPI := analyzer.NewAIAgentAPI(liveAnalyzer)
    
    return &TonicAgent{
        database:     database,
        liveAnalyzer: liveAnalyzer,
        aiAPI:        aiAPI,
        graphPath:    dbPath,
    }, nil
}
```

### Cypher Query Tool for Agent
```go
// Tool for TypeScript agent to run Cypher queries
func (t *TonicAgent) RunCypherQuery(query string) (string, error) {
    result, err := t.database.ExecuteQuery(query)
    if err != nil {
        return "", err
    }
    return result, nil
}

// Example queries the agent can use
var AgentQueries = map[string]string{
    "find_function": `
        MATCH (f:Function {name: $name})
        RETURN f.file_path, f.line_number, f.signature
    `,
    "find_references": `
        MATCH (f:Function {name: $name})<-[:CALLS]-(caller)
        RETURN caller.name, caller.file_path, caller.line_number
    `,
    "get_class_methods": `
        MATCH (c:Class {name: $name})-[:CONTAINS]->(m:Method)
        RETURN m.name, m.signature, m.file_path, m.line_number
    `,
    "find_untested_functions": `
        MATCH (f:Function)
        WHERE NOT f.is_test AND NOT EXISTS((f)<-[:TESTS|COVERS]-(:TestFunction))
        RETURN f.name, f.file_path, f.line_number, f.complexity
        ORDER BY f.complexity DESC
        LIMIT 10
    `,
}
```

### File Reading Integration
```go
// Get file path from entity and read content
func (t *TonicAgent) ReadEntityFile(entityID string) (string, error) {
    query := `MATCH (e:Entity {id: $id}) RETURN e.file_path, e.line_number`
    result, err := t.database.ExecuteQuery(query)
    if err != nil {
        return "", err
    }
    
    // Parse result to get file_path
    // Then read file content
    content, err := os.ReadFile(filePath)
    return string(content), err
}
```

### TypeScript Agent Tool Interface
```typescript
// TypeScript side of tonic agent
interface GraphServiceTool {
    // Run arbitrary Cypher query
    runCypher(query: string): Promise<QueryResult>;
    
    // Get file content by entity ID
    readEntityFile(entityId: string): Promise<string>;
    
    // High-level queries
    findFunction(name: string): Promise<FunctionInfo>;
    findReferences(functionName: string): Promise<Reference[]>;
    getClassMethods(className: string): Promise<Method[]>;
    findUntestedCode(): Promise<UntestedFunction[]>;
    getTestCoverage(filePath?: string): Promise<CoverageReport>;
}

// Example usage in agent
async function analyzeCodebase(tool: GraphServiceTool) {
    // Find all untested complex functions
    const result = await tool.runCypher(`
        MATCH (f:Function)
        WHERE f.complexity > 10 AND NOT EXISTS((f)<-[:TESTS]-())
        RETURN f.name, f.file_path, f.complexity
    `);
    
    // Read the most complex function's file
    if (result.rows.length > 0) {
        const filePath = result.rows[0].file_path;
        const content = await tool.readEntityFile(result.rows[0].id);
        // Analyze content...
    }
}
```

## Key Features for Tonic

### 1. Real-time Graph Updates
```go
// Start watching for file changes
func (t *TonicAgent) StartWatching(path string) error {
    return t.liveAnalyzer.StartWatching(path)
}

// Set callbacks for updates
t.liveAnalyzer.SetCallbacks(
    func(filePath string, changeType analyzer.FileChangeType) {
        // Notify TypeScript agent of file change
    },
    func(stats *analyzer.UpdateStats) {
        // Update graph statistics
    },
    func(err error) {
        // Handle errors
    },
)
```

### 2. Code Quality Analysis
```go
// Get code quality metrics for agent decision making
metrics, _ := t.aiAPI.AnalyzeCodeQuality("main.go")
// Returns: complexity, maintainability, test coverage, etc.

// Detect architectural patterns
patterns, _ := t.aiAPI.DetectArchitecturalPatterns()
// Returns: MVC, microservices, layered architecture, etc.

// Analyze change impact
impact, _ := t.aiAPI.GetChangeImpact("critical_file.go")
// Returns: affected files, broken dependencies, test impacts
```

### 3. Test Coverage Queries
```go
// Coverage queries the agent can use
coverageQueries := map[string]string{
    "overall_coverage": `
        MATCH (f:Function) WHERE NOT f.is_test
        WITH COUNT(f) as total
        MATCH (t:TestFunction)-[:TESTS|COVERS]->(f:Function)
        WHERE NOT f.is_test
        RETURN ROUND(100.0 * COUNT(DISTINCT f) / total, 2) as coverage
    `,
    "uncovered_by_file": `
        MATCH (f:Function)
        WHERE f.file_path = $filePath
          AND NOT f.is_test
          AND NOT EXISTS((f)<-[:TESTS|COVERS]-(:TestFunction))
        RETURN f.name, f.line_number
    `,
}
```

## Database Location Strategy

```go
// Recommended structure for tonic
/*
project-root/
├── .tonic/
│   ├── graph.db/          # KuzuDB files
│   │   ├── catalog.kz
│   │   ├── data.kz
│   │   └── ...
│   └── analysis-cache/     # Optional caching
└── src/                    # User's code
*/

func GetTonicDBPath(projectRoot string) string {
    return filepath.Join(projectRoot, ".tonic", "graph.db")
}
```

## API Surface Summary

### Core Functions
- `BuildGraph()` - Initial graph construction
- `ExecuteQuery()` - Run Cypher queries
- `StartWatching()` - Live file monitoring
- `StopWatching()` - Stop monitoring
- `Close()` - Clean shutdown

### Entity Properties Available
All entities have these core properties accessible via Cypher:
- `id` - Unique identifier
- `name` - Entity name
- `file_path` - Full file path (for agent file reading)
- `line_number` - Line in file
- `type` - Entity type
- `language` - Programming language

### Relationship Types
The agent can query these relationships:
- `CONTAINS` - File/class containment
- `CALLS` - Function calls
- `IMPORTS` - Import relationships  
- `INHERITS` - Inheritance
- `IMPLEMENTS` - Interface implementation
- `TESTS` - Test coverage
- `COVERS` - Code coverage
- `TESTS_API` - API endpoint testing
- `TESTS_COMPONENT` - Component testing

## Performance Considerations

1. **Database Location**: Store `.tonic/graph.db` outside the watched directory to avoid recursion
2. **Ignore Patterns**: Add `.tonic` to file watcher ignore list
3. **Query Optimization**: Use indexed properties (name, file_path) in WHERE clauses
4. **Batch Updates**: Use transaction batching for multiple file changes
5. **Memory Management**: Close database connection when tonic exits

## Testing the Integration

```go
// Test script for tonic integration
func TestTonicIntegration(t *testing.T) {
    // Create test repo
    testRepo := "/tmp/test-repo"
    os.MkdirAll(testRepo, 0755)
    
    // Initialize tonic agent
    agent, err := NewTonicAgent(testRepo, "/tmp/tonic-test.db")
    assert.NoError(t, err)
    defer agent.Close()
    
    // Test Cypher query
    result, err := agent.RunCypherQuery("MATCH (n) RETURN COUNT(n)")
    assert.NoError(t, err)
    assert.NotEmpty(t, result)
    
    // Test file reading
    content, err := agent.ReadEntityFile("test-entity-id")
    assert.NoError(t, err)
}
```

## Troubleshooting

### Common Issues

1. **Import errors after migration**
   - Ensure all `goru/` references are replaced with `tonic/`
   - Run `go mod tidy` after adding dependencies

2. **Database initialization fails**
   - Check write permissions for database directory
   - Ensure KuzuDB native library is installed

3. **File watching not working**
   - Verify fsnotify supports your OS
   - Check ignore patterns aren't too broad

4. **Memory usage high**
   - Limit file watcher scope
   - Use query result limits
   - Close unused database connections

## Next Steps

1. Copy graph_service to tonic project
2. Run migration scripts (5 minutes)
3. Implement TonicAgent wrapper
4. Add TypeScript tool interface
5. Test with sample repository
6. Optimize queries for agent use cases

The graph_service is fully ready for integration with tonic. All entities have file_path properties, the API is clean and well-documented, and the system is designed for exactly this kind of AI agent integration.