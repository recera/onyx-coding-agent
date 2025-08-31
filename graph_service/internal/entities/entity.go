// Package entities defines the core data structures for representing code entities
// and relationships in the knowledge graph.
//
// This package provides a sophisticated entity system that can represent various
// programming language constructs (functions, classes, methods, variables, etc.)
// and their relationships (calls, inheritance, containment, etc.). The system is
// designed to be language-agnostic while providing specialized support for
// specific language features.
//
// Core Concepts:
//
// Entity: Represents any identifiable code construct with metadata, position
// information, and hierarchical relationships. Examples include functions,
// classes, methods, variables, imports, and framework-specific constructs.
//
// Relationship: Represents connections between entities such as function calls,
// class inheritance, module imports, and structural containment.
//
// File: Represents a source file containing multiple entities with file-level
// metadata and statistics.
//
// The entity system supports:
//   - Multi-language constructs (Go, Python, TypeScript)
//   - Hierarchical relationships (classes contain methods)
//   - Cross-reference tracking (function calls, imports)
//   - Framework-specific entities (React components, API endpoints)
//   - Extensible property system for language-specific data
//
// Example usage:
//
//	// Create a function entity
//	funcEntity := NewEntity("func_123", "calculateSum", EntityTypeFunction, "main.go", node)
//	funcEntity.Signature = "func calculateSum(a, b int) int"
//	funcEntity.Body = "return a + b"
//
//	// Create a class entity and add the function as a method
//	classEntity := NewEntity("class_456", "Calculator", EntityTypeClass, "main.go", classNode)
//	classEntity.AddChild(funcEntity)
//
//	// Add symbol references
//	funcEntity.AddSymbol("calls", callNode)
//	funcEntity.SetProperty("complexity", 1)
//
// Thread Safety:
//   - Entity structures are not thread-safe for concurrent modification
//   - Read operations are safe across multiple goroutines
//   - Use synchronization when modifying entities from multiple goroutines
package entities

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

