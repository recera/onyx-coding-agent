# File Path Feature Documentation

## Overview

All nodes in the knowledge graph now include a `file_path` property that stores the complete file path where the entity is defined. This critical feature enables AI coders to easily locate and retrieve source files for any code entity.

## Database Schema Updates

All node types (except File nodes which already have a path) now include a `file_path` property:

### Node Properties

#### File Node
```cypher
CREATE NODE TABLE File(
    path STRING,         // Primary key - full file path
    name STRING,         // File name
    language STRING      // Programming language
)
```

#### Function Node
```cypher
CREATE NODE TABLE Function(
    id STRING,           // Primary key
    name STRING,         // Function name
    signature STRING,    // Function signature
    body STRING,         // Function body
    file_path STRING     // NEW: Path to containing file
)
```

#### Class Node
```cypher
CREATE NODE TABLE Class(
    id STRING,           // Primary key
    name STRING,         // Class name
    signature STRING,    // Class signature
    file_path STRING     // NEW: Path to containing file
)
```

#### Method Node
```cypher
CREATE NODE TABLE Method(
    id STRING,           // Primary key
    name STRING,         // Method name
    signature STRING,    // Method signature
    body STRING,         // Method body
    receiver_type STRING,// Go receiver type
    file_path STRING     // NEW: Path to containing file
)
```

#### Struct Node
```cypher
CREATE NODE TABLE Struct(
    id STRING,           // Primary key
    name STRING,         // Struct name
    type_definition STRING,
    file_path STRING     // NEW: Path to containing file
)
```

#### Interface Node
```cypher
CREATE NODE TABLE Interface(
    id STRING,           // Primary key
    name STRING,         // Interface name
    type_definition STRING,
    file_path STRING     // NEW: Path to containing file
)
```

#### Import Node
```cypher
CREATE NODE TABLE Import(
    id STRING,           // Primary key
    name STRING,         // Import name
    path STRING,         // Import path
    alias STRING,        // Import alias
    file_path STRING     // NEW: Path to containing file
)
```

#### Variable Node
```cypher
CREATE NODE TABLE Variable(
    id STRING,           // Primary key
    name STRING,         // Variable name
    type STRING,         // Variable type
    value STRING,        // Variable value
    file_path STRING     // NEW: Path to containing file
)
```

## Query Examples for AI Coders

### Find the file path for a specific function
```cypher
MATCH (f:Function {name: "calculateSum"})
RETURN f.file_path
```

### Get all entities from a specific file
```cypher
MATCH (n)
WHERE n.file_path = "src/main.go"
RETURN n.name, labels(n)[0] as type
```

### Find all files containing a specific class
```cypher
MATCH (c:Class {name: "UserController"})
RETURN DISTINCT c.file_path
```

### Get file paths for all methods of a class
```cypher
MATCH (c:Class {name: "Calculator"})-[:Contains]->(m:Method)
RETURN m.name, m.file_path
```

### Find cross-file dependencies
```cypher
MATCH (caller)-[:CALLS]->(callee)
WHERE caller.file_path <> callee.file_path
RETURN DISTINCT caller.file_path as calling_file, callee.file_path as called_file
```

## Implementation Details

### Entity Creation
All entities are created with their file path set from the analyzer's current file context:

```go
// Example from Go analyzer
entity := entities.NewEntity(id, name, entities.EntityTypeFunction, ga.currentFile.Path, node)
```

### Database Storage
The `StoreEntity` method in KuzuDB includes file_path in all CREATE queries:

```go
query = fmt.Sprintf(`CREATE (f:Function {id: "%s", name: "%s", signature: "%s", body: "%s", file_path: "%s"})`,
    entity.ID, safeName, safeSignature, safeBody, safeFilePath)
```

## Benefits for AI Coders

1. **Direct File Access**: AI agents can retrieve the exact file path for any code entity
2. **Context Understanding**: Know which file contains specific functions, classes, or variables
3. **Cross-File Analysis**: Easily identify dependencies between files
4. **Code Navigation**: Jump directly to source locations
5. **Refactoring Support**: Track which files need updates when modifying code
6. **Documentation Generation**: Include accurate file references in generated docs

## Migration Notes

For existing databases created before this feature:
- The schema needs to be recreated with the new file_path fields
- All entities need to be re-analyzed to populate file_path values
- Old databases without file_path will need full re-analysis

## Usage in AI Agent API

The AI Agent API can leverage file paths for enhanced functionality:

```go
// Get all entities from a specific file
entities := aiAPI.GetEntitiesFromFile("src/main.go")

// Find which files contain references to a function
files := aiAPI.GetFilesReferencingEntity("calculateSum")

// Analyze file-level metrics
metrics := aiAPI.AnalyzeFile("src/complex_module.go")
```

## Future Enhancements

1. **Relative vs Absolute Paths**: Currently stores paths as provided by analyzer, could normalize
2. **Path Aliases**: Support for module aliases and shortened paths
3. **File Movement Tracking**: Update paths when files are moved/renamed
4. **Path Validation**: Ensure paths remain valid across different environments