package analyzer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/onyx/onyx-tui/graph_service/internal/entities"

	ts "github.com/tree-sitter/go-tree-sitter"
	python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// PythonAnalyzer analyzes Python source code and extracts entities and relationships
type PythonAnalyzer struct {
	parser        *ts.Parser
	language      *ts.Language
	currentFile   *entities.File
	relationships []*entities.Relationship
}

// NewPythonAnalyzer creates a new Python analyzer
func NewPythonAnalyzer() *PythonAnalyzer {
	parser := ts.NewParser()
	language := ts.NewLanguage(python.Language())
	parser.SetLanguage(language)

	return &PythonAnalyzer{
		parser:        parser,
		language:      language,
		relationships: make([]*entities.Relationship, 0),
	}
}

// AnalyzeFile analyzes a Python file and returns the File entity with all extracted entities
func (pa *PythonAnalyzer) AnalyzeFile(filePath string, content []byte) (*entities.File, []*entities.Relationship, error) {
	// Parse the file
	tree := pa.parser.ParseCtx(context.Background(), content, nil)
	if tree == nil {
		return nil, nil, fmt.Errorf("failed to parse file %s", filePath)
	}

	// Create File entity
	file := entities.NewFile(filePath, "python", tree, content)
	pa.currentFile = file
	pa.relationships = make([]*entities.Relationship, 0)

	// Extract entities from the parse tree
	rootNode := tree.RootNode()
	pa.extractEntities(rootNode, nil)

	// Extract relationships (function calls, imports, inheritance, etc.)
	pa.extractRelationships(rootNode)

	return file, pa.relationships, nil
}

// extractEntities recursively extracts entities from the parse tree
func (pa *PythonAnalyzer) extractEntities(node *ts.Node, parent *entities.Entity) {
	nodeType := node.Kind()

	switch nodeType {
	case "function_definition":
		entity := pa.extractFunction(node, parent)
		if entity != nil {
			pa.currentFile.AddEntity(entity)
			if parent != nil {
				parent.AddChild(entity)
			}
			// Recursively process the function body for nested entities (only child nodes)
			for i := uint(0); i < node.ChildCount(); i++ {
				child := node.Child(i)
				pa.extractEntities(child, entity)
			}
			return
		}

	case "class_definition":
		entity := pa.extractClass(node, parent)
		if entity != nil {
			pa.currentFile.AddEntity(entity)
			if parent != nil {
				parent.AddChild(entity)
			}
			// Recursively process the class body for methods (only child nodes)
			for i := uint(0); i < node.ChildCount(); i++ {
				child := node.Child(i)
				pa.extractEntities(child, entity)
			}
			return
		}

	case "import_statement", "import_from_statement":
		entity := pa.extractImport(node)
		if entity != nil {
			pa.currentFile.AddEntity(entity)
		}

	case "assignment":
		entity := pa.extractVariable(node, parent)
		if entity != nil {
			pa.currentFile.AddEntity(entity)
		}
	}

	// Recursively process child nodes
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		pa.extractEntities(child, parent)
	}
}

// extractFunction extracts a function entity with enhanced capabilities
func (pa *PythonAnalyzer) extractFunction(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := pa.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := pa.generateEntityID("function", name, node)

	// Determine if this is a method (inside a class)
	entityType := entities.EntityTypeFunction
	if parent != nil && parent.Type == entities.EntityTypeClass {
		entityType = entities.EntityTypeMethod
	}

	entity := entities.NewEntity(id, name, entityType, pa.currentFile.Path, node)

	// Extract signature (parameters)
	parametersNode := node.ChildByFieldName("parameters")
	if parametersNode != nil {
		entity.Signature = pa.getNodeText(parametersNode)
		// Extract parameter types
		pa.extractParameterTypes(parametersNode, entity)
	} else {
		entity.Signature = "()"
	}

	// Extract return type annotation
	returnTypeNode := node.ChildByFieldName("return_type")
	if returnTypeNode != nil {
		entity.SetProperty("return_type", pa.getNodeText(returnTypeNode))
		entity.AddSymbol("return_type", returnTypeNode)
	}

	// Extract decorators
	pa.extractDecorators(node, entity)

	// Extract body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = pa.getNodeText(bodyNode)
	}

	// Extract docstring
	entity.DocString = pa.extractDocstring(node)

	// Extract function calls within this function
	pa.extractFunctionCalls(node, entity)

	// Check if this is a test function and enhance accordingly
	pa.enhanceTestFunction(entity, node)

	return entity
}