// Entity represents a code construct (function, class, method, variable, etc.)
// discovered during source code analysis. It captures both the structural
// information (location, hierarchy) and semantic information (symbols, relationships)
// of programming language elements.
//
// Each entity corresponds to a specific syntactic construct in the source code
// and maintains references to its Tree-sitter AST node for detailed analysis.
// Entities form a hierarchical structure where classes contain methods, modules
// contain functions, etc.
//
// Key Features:
//   - Unique identification for global reference
//   - Hierarchical parent-child relationships
//   - Symbol tracking for cross-references
//   - Extensible property system for language-specific data
//   - Position tracking for source code navigation
//
// Example entity types:
//   - Function: Standalone function or method
//   - Class: Class definition with methods and properties
//   - Variable: Variable declaration or assignment
//   - Import: Module or package import statement
//   - Interface: Type interface definition (Go, TypeScript)
//
// Usage patterns:
//
//	// Create and configure entity
//	entity := NewEntity("func_main_123", "main", EntityTypeFunction, "main.go", node)
//	entity.Signature = "func main()"
//	entity.SetProperty("visibility", "public")
//
//	// Track symbol references
//	entity.AddSymbol("calls", functionCallNode)
//	entity.AddSymbol("variables", variableNode)
//
//	// Establish hierarchy
//	classEntity.AddChild(methodEntity)
type Entity struct {
	// ID is a globally unique identifier for this entity across the entire
	// codebase analysis. It typically combines entity type, location, and
	// a hash to ensure uniqueness even with name collisions.
	//
	// Format examples:
	//   - "func_main.go_calculateSum_123"
	//   - "class_models.py_User_456"
	//   - "method_User_getName_789"
	ID string

	// Name is the identifier or name of the entity as it appears in source code.
	// For functions this is the function name, for classes the class name, etc.
	//
	// Examples: "calculateSum", "User", "getName", "PI"
	Name string

	// Type categorizes the entity according to its programming language construct.
	// This determines how the entity is processed, stored, and queried.
	Type EntityType

	// FilePath specifies the source file containing this entity, typically
	// as a relative path from the repository root. Used for navigation and
	// cross-file relationship analysis.
	//
	// Examples: "src/main.go", "models/user.py", "components/Header.tsx"
	FilePath string

	// Node holds a reference to the Tree-sitter AST node representing this
	// entity in the parsed syntax tree. Provides access to detailed syntactic
	// information and enables precise source code extraction.
	Node *ts.Node

	// StartByte indicates the byte offset where this entity begins in the
	// source file. Used for precise source code positioning and extraction.
	StartByte uint32

	// EndByte indicates the byte offset where this entity ends in the source
	// file. Combined with StartByte, defines the exact source code span.
	EndByte uint32

	// Signature contains the complete signature or declaration of the entity
	// as it appears in source code, including parameters, return types, etc.
	//
	// Examples:
	//   - "func calculateSum(a, b int) int"
	//   - "class User(BaseModel):"
	//   - "const API_URL: string"
	Signature string

	// Body contains the complete source code body of the entity, including
	// implementation details. For functions this includes the function body,
	// for classes it includes all methods and properties.
	Body string

	// DocString contains documentation comments or docstrings associated with
	// the entity. Extracted from language-specific documentation formats
	// (Go comments, Python docstrings, JSDoc, etc.).
	DocString string

	// Symbols maps symbol types to lists of Tree-sitter nodes representing
	// references within this entity. Used to track function calls, variable
	// references, type usage, etc.
	//
	// Common symbol types:
	//   - "calls": Function/method calls
	//   - "variables": Variable references
	//   - "types": Type references
	//   - "imports": Import statements
	//
	// Example: entity.Symbols["calls"] contains nodes for all function calls
	// made within this entity.
	Symbols map[string][]*ts.Node

	// Children contains entities that are structurally contained within this
	// entity. For example, a class entity contains method entities as children.
	// This creates a hierarchical tree structure of the codebase.
	Children []*Entity

	// Parent references the entity that contains this entity in the code
	// structure. For example, a method's parent is its containing class.
	// Forms the upward link in the hierarchical structure.
	Parent *Entity

	// Properties provides an extensible key-value store for language-specific
	// or analysis-specific metadata that doesn't fit in the standard fields.
	//
	// Common properties:
	//   - "visibility": "public", "private", "protected"
	//   - "static": true/false for static methods
	//   - "async": true/false for async functions
	//   - "complexity": cyclomatic complexity score
	//   - "parameters": detailed parameter information
	Properties map[string]interface{}
}

// EntityType categorizes code entities according to their programming language
// construct type. This enumeration supports multiple programming languages and
// frameworks, with extensibility for specialized constructs.
//
// The type system is organized in phases:
//   - Core types: Universal constructs (Function, Class, Variable)
//   - Language-specific: Go structs, Python classes, TypeScript interfaces
//   - Framework-specific: React components, API endpoints, middleware
//
// Each type determines:
//   - How the entity is parsed and analyzed
//   - What properties and relationships are relevant
//   - How it's stored in the graph database
//   - What queries and analyses apply
type EntityType string

const (
	EntityTypeFile      EntityType = "File"     // Source file entity
	EntityTypeFunction  EntityType = "Function"
	EntityTypeClass     EntityType = "Class"
	EntityTypeMethod    EntityType = "Method"
	EntityTypeVariable  EntityType = "Variable"
	EntityTypeImport    EntityType = "Import"
	EntityTypeInterface EntityType = "Interface"
	EntityTypeStruct    EntityType = "Struct"
	EntityTypeType      EntityType = "Type"     // TypeScript type aliases
	EntityTypeEnum      EntityType = "Enum"     // TypeScript enums
	EntityTypeProperty  EntityType = "Property" // Class/interface properties
	EntityTypeExport    EntityType = "Export"   // Export statements

	// Phase 2: Advanced TypeScript entities
	EntityTypeDecorator EntityType = "Decorator" // TypeScript decorators
	EntityTypeGeneric   EntityType = "Generic"   // Generic type parameters with constraints
	EntityTypeComponent EntityType = "Component" // React/Vue/Angular components
	EntityTypeService   EntityType = "Service"   // Injectable services
	EntityTypeHook      EntityType = "Hook"      // React hooks
	EntityTypeNamespace EntityType = "Namespace" // TypeScript namespaces
	EntityTypeModule    EntityType = "Module"    // Module declarations

	// Phase 3: Framework Integration entities
	EntityTypeEndpoint   EntityType = "Endpoint"   // API endpoints (Express routes, etc.)
	EntityTypeModel      EntityType = "Model"      // Data models (ORM entities, etc.)
	EntityTypeMiddleware EntityType = "Middleware" // Express/framework middleware
	EntityTypeController EntityType = "Controller" // Framework controllers
	EntityTypeRoute      EntityType = "Route"      // Routing definitions
	EntityTypeJSXElement EntityType = "JSXElement" // JSX elements and components
	EntityTypeAPICall    EntityType = "APICall"    // API call expressions
	EntityTypeProp       EntityType = "Prop"       // Component props/properties

	// Test Coverage entities
	EntityTypeTestFunction EntityType = "TestFunction" // Test functions/methods
	EntityTypeTestCase     EntityType = "TestCase"     // Individual test cases within a test function
	EntityTypeTestSuite    EntityType = "TestSuite"    // Test suites/classes containing multiple tests
	EntityTypeAssertion    EntityType = "Assertion"    // Individual assertions within tests
	EntityTypeMock         EntityType = "Mock"         // Mock objects/functions used in tests
	EntityTypeFixture      EntityType = "Fixture"      // Test fixtures and test data
)

