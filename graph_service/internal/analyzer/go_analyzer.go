package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/onyx/onyx-tui/graph_service/internal/entities"

	ts "github.com/tree-sitter/go-tree-sitter"
	golang "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

// GoAnalyzer analyzes Go source code and extracts entities and relationships
type GoAnalyzer struct {
	parser        *ts.Parser
	language      *ts.Language
	currentFile   *entities.File
	relationships []*entities.Relationship
}

// NewGoAnalyzer creates a new Go analyzer
func NewGoAnalyzer() *GoAnalyzer {
	parser := ts.NewParser()
	language := ts.NewLanguage(golang.Language())
	parser.SetLanguage(language)

	return &GoAnalyzer{
		parser:        parser,
		language:      language,
		relationships: make([]*entities.Relationship, 0),
	}
}

// AnalyzeFile analyzes a Go file and returns the File entity with all extracted entities
func (ga *GoAnalyzer) AnalyzeFile(filePath string, content []byte) (*entities.File, []*entities.Relationship, error) {
	// Parse the file
	tree := ga.parser.Parse(content, nil)
	if tree == nil {
		return nil, nil, fmt.Errorf("failed to parse file %s", filePath)
	}

	// Create File entity
	file := entities.NewFile(filePath, "go", tree, content)
	ga.currentFile = file
	ga.relationships = make([]*entities.Relationship, 0)

	// Extract entities from the parse tree
	rootNode := tree.RootNode()
	ga.extractEntities(rootNode, nil)

	// Extract basic relationships
	ga.extractRelationships(rootNode)

	return file, ga.relationships, nil
}

// extractEntities recursively extracts entities from the parse tree
func (ga *GoAnalyzer) extractEntities(node *ts.Node, parent *entities.Entity) {
	nodeType := node.Kind()

	switch nodeType {
	case "function_declaration":
		entity := ga.extractFunction(node, parent)
		if entity != nil {
			ga.currentFile.AddEntity(entity)
			if parent != nil {
				parent.AddChild(entity)
			}
		}

	case "method_declaration":
		entity := ga.extractMethod(node, parent)
		if entity != nil {
			ga.currentFile.AddEntity(entity)
			if parent != nil {
				parent.AddChild(entity)
			}
		}

	case "type_declaration":
		ga.extractTypeDeclarations(node)

	case "import_declaration":
		ga.extractImports(node)
	}

	// Recursively process child nodes
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		ga.extractEntities(child, parent)
	}
}

// extractFunction extracts a function entity
func (ga *GoAnalyzer) extractFunction(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ga.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ga.generateEntityID("function", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeFunction, ga.currentFile.Path, node)

	// Extract signature
	parametersNode := node.ChildByFieldName("parameters")
	resultNode := node.ChildByFieldName("result")

	signature := name
	if parametersNode != nil {
		signature += ga.getNodeText(parametersNode)
	} else {
		signature += "()"
	}

	if resultNode != nil {
		signature += " " + ga.getNodeText(resultNode)
	}

	entity.Signature = signature

	// Extract body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = ga.getNodeText(bodyNode)
	}

	// Check if this is a test function and enhance accordingly
	ga.enhanceTestFunction(entity, node)

	return entity
}

// extractMethod extracts a method entity (functions with receivers)
func (ga *GoAnalyzer) extractMethod(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ga.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ga.generateEntityID("method", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeMethod, ga.currentFile.Path, node)

	// Extract receiver
	receiverNode := node.ChildByFieldName("receiver")
	if receiverNode != nil {
		receiverText := ga.getNodeText(receiverNode)
		entity.SetProperty("receiver", receiverText)
	}

	// Extract signature
	signature := ""
	if receiverNode != nil {
		signature += ga.getNodeText(receiverNode) + " "
	}
	signature += name

	parametersNode := node.ChildByFieldName("parameters")
	resultNode := node.ChildByFieldName("result")

	if parametersNode != nil {
		signature += ga.getNodeText(parametersNode)
	} else {
		signature += "()"
	}

	if resultNode != nil {
		signature += " " + ga.getNodeText(resultNode)
	}

	entity.Signature = signature

	// Extract body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = ga.getNodeText(bodyNode)
	}

	// Check if this is a test method and enhance accordingly
	ga.enhanceTestFunction(entity, node)

	return entity
}