// extractClass extracts a class entity with inheritance tracking
func (pa *PythonAnalyzer) extractClass(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := pa.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := pa.generateEntityID("class", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeClass, pa.currentFile.Path, node)

	// Extract superclasses (inheritance)
	superclassesNode := node.ChildByFieldName("superclasses")
	if superclassesNode != nil {
		superclassText := pa.getNodeText(superclassesNode)
		entity.SetProperty("superclasses", superclassText)

		// Parse individual base classes
		pa.extractBaseClasses(superclassesNode, entity)
	}

	// Extract decorators
	pa.extractDecorators(node, entity)

	// Extract body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = pa.getNodeText(bodyNode)
	}

	// Extract docstring
	entity.DocString = pa.extractDocstring(node)

	return entity
}

// extractImport extracts import entities with better parsing
func (pa *PythonAnalyzer) extractImport(node *ts.Node) *entities.Entity {
	nodeType := node.Kind()
	var importText, moduleName string

	switch nodeType {
	case "import_statement":
		// Handle: import module1, module2
		importText = pa.getNodeText(node)
		// Extract the first module name for the entity name
		nameNode := node.Child(1) // Skip 'import' keyword
		if nameNode != nil {
			moduleName = pa.getNodeText(nameNode)
		}
	case "import_from_statement":
		// Handle: from module import name1, name2
		importText = pa.getNodeText(node)
		moduleNameNode := node.ChildByFieldName("module_name")
		if moduleNameNode != nil {
			moduleName = pa.getNodeText(moduleNameNode)
		}
	default:
		return nil
	}

	if importText == "" {
		return nil
	}

	// Use module name if available, otherwise full import text
	entityName := moduleName
	if entityName == "" {
		entityName = importText
	}

	// Generate unique ID based on import statement
	id := pa.generateEntityID("import", importText, node)

	entity := entities.NewEntity(id, entityName, entities.EntityTypeImport, pa.currentFile.Path, node)
	entity.Signature = importText
	entity.SetProperty("import_type", nodeType)
	entity.SetProperty("full_import", importText)

	return entity
}

// extractVariable extracts a variable assignment
func (pa *PythonAnalyzer) extractVariable(node *ts.Node, parent *entities.Entity) *entities.Entity {
	leftNode := node.ChildByFieldName("left")
	if leftNode == nil {
		return nil
	}

	name := pa.getNodeText(leftNode)
	if name == "" || strings.Contains(name, ".") {
		// Skip complex assignments and attribute assignments
		return nil
	}

	// Generate unique ID
	id := pa.generateEntityID("variable", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeVariable, pa.currentFile.Path, node)

	// Get the assignment value
	rightNode := node.ChildByFieldName("right")
	if rightNode != nil {
		entity.SetProperty("value", pa.getNodeText(rightNode))
	}

	// Extract type annotation if present
	typeNode := node.ChildByFieldName("type")
	if typeNode != nil {
		entity.SetProperty("type_annotation", pa.getNodeText(typeNode))
	}

	return entity
}

// extractParameterTypes extracts parameter type annotations
func (pa *PythonAnalyzer) extractParameterTypes(parametersNode *ts.Node, entity *entities.Entity) {
	pa.walkNode(parametersNode, func(n *ts.Node) {
		if n.Kind() == "typed_parameter" {
			typeNode := n.ChildByFieldName("type")
			if typeNode != nil {
				entity.AddSymbol("parameter_type", typeNode)
			}
		}
	})
}

// extractBaseClasses extracts base classes from superclasses node
func (pa *PythonAnalyzer) extractBaseClasses(superclassesNode *ts.Node, entity *entities.Entity) {
	pa.walkNode(superclassesNode, func(n *ts.Node) {
		if n.Kind() == "identifier" || n.Kind() == "attribute" {
			entity.AddSymbol("base_class", n)
		}
	})
}