// NewEntity creates a new Entity instance
func NewEntity(id, name string, entityType EntityType, filePath string, node *ts.Node) *Entity {
	return &Entity{
		ID:         id,
		Name:       name,
		Type:       entityType,
		FilePath:   filePath,
		Node:       node,
		StartByte:  uint32(node.StartByte()),
		EndByte:    uint32(node.EndByte()),
		Symbols:    make(map[string][]*ts.Node),
		Children:   make([]*Entity, 0),
		Properties: make(map[string]interface{}),
	}
}

// AddSymbol adds a symbol reference to this entity
func (e *Entity) AddSymbol(symbolType string, node *ts.Node) {
	if e.Symbols[symbolType] == nil {
		e.Symbols[symbolType] = make([]*ts.Node, 0)
	}
	e.Symbols[symbolType] = append(e.Symbols[symbolType], node)
}

// AddChild adds a child entity (e.g., method to a class)
func (e *Entity) AddChild(child *Entity) {
	child.Parent = e
	e.Children = append(e.Children, child)
}

// GetSymbols returns all symbols of a given type
func (e *Entity) GetSymbols(symbolType string) []*ts.Node {
	return e.Symbols[symbolType]
}

// HasSymbols checks if the entity has symbols of a given type
func (e *Entity) HasSymbols(symbolType string) bool {
	symbols, exists := e.Symbols[symbolType]
	return exists && len(symbols) > 0
}

// SetProperty sets a custom property for this entity
func (e *Entity) SetProperty(key string, value interface{}) {
	e.Properties[key] = value
}

// GetProperty gets a custom property for this entity
func (e *Entity) GetProperty(key string) interface{} {
	return e.Properties[key]
}

// IsMethod returns true if this entity is a method (function inside a class)
func (e *Entity) IsMethod() bool {
	return e.Type == EntityTypeMethod || (e.Type == EntityTypeFunction && e.Parent != nil && e.Parent.Type == EntityTypeClass)
}

// GetFullName returns the full qualified name of the entity (e.g., "ClassName.method_name")
func (e *Entity) GetFullName() string {
	if e.Parent != nil {
		return e.Parent.GetFullName() + "." + e.Name
	}
	return e.Name
}

// Test-related methods and properties

// IsTest returns true if this entity is a test-related entity
func (e *Entity) IsTest() bool {
	return e.Type == EntityTypeTestFunction || e.Type == EntityTypeTestCase || 
		   e.Type == EntityTypeTestSuite || e.Type == EntityTypeAssertion ||
		   e.Type == EntityTypeMock || e.Type == EntityTypeFixture ||
		   e.GetTestType() != ""
}

// GetTestType returns the type of test this entity represents
func (e *Entity) GetTestType() string {
	if testType := e.GetProperty("test_type"); testType != nil {
		if str, ok := testType.(string); ok {
			return str
		}
	}
	return ""
}