// extractTypeDeclarations extracts type declarations (structs, interfaces, etc.)
func (ga *GoAnalyzer) extractTypeDeclarations(node *ts.Node) {
	ga.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "type_spec" {
			nameNode := n.ChildByFieldName("name")
			typeNode := n.ChildByFieldName("type")

			if nameNode != nil && typeNode != nil {
				name := ga.getNodeText(nameNode)
				typeText := ga.getNodeText(typeNode)

				// Determine entity type
				var entityType entities.EntityType
				if strings.Contains(typeText, "interface") {
					entityType = entities.EntityTypeInterface
				} else if strings.Contains(typeText, "struct") {
					entityType = entities.EntityTypeStruct
				} else {
					entityType = entities.EntityTypeClass // Default for type aliases
				}

				id := ga.generateEntityID(string(entityType), name, n)
				entity := entities.NewEntity(id, name, entityType, ga.currentFile.Path, n)
				entity.SetProperty("type_definition", typeText)

				ga.currentFile.AddEntity(entity)
			}
		}
	})
}

// extractImports extracts import declarations
func (ga *GoAnalyzer) extractImports(node *ts.Node) {
	ga.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "import_spec" {
			pathNode := n.ChildByFieldName("path")
			if pathNode != nil {
				importPath := ga.getNodeText(pathNode)
				// Remove quotes from import path
				importPath = strings.Trim(importPath, "\"")

				id := ga.generateEntityID("import", importPath, n)
				entity := entities.NewEntity(id, importPath, entities.EntityTypeImport, ga.currentFile.Path, n)
				entity.SetProperty("path", importPath)

				ga.currentFile.AddEntity(entity)
			}
		}
	})
}

// extractRelationships extracts basic relationships between entities
func (ga *GoAnalyzer) extractRelationships(node *ts.Node) {
	ga.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "call_expression" {
			ga.extractCallRelationship(n)
		}
	})
}

// extractCallRelationship extracts function/method call relationships
func (ga *GoAnalyzer) extractCallRelationship(callNode *ts.Node) {
	functionNode := callNode.ChildByFieldName("function")
	if functionNode == nil {
		return
	}

	// Find the containing function/method
	containingFunction := ga.findContainingFunction(callNode)
	if containingFunction == nil {
		return
	}

	functionName := ga.getNodeText(functionNode)

	// Create relationship using IDs (since we don't have the target entity loaded)
	relID := ga.generateRelationshipID("calls", containingFunction.ID, functionName)
	relationship := entities.NewRelationshipByID(
		relID,
		entities.RelationshipTypeCalls,
		containingFunction.ID,
		functionName, // This would be resolved to an actual entity ID in a more complete system
		containingFunction.Type, // Source entity type (function or method)
		entities.EntityTypeFunction, // Target is assumed to be a function
	)

	ga.relationships = append(ga.relationships, relationship)
}

// findContainingFunction finds the function or method that contains the given node
func (ga *GoAnalyzer) findContainingFunction(node *ts.Node) *entities.Entity {
	current := node.Parent()
	for current != nil {
		if current.Kind() == "function_declaration" || current.Kind() == "method_declaration" {
			nameNode := current.ChildByFieldName("name")
			if nameNode != nil {
				name := ga.getNodeText(nameNode)
				// Find the entity in our current file
				for _, entity := range ga.currentFile.Functions {
					if entity.Name == name {
						return entity
					}
				}
				for _, entity := range ga.currentFile.Methods {
					if entity.Name == name {
						return entity
					}
				}
			}
		}
		current = current.Parent()
	}
	return nil
}

// getNodeText extracts text content from a tree-sitter node
func (ga *GoAnalyzer) getNodeText(node *ts.Node) string {
	if node == nil {
		return ""
	}
	return node.Utf8Text(ga.currentFile.Content)
}