// extractDecorators extracts decorators from functions and classes
func (pa *PythonAnalyzer) extractDecorators(node *ts.Node, entity *entities.Entity) {
	// Look for decorator siblings before the function/class definition
	current := node.PrevSibling()
	decorators := make([]string, 0)

	for current != nil && current.Kind() == "decorator" {
		decoratorText := pa.getNodeText(current)
		decorators = append([]string{decoratorText}, decorators...) // Prepend to maintain order
		entity.AddSymbol("decorator", current)
		current = current.PrevSibling()
	}

	if len(decorators) > 0 {
		entity.SetProperty("decorators", decorators)
	}
}

// extractFunctionCalls finds function calls within a node and adds them as symbols
func (pa *PythonAnalyzer) extractFunctionCalls(node *ts.Node, entity *entities.Entity) {
	pa.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "call" {
			functionNode := n.ChildByFieldName("function")
			if functionNode != nil {
				entity.AddSymbol("call", functionNode)
			}
		}
	})
}

// extractRelationships extracts relationships between entities
func (pa *PythonAnalyzer) extractRelationships(node *ts.Node) {
	// Extract function calls as relationships
	pa.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "call" {
			pa.extractCallRelationship(n)
		}
	})

	// Extract inheritance relationships
	for _, entity := range pa.currentFile.GetAllEntities() {
		if entity.Type == entities.EntityTypeClass {
			pa.extractInheritanceRelationships(entity)
		}
	}

	// Extract file-entity containment relationships
	for _, entity := range pa.currentFile.GetAllEntities() {
		if entity.Type != entities.EntityTypeImport {
			rel := entities.NewRelationshipByID(
				pa.generateRelationshipID("contains", pa.currentFile.Path, entity.ID),
				entities.RelationshipTypeContains,
				pa.currentFile.Path, // File as source
				entity.ID,
				entities.EntityTypeFile, // Source is a file
				entity.Type,            // Target type depends on the actual entity
			)
			pa.relationships = append(pa.relationships, rel)
		}
	}
}

// extractInheritanceRelationships creates inheritance relationships for classes
func (pa *PythonAnalyzer) extractInheritanceRelationships(classEntity *entities.Entity) {
	baseClassSymbols := classEntity.GetSymbols("base_class")
	for _, baseClassNode := range baseClassSymbols {
		baseClassName := pa.getNodeText(baseClassNode)
		if baseClassName != "" {
			relID := pa.generateRelationshipID("inherits", classEntity.ID, baseClassName)
			rel := entities.NewRelationshipByID(relID, entities.RelationshipTypeInherits, classEntity.ID, baseClassName, entities.EntityTypeClass, entities.EntityTypeClass)
			rel.SetLocation(pa.currentFile.Path, uint32(baseClassNode.StartByte()), uint32(baseClassNode.EndByte()))
			pa.relationships = append(pa.relationships, rel)
		}
	}
}

// extractCallRelationship extracts a function call relationship
func (pa *PythonAnalyzer) extractCallRelationship(callNode *ts.Node) {
	functionNode := callNode.ChildByFieldName("function")
	if functionNode == nil {
		return
	}

	calledFunction := pa.getNodeText(functionNode)
	if calledFunction == "" {
		return
	}

	// Find the calling function (walk up the tree)
	callingFunction := pa.findContainingFunction(callNode)
	if callingFunction == nil {
		return
	}

	// Create relationship
	relID := pa.generateRelationshipID("calls", callingFunction.ID, calledFunction)
	rel := entities.NewRelationshipByID(relID, entities.RelationshipTypeCalls, callingFunction.ID, calledFunction, callingFunction.Type, entities.EntityTypeFunction)
	rel.SetLocation(pa.currentFile.Path, uint32(callNode.StartByte()), uint32(callNode.EndByte()))

	pa.relationships = append(pa.relationships, rel)
}

// findContainingFunction finds the function that contains the given node
func (pa *PythonAnalyzer) findContainingFunction(node *ts.Node) *entities.Entity {
	current := node.Parent()
	for current != nil {
		if current.Kind() == "function_definition" {
			nameNode := current.ChildByFieldName("name")
			if nameNode != nil {
				name := pa.getNodeText(nameNode)
				// Find the entity in the current file
				for _, entity := range pa.currentFile.GetAllEntities() {
					if entity.Name == name && (entity.Type == entities.EntityTypeFunction || entity.Type == entities.EntityTypeMethod) {
						return entity
					}
				}
			}
		}
		current = current.Parent()
	}
	return nil
}