// SetTestType sets the type of test (unit, integration, e2e, benchmark, etc.)
func (e *Entity) SetTestType(testType string) {
	e.SetProperty("test_type", testType)
}

// GetTestTarget returns the entity ID that this test is targeting
func (e *Entity) GetTestTarget() string {
	if target := e.GetProperty("test_target"); target != nil {
		if str, ok := target.(string); ok {
			return str
		}
	}
	return ""
}

// SetTestTarget sets the entity ID that this test is targeting
func (e *Entity) SetTestTarget(targetEntityID string) {
	e.SetProperty("test_target", targetEntityID)
}

// GetAssertionCount returns the number of assertions in this test
func (e *Entity) GetAssertionCount() int {
	if count := e.GetProperty("assertion_count"); count != nil {
		if num, ok := count.(int); ok {
			return num
		}
	}
	return 0
}

// SetAssertionCount sets the number of assertions in this test
func (e *Entity) SetAssertionCount(count int) {
	e.SetProperty("assertion_count", count)
}

// GetTestFramework returns the testing framework used (jest, mocha, pytest, go test, etc.)
func (e *Entity) GetTestFramework() string {
	if framework := e.GetProperty("test_framework"); framework != nil {
		if str, ok := framework.(string); ok {
			return str
		}
	}
	return ""
}

// SetTestFramework sets the testing framework used
func (e *Entity) SetTestFramework(framework string) {
	e.SetProperty("test_framework", framework)
}

// GetCoverageData returns test coverage metadata
func (e *Entity) GetCoverageData() map[string]any {
	if coverage := e.GetProperty("coverage_data"); coverage != nil {
		if data, ok := coverage.(map[string]any); ok {
			return data
		}
	}
	return make(map[string]any)
}

// SetCoverageData sets test coverage metadata
func (e *Entity) SetCoverageData(data map[string]any) {
	e.SetProperty("coverage_data", data)
}

// AddCoverageMetric adds a single coverage metric
func (e *Entity) AddCoverageMetric(key string, value any) {
	data := e.GetCoverageData()
	data[key] = value
	e.SetCoverageData(data)
}

// IsTestFile returns true if this entity represents a test file
func (e *Entity) IsTestFile() bool {
	if e.Type != EntityTypeFunction && e.Type != EntityTypeClass && e.Type != EntityTypeMethod {
		return false
	}
	
	// Check file path patterns for test files
	filePath := e.FilePath
	if filePath == "" {
		return false
	}
	
	// Common test file patterns across languages
	return e.matchesTestFilePattern(filePath)
}

// matchesTestFilePattern checks if a file path matches common test file patterns
func (e *Entity) matchesTestFilePattern(filePath string) bool {
	// Go test patterns
	if len(filePath) > 8 && filePath[len(filePath)-8:] == "_test.go" {
		return true
	}
	
	// Python test patterns
	if len(filePath) > 3 && filePath[len(filePath)-3:] == ".py" {
		if len(filePath) > 8 && (filePath[0:5] == "test_" || 
			filePath[len(filePath)-8:len(filePath)-3] == "_test") {
			return true
		}
		// Check for "test" anywhere in the path
		for i := 0; i < len(filePath)-4; i++ {
			if filePath[i:i+4] == "test" {
				return true
			}
		}
	}
	
	// TypeScript/JavaScript test patterns
	if len(filePath) > 3 {
		ext := filePath[len(filePath)-3:]
		if ext == ".ts" || ext == ".js" {
			if len(filePath) > 8 {
				ending := filePath[len(filePath)-8:]
				if ending == ".test.ts" || ending == ".test.js" || 
				   ending == ".spec.ts" || ending == ".spec.js" {
					return true
				}
			}
		}
	}
	
	return false
}

// GetTestMetadata returns comprehensive test metadata
func (e *Entity) GetTestMetadata() map[string]any {
	metadata := make(map[string]any)
	metadata["is_test"] = e.IsTest()
	metadata["test_type"] = e.GetTestType()
	metadata["test_target"] = e.GetTestTarget()
	metadata["assertion_count"] = e.GetAssertionCount()
	metadata["test_framework"] = e.GetTestFramework()
	metadata["coverage_data"] = e.GetCoverageData()
	metadata["is_test_file"] = e.IsTestFile()
	return metadata
}