// generateEntityID generates a unique ID for an entity
func (ga *GoAnalyzer) generateEntityID(entityType, name string, node *ts.Node) string {
	data := fmt.Sprintf("%s:%s:%s:%d:%d",
		ga.currentFile.Path, entityType, name,
		node.StartPosition().Row, node.StartPosition().Column)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// generateRelationshipID generates a unique ID for a relationship
func (ga *GoAnalyzer) generateRelationshipID(relType, source, target string) string {
	data := fmt.Sprintf("%s:%s:%s:%s", ga.currentFile.Path, relType, source, target)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// walkNode recursively walks through all nodes in the subtree
func (ga *GoAnalyzer) walkNode(node *ts.Node, visitor func(*ts.Node)) {
	visitor(node)
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		ga.walkNode(child, visitor)
	}
}

// Test Detection and Enhancement Methods

// enhanceTestFunction analyzes a function/method entity and enhances it with test-specific information if it's a test
func (ga *GoAnalyzer) enhanceTestFunction(entity *entities.Entity, node *ts.Node) {
	if entity == nil || node == nil {
		return
	}

	// Check if this is a test function
	if ga.isTestFunction(entity) {
		// Convert to test function type
		entity.Type = entities.EntityTypeTestFunction
		
		// Set test framework
		entity.SetTestFramework("go test")
		
		// Determine test type
		testType := ga.determineGoTestType(entity)
		entity.SetTestType(testType)
		
		// Count assertions in the function body
		assertionCount := ga.countGoTestAssertions(node)
		entity.SetAssertionCount(assertionCount)
		
		// Try to determine what this test is testing
		testTarget := ga.determineTestTarget(entity)
		if testTarget != "" {
			entity.SetTestTarget(testTarget)
		}
		
		// Add test-specific properties
		ga.addGoTestProperties(entity, node)
		
		// Extract test relationships
		ga.extractTestRelationships(entity, node)
	}
}

// isTestFunction determines if a function is a Go test function
func (ga *GoAnalyzer) isTestFunction(entity *entities.Entity) bool {
	if entity == nil {
		return false
	}
	
	// Check if it's in a test file
	if !ga.isTestFile(entity.FilePath) {
		return false
	}
	
	// Check function name patterns
	name := entity.Name
	
	// Go test function patterns
	if len(name) > 4 && name[:4] == "Test" && ga.isUpperCase(name[4:5]) {
		return true
	}
	
	// Go benchmark function patterns
	if len(name) > 9 && name[:9] == "Benchmark" && ga.isUpperCase(name[9:10]) {
		return true
	}
	
	// Go example function patterns
	if len(name) > 7 && name[:7] == "Example" {
		return true
	}
	
	// Check function signature for testing.T or testing.B parameters
	signature := entity.Signature
	if signature != "" {
		if strings.Contains(signature, "*testing.T") || 
		   strings.Contains(signature, "*testing.B") || 
		   strings.Contains(signature, "*testing.M") {
			return true
		}
	}
	
	return false
}

// isTestFile checks if a file path indicates a test file
func (ga *GoAnalyzer) isTestFile(filePath string) bool {
	if filePath == "" {
		return false
	}
	
	// Go test files end with _test.go
	return len(filePath) > 8 && filePath[len(filePath)-8:] == "_test.go"
}

// isUpperCase checks if the first character of a string is uppercase
func (ga *GoAnalyzer) isUpperCase(s string) bool {
	if s == "" {
		return false
	}
	
	char := s[0]
	return char >= 'A' && char <= 'Z'
}

// determineGoTestType determines the type of Go test
func (ga *GoAnalyzer) determineGoTestType(entity *entities.Entity) string {
	if entity == nil {
		return ""
	}
	
	name := entity.Name
	signature := entity.Signature
	
	// Benchmark tests
	if len(name) > 9 && name[:9] == "Benchmark" {
		return "benchmark"
	}
	
	// Example tests
	if len(name) > 7 && name[:7] == "Example" {
		return "example"
	}
	
	// Integration tests (heuristics)
	if strings.Contains(strings.ToLower(name), "integration") ||
	   strings.Contains(strings.ToLower(name), "e2e") ||
	   strings.Contains(strings.ToLower(name), "endtoend") {
		return "integration"
	}
	
	// Check for testing.M (usually indicates setup/teardown)
	if strings.Contains(signature, "*testing.M") {
		return "suite"
	}
	
	// Default to unit test
	return "unit"
}

// countGoTestAssertions counts assertions and test calls in the function body
func (ga *GoAnalyzer) countGoTestAssertions(node *ts.Node) int {
	if node == nil {
		return 0
	}
	
	count := 0
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return 0
	}
	
	// Walk through the body and count assertion patterns
	ga.walkNode(bodyNode, func(n *ts.Node) {
		if n.Kind() == "call_expression" {
			// Get the function being called
			functionNode := n.ChildByFieldName("function")
			if functionNode != nil {
				callText := ga.getNodeText(functionNode)
				
				// Common Go test assertion patterns
				if ga.isGoTestAssertion(callText) {
					count++
				}
			}
		}
	})
	
	return count
}

// isGoTestAssertion checks if a function call is a test assertion
func (ga *GoAnalyzer) isGoTestAssertion(callText string) bool {
	if callText == "" {
		return false
	}
	
	// Common Go testing patterns
	assertionPatterns := []string{
		"t.Error", "t.Errorf", "t.Fatal", "t.Fatalf",
		"t.Fail", "t.FailNow", "t.Log", "t.Logf",
		"t.Skip", "t.Skipf", "t.SkipNow",
		"assert.", "require.", "Equal", "NotEqual",
		"True", "False", "Nil", "NotNil", "Zero", "NotZero",
		"Contains", "NotContains", "Greater", "Less",
		"Empty", "NotEmpty", "Len", "InDelta",
	}
	
	for _, pattern := range assertionPatterns {
		if strings.Contains(callText, pattern) {
			return true
		}
	}
	
	return false
}