// extractDocstring extracts the docstring from a function or class
func (pa *PythonAnalyzer) extractDocstring(node *ts.Node) string {
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return ""
	}

	// Look for the first string literal in the body
	if bodyNode.ChildCount() > 0 {
		firstChild := bodyNode.Child(0)
		if firstChild.Kind() == "expression_statement" && firstChild.ChildCount() > 0 {
			expr := firstChild.Child(0)
			if expr.Kind() == "string" {
				return pa.getNodeText(expr)
			}
		}
	}

	return ""
}

// Helper functions

// getNodeText extracts text content from a tree-sitter node
func (pa *PythonAnalyzer) getNodeText(node *ts.Node) string {
	if node == nil {
		return ""
	}

	start := node.StartByte()
	end := node.EndByte()

	if start >= uint(len(pa.currentFile.Content)) || end > uint(len(pa.currentFile.Content)) {
		return ""
	}

	return string(pa.currentFile.Content[start:end])
}

// generateEntityID generates a unique ID for an entity
func (pa *PythonAnalyzer) generateEntityID(entityType, name string, node *ts.Node) string {
	base := fmt.Sprintf("%s:%s:%s:%d:%d",
		entityType,
		pa.currentFile.Path,
		name,
		node.StartByte(),
		node.EndByte())

	hash := sha256.Sum256([]byte(base))
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars for shorter IDs
}

// generateRelationshipID generates a unique ID for a relationship
func (pa *PythonAnalyzer) generateRelationshipID(relType, source, target string) string {
	base := fmt.Sprintf("%s:%s:%s", relType, source, target)
	hash := sha256.Sum256([]byte(base))
	return hex.EncodeToString(hash[:])[:16]
}

// walkNode recursively walks a tree-sitter node and calls the visitor function on each node
func (pa *PythonAnalyzer) walkNode(node *ts.Node, visitor func(*ts.Node)) {
	visitor(node)

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		pa.walkNode(child, visitor)
	}
}

// Python Test Detection and Enhancement Methods

// enhanceTestFunction analyzes a function/method entity and enhances it with test-specific information if it's a test
func (pa *PythonAnalyzer) enhanceTestFunction(entity *entities.Entity, node *ts.Node) {
	if entity == nil || node == nil {
		return
	}

	// Check if this is a test function
	if pa.isPythonTestFunction(entity, node) {
		// Convert to test function type
		entity.Type = entities.EntityTypeTestFunction
		
		// Determine test framework
		testFramework := pa.determinePythonTestFramework(entity, node)
		entity.SetTestFramework(testFramework)
		
		// Determine test type
		testType := pa.determinePythonTestType(entity)
		entity.SetTestType(testType)
		
		// Count assertions in the function body
		assertionCount := pa.countPythonTestAssertions(node)
		entity.SetAssertionCount(assertionCount)
		
		// Try to determine what this test is testing
		testTarget := pa.determinePythonTestTarget(entity)
		if testTarget != "" {
			entity.SetTestTarget(testTarget)
		}
		
		// Add test-specific properties
		pa.addPythonTestProperties(entity, node)
		
		// Extract test relationships
		pa.extractPythonTestRelationships(entity, node)
	}
}

// isPythonTestFunction determines if a function is a Python test function
func (pa *PythonAnalyzer) isPythonTestFunction(entity *entities.Entity, node *ts.Node) bool {
	if entity == nil {
		return false
	}
	
	// Check if it's in a test file
	if !pa.isPythonTestFile(entity.FilePath) {
		return false
	}
	
	// Check function name patterns
	name := entity.Name
	
	// pytest/unittest function patterns
	if len(name) > 5 && name[:5] == "test_" {
		return true
	}
	
	// unittest class method patterns (if parent is a test class)
	if entity.Parent != nil && entity.Parent.Type == entities.EntityTypeClass {
		parentName := entity.Parent.Name
		// Check if parent is a test class
		if pa.isPythonTestClass(parentName) {
			return true
		}
	}
	
	// Check for unittest method patterns within TestCase classes
	if len(name) > 4 && name[:4] == "test" && pa.isUpperCase(name[4:5]) {
		return true
	}
	
	// Check decorators for test frameworks
	return pa.hasPythonTestDecorators(node)
}

// isPythonTestFile checks if a file path indicates a Python test file
func (pa *PythonAnalyzer) isPythonTestFile(filePath string) bool {
	if filePath == "" {
		return false
	}
	
	// Common Python test file patterns
	fileName := filePath
	if strings.Contains(filePath, "/") {
		parts := strings.Split(filePath, "/")
		fileName = parts[len(parts)-1]
	}
	
	// test_*.py pattern
	if len(fileName) > 5 && fileName[:5] == "test_" && fileName[len(fileName)-3:] == ".py" {
		return true
	}
	
	// *_test.py pattern
	if len(fileName) > 8 && fileName[len(fileName)-8:] == "_test.py" {
		return true
	}
	
	// tests.py pattern
	if fileName == "tests.py" {
		return true
	}
	
	// Check for "test" anywhere in path
	return strings.Contains(strings.ToLower(filePath), "test")
}

// isPythonTestClass checks if a class name indicates a test class
func (pa *PythonAnalyzer) isPythonTestClass(className string) bool {
	if className == "" {
		return false
	}
	
	// Common test class patterns
	if len(className) > 4 && className[:4] == "Test" {
		return true
	}
	
	if strings.Contains(className, "Test") || strings.Contains(className, "test") {
		return true
	}
	
	return false
}

// isUpperCase checks if the first character of a string is uppercase
func (pa *PythonAnalyzer) isUpperCase(s string) bool {
	if s == "" {
		return false
	}
	
	char := s[0]
	return char >= 'A' && char <= 'Z'
}

// hasPythonTestDecorators checks if a function has test-related decorators
func (pa *PythonAnalyzer) hasPythonTestDecorators(node *ts.Node) bool {
	if node == nil {
		return false
	}
	
	// Look for decorators
	parent := node.Parent()
	if parent == nil {
		return false
	}
	
	// Check previous siblings for decorators
	for i := uint(0); i < parent.ChildCount(); i++ {
		child := parent.Child(i)
		if child == node {
			break
		}
		
		if child.Kind() == "decorator" {
			decoratorText := pa.getNodeText(child)
			
			// Common test decorators
			testDecorators := []string{
				"@pytest.mark", "@unittest", "@parameterized", "@mock.patch",
				"@patch", "@skip", "@skipIf", "@expectedFailure", 
				"@test", "@Test", "@given", "@hypothesis", "@property_test",
			}
			
			for _, decorator := range testDecorators {
				if strings.Contains(decoratorText, decorator) {
					return true
				}
			}
		}
	}
	
	return false
}

// determinePythonTestFramework determines the testing framework used
func (pa *PythonAnalyzer) determinePythonTestFramework(entity *entities.Entity, node *ts.Node) string {
	if entity == nil || node == nil {
		return "unknown"
	}
	
	// Check imports in the file to determine framework
	fileContent := string(pa.currentFile.Content)
	
	// pytest indicators
	if strings.Contains(fileContent, "import pytest") || 
	   strings.Contains(fileContent, "from pytest") ||
	   strings.Contains(fileContent, "@pytest") {
		return "pytest"
	}
	
	// unittest indicators
	if strings.Contains(fileContent, "import unittest") || 
	   strings.Contains(fileContent, "from unittest") ||
	   strings.Contains(fileContent, "TestCase") {
		return "unittest"
	}
	
	// nose indicators
	if strings.Contains(fileContent, "import nose") || 
	   strings.Contains(fileContent, "from nose") {
		return "nose"
	}
	
	// hypothesis indicators
	if strings.Contains(fileContent, "import hypothesis") || 
	   strings.Contains(fileContent, "@given") {
		return "hypothesis"
	}
	
	// Check function name patterns
	if len(entity.Name) > 5 && entity.Name[:5] == "test_" {
		return "pytest"  // Most likely pytest
	}
	
	return "unittest"  // Default assumption
}

// determinePythonTestType determines the type of Python test
func (pa *PythonAnalyzer) determinePythonTestType(entity *entities.Entity) string {
	if entity == nil {
		return ""
	}
	
	name := strings.ToLower(entity.Name)
	
	// Integration test patterns
	if strings.Contains(name, "integration") ||
	   strings.Contains(name, "e2e") ||
	   strings.Contains(name, "end_to_end") ||
	   strings.Contains(name, "endtoend") {
		return "integration"
	}
	
	// Performance/benchmark test patterns
	if strings.Contains(name, "performance") ||
	   strings.Contains(name, "benchmark") ||
	   strings.Contains(name, "speed") ||
	   strings.Contains(name, "timing") {
		return "performance"
	}
	
	// Smoke test patterns
	if strings.Contains(name, "smoke") {
		return "smoke"
	}
	
	// Property-based test patterns
	if strings.Contains(name, "property") ||
	   strings.Contains(name, "hypothesis") {
		return "property"
	}
	
	// Default to unit test
	return "unit"
}