// determineTestTarget tries to determine what entity this test is testing
func (ga *GoAnalyzer) determineTestTarget(entity *entities.Entity) string {
	if entity == nil {
		return ""
	}
	
	name := entity.Name
	
	// Remove Test prefix to get potential target name
	if len(name) > 4 && name[:4] == "Test" {
		potentialTarget := name[4:]
		
		// Look for a function with this name in the non-test files
		// This is a simplified heuristic - in practice, you'd want to
		// scan the entire project for matching function names
		return potentialTarget
	}
	
	return ""
}

// addGoTestProperties adds Go-specific test properties
func (ga *GoAnalyzer) addGoTestProperties(entity *entities.Entity, node *ts.Node) {
	if entity == nil || node == nil {
		return
	}
	
	// Check for common Go test patterns in the body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		bodyText := ga.getNodeText(bodyNode)
		
		// Check for parallel test
		if strings.Contains(bodyText, "t.Parallel()") {
			entity.SetProperty("parallel", true)
		}
		
		// Check for cleanup functions
		if strings.Contains(bodyText, "t.Cleanup(") {
			entity.SetProperty("has_cleanup", true)
		}
		
		// Check for subtest patterns
		if strings.Contains(bodyText, "t.Run(") {
			entity.SetProperty("has_subtests", true)
		}
		
		// Check for table-driven tests
		if strings.Contains(bodyText, "tests := []") || strings.Contains(bodyText, "testCases := []") {
			entity.SetProperty("table_driven", true)
		}
		
		// Check for mock usage (common patterns)
		if strings.Contains(bodyText, "mock") || strings.Contains(bodyText, "Mock") ||
		   strings.Contains(bodyText, "gomock") || strings.Contains(bodyText, "testify") {
			entity.SetProperty("uses_mocks", true)
		}
	}
}

// extractTestRelationships extracts relationships specific to test functions
func (ga *GoAnalyzer) extractTestRelationships(testEntity *entities.Entity, node *ts.Node) {
	if testEntity == nil || node == nil {
		return
	}
	
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return
	}
	
	// Track function calls made by this test
	ga.walkNode(bodyNode, func(n *ts.Node) {
		if n.Kind() == "call_expression" {
			functionNode := n.ChildByFieldName("function")
			if functionNode != nil {
				callText := ga.getNodeText(functionNode)
				
				// If this is a function call (not a test assertion), create a TESTS relationship
				if !ga.isGoTestAssertion(callText) && ga.isProductionFunctionCall(callText) {
					// Try to find the target entity
					targetEntity := ga.findEntityByName(callText)
					if targetEntity != nil {
						// Create TESTS relationship
						relID := ga.generateRelationshipID("TESTS", testEntity.ID, targetEntity.ID)
						rel := entities.NewRelationship(relID, entities.RelationshipTypeTests, testEntity, targetEntity)
						rel.SetConfidenceScore(0.8) // High confidence for direct function calls
						ga.relationships = append(ga.relationships, rel)
						
						// Also create COVERS relationship
						coverRelID := ga.generateRelationshipID("COVERS", testEntity.ID, targetEntity.ID)
						coverRel := entities.NewRelationship(coverRelID, entities.RelationshipTypeCovers, testEntity, targetEntity)
						coverRel.SetCoverageType("direct")
						ga.relationships = append(ga.relationships, coverRel)
					}
				}
			}
		}
	})
}

// isProductionFunctionCall determines if a function call is to production code (not test utilities)
func (ga *GoAnalyzer) isProductionFunctionCall(callText string) bool {
	if callText == "" {
		return false
	}
	
	// Skip test utility calls
	testUtilities := []string{
		"t.", "b.", "m.", "testing.", "assert.", "require.",
		"Mock", "mock", "Stub", "stub", "Spy", "spy",
		"setup", "teardown", "Setup", "Teardown",
	}
	
	for _, util := range testUtilities {
		if strings.Contains(callText, util) {
			return false
		}
	}
	
	// Skip standard library test-related calls
	if strings.HasPrefix(callText, "fmt.") || 
	   strings.HasPrefix(callText, "log.") ||
	   strings.HasPrefix(callText, "os.") {
		return false
	}
	
	return true
}

// findEntityByName looks for an entity with the given name in the current file
func (ga *GoAnalyzer) findEntityByName(name string) *entities.Entity {
	if ga.currentFile == nil {
		return nil
	}
	
	// Clean up the name (remove package prefixes, etc.)
	cleanName := name
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		if len(parts) > 0 {
			cleanName = parts[len(parts)-1]
		}
	}
	
	// Search through all entities in the current file
	for _, entity := range ga.currentFile.GetAllEntities() {
		if entity.Name == cleanName {
			return entity
		}
	}
	
	return nil
}