// countPythonTestAssertions counts assertions in a Python test function
func (pa *PythonAnalyzer) countPythonTestAssertions(node *ts.Node) int {
	if node == nil {
		return 0
	}
	
	count := 0
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return 0
	}
	
	// Walk through the body and count assertion patterns
	pa.walkNode(bodyNode, func(n *ts.Node) {
		if n.Kind() == "call" {
			// Get the function being called
			functionNode := n.ChildByFieldName("function")
			if functionNode != nil {
				callText := pa.getNodeText(functionNode)
				
				// Common Python test assertion patterns
				if pa.isPythonTestAssertion(callText) {
					count++
				}
			}
		}
		
		// Also check for assert statements (Python built-in)
		if n.Kind() == "assert_statement" {
			count++
		}
	})
	
	return count
}

// isPythonTestAssertion checks if a function call is a test assertion
func (pa *PythonAnalyzer) isPythonTestAssertion(callText string) bool {
	if callText == "" {
		return false
	}
	
	// unittest assertion patterns
	unittestAssertions := []string{
		"self.assert", "self.assertEqual", "self.assertNotEqual",
		"self.assertTrue", "self.assertFalse", "self.assertIsNone",
		"self.assertIsNotNone", "self.assertIn", "self.assertNotIn",
		"self.assertIsInstance", "self.assertNotIsInstance",
		"self.assertRaises", "self.assertRaisesRegex",
		"self.assertAlmostEqual", "self.assertNotAlmostEqual",
		"self.assertGreater", "self.assertGreaterEqual",
		"self.assertLess", "self.assertLessEqual",
		"self.assertCountEqual", "self.assertMultiLineEqual",
	}
	
	// pytest assertion patterns
	pytestAssertions := []string{
		"pytest.raises", "pytest.warns", "pytest.approx",
		"pytest.fail", "pytest.skip", "pytest.xfail",
	}
	
	// Third-party assertion libraries
	thirdPartyAssertions := []string{
		"nose.assert", "sure.assert", "expect(", "should.",
		"assertThat", "hamcrest.", "mock.assert",
	}
	
	allAssertions := append(unittestAssertions, pytestAssertions...)
	allAssertions = append(allAssertions, thirdPartyAssertions...)
	
	for _, assertion := range allAssertions {
		if strings.Contains(callText, assertion) {
			return true
		}
	}
	
	return false
}

// determinePythonTestTarget tries to determine what this test is testing
func (pa *PythonAnalyzer) determinePythonTestTarget(entity *entities.Entity) string {
	if entity == nil {
		return ""
	}
	
	name := entity.Name
	
	// Remove test_ prefix to get potential target name
	if len(name) > 5 && name[:5] == "test_" {
		potentialTarget := name[5:]
		return potentialTarget
	}
	
	// Remove Test prefix for unittest style
	if len(name) > 4 && name[:4] == "test" {
		potentialTarget := name[4:]
		if len(potentialTarget) > 0 {
			// Convert first letter to lowercase
			potentialTarget = strings.ToLower(potentialTarget[:1]) + potentialTarget[1:]
			return potentialTarget
		}
	}
	
	return ""
}

// addPythonTestProperties adds Python-specific test properties
func (pa *PythonAnalyzer) addPythonTestProperties(entity *entities.Entity, node *ts.Node) {
	if entity == nil || node == nil {
		return
	}
	
	// Check for common Python test patterns in the body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		bodyText := pa.getNodeText(bodyNode)
		
		// Check for parameterized tests
		if strings.Contains(bodyText, "@parameterized") ||
		   strings.Contains(bodyText, "@pytest.mark.parametrize") {
			entity.SetProperty("parameterized", true)
		}
		
		// Check for skip patterns
		if strings.Contains(bodyText, "pytest.skip") ||
		   strings.Contains(bodyText, "unittest.skip") ||
		   strings.Contains(bodyText, "@skip") {
			entity.SetProperty("can_skip", true)
		}
		
		// Check for mock usage
		if strings.Contains(bodyText, "mock") || strings.Contains(bodyText, "Mock") ||
		   strings.Contains(bodyText, "@patch") || strings.Contains(bodyText, "MagicMock") {
			entity.SetProperty("uses_mocks", true)
		}
		
		// Check for fixtures (pytest)
		if strings.Contains(bodyText, "@fixture") ||
		   strings.Contains(bodyText, "fixture") {
			entity.SetProperty("uses_fixtures", true)
		}
		
		// Check for database/external dependencies
		if strings.Contains(bodyText, "database") || strings.Contains(bodyText, "db") ||
		   strings.Contains(bodyText, "requests") || strings.Contains(bodyText, "http") {
			entity.SetProperty("external_dependencies", true)
		}
	}
	
	// Check decorators for additional properties
	if pa.hasPythonTestDecorators(node) {
		entity.SetProperty("has_decorators", true)
	}
}

// extractPythonTestRelationships extracts relationships specific to Python test functions
func (pa *PythonAnalyzer) extractPythonTestRelationships(testEntity *entities.Entity, node *ts.Node) {
	if testEntity == nil || node == nil {
		return
	}
	
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return
	}
	
	// Track function calls made by this test
	pa.walkNode(bodyNode, func(n *ts.Node) {
		if n.Kind() == "call" {
			functionNode := n.ChildByFieldName("function")
			if functionNode != nil {
				callText := pa.getNodeText(functionNode)
				
				// If this is a function call (not a test assertion), create a TESTS relationship
				if !pa.isPythonTestAssertion(callText) && pa.isPythonProductionFunctionCall(callText) {
					// Try to find the target entity
					targetEntity := pa.findPythonEntityByName(callText)
					if targetEntity != nil {
						// Create TESTS relationship
						relID := pa.generateRelationshipID("TESTS", testEntity.ID, targetEntity.ID)
						rel := entities.NewRelationship(relID, entities.RelationshipTypeTests, testEntity, targetEntity)
						rel.SetConfidenceScore(0.8) // High confidence for direct function calls
						pa.relationships = append(pa.relationships, rel)
						
						// Also create COVERS relationship
						coverRelID := pa.generateRelationshipID("COVERS", testEntity.ID, targetEntity.ID)
						coverRel := entities.NewRelationship(coverRelID, entities.RelationshipTypeCovers, testEntity, targetEntity)
						coverRel.SetCoverageType("direct")
						pa.relationships = append(pa.relationships, coverRel)
					}
				}
			}
		}
	})
}

// isPythonProductionFunctionCall determines if a function call is to production code
func (pa *PythonAnalyzer) isPythonProductionFunctionCall(callText string) bool {
	if callText == "" {
		return false
	}
	
	// Skip test utility calls
	testUtilities := []string{
		"self.assert", "pytest.", "unittest.", "mock.", "Mock",
		"patch", "fixture", "setUp", "tearDown", "setUpClass", "tearDownClass",
		"skip", "xfail", "raises", "warns", "approx",
	}
	
	for _, util := range testUtilities {
		if strings.Contains(callText, util) {
			return false
		}
	}
	
	// Skip standard library test-related calls
	if strings.HasPrefix(callText, "print(") || 
	   strings.HasPrefix(callText, "len(") ||
	   strings.HasPrefix(callText, "str(") ||
	   strings.HasPrefix(callText, "int(") {
		return false
	}
	
	return true
}

// findPythonEntityByName looks for an entity with the given name in the current file
func (pa *PythonAnalyzer) findPythonEntityByName(name string) *entities.Entity {
	if pa.currentFile == nil {
		return nil
	}
	
	// Clean up the name (remove self., class prefixes, etc.)
	cleanName := name
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		if len(parts) > 0 {
			cleanName = parts[len(parts)-1]
		}
	}
	
	// Remove parentheses if they exist
	if strings.Contains(cleanName, "(") {
		cleanName = strings.Split(cleanName, "(")[0]
	}
	
	// Search through all entities in the current file
	for _, entity := range pa.currentFile.GetAllEntities() {
		if entity.Name == cleanName {
			return entity
		}
	}
	
	return nil
}
