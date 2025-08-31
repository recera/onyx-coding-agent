package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/onyx/onyx-tui/graph_service/internal/entities"

	ts "github.com/tree-sitter/go-tree-sitter"
	golang "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

// EnhancedGoAnalyzer provides advanced Go source code analysis
type EnhancedGoAnalyzer struct {
	parser        *ts.Parser
	language      *ts.Language
	currentFile   *entities.File
	relationships []*entities.Relationship

	// Enhanced tracking
	typeRegistry      map[string]*entities.Entity   // Track all types (structs, interfaces, aliases)
	methodRegistry    map[string][]*entities.Entity // Track methods by receiver type
	interfaceRegistry map[string]*InterfaceInfo     // Track interface definitions
	packageImports    map[string]string             // Track import aliases
}

// InterfaceInfo stores information about interface definitions
type InterfaceInfo struct {
	Entity  *entities.Entity
	Methods map[string]*MethodSignature
}

// MethodSignature represents a method signature for interface analysis
type MethodSignature struct {
	Name       string
	Parameters string
	ReturnType string
}

// NewEnhancedGoAnalyzer creates a new enhanced Go analyzer
func NewEnhancedGoAnalyzer() *EnhancedGoAnalyzer {
	parser := ts.NewParser()
	language := ts.NewLanguage(golang.Language())
	parser.SetLanguage(language)

	return &EnhancedGoAnalyzer{
		parser:            parser,
		language:          language,
		relationships:     make([]*entities.Relationship, 0),
		typeRegistry:      make(map[string]*entities.Entity),
		methodRegistry:    make(map[string][]*entities.Entity),
		interfaceRegistry: make(map[string]*InterfaceInfo),
		packageImports:    make(map[string]string),
	}
}

// AnalyzeFile analyzes a Go file with enhanced relationship detection
func (ega *EnhancedGoAnalyzer) AnalyzeFile(filePath string, content []byte) (*entities.File, []*entities.Relationship, error) {
	// Parse the file
	tree := ega.parser.Parse(content, nil)
	if tree == nil {
		return nil, nil, fmt.Errorf("failed to parse file %s", filePath)
	}

	// Create File entity
	file := entities.NewFile(filePath, "go", tree, content)
	ega.currentFile = file
	ega.relationships = make([]*entities.Relationship, 0)

	// Reset registries for each file
	ega.typeRegistry = make(map[string]*entities.Entity)
	ega.methodRegistry = make(map[string][]*entities.Entity)
	ega.interfaceRegistry = make(map[string]*InterfaceInfo)
	ega.packageImports = make(map[string]string)

	rootNode := tree.RootNode()

	// Phase 1: Extract basic entities
	ega.extractEntities(rootNode, nil)

	// Phase 2: Build type registry and analyze interfaces
	ega.buildTypeRegistry()

	// Phase 3: Extract enhanced relationships
	ega.extractEnhancedRelationships(rootNode)

	// Phase 4: Detect interface implementations
	ega.detectInterfaceImplementations()

	return file, ega.relationships, nil
}

// extractEntities extracts entities with enhanced type tracking
func (ega *EnhancedGoAnalyzer) extractEntities(node *ts.Node, parent *entities.Entity) {
	nodeType := node.Kind()

	switch nodeType {
	case "function_declaration":
		entity := ega.extractFunction(node, parent)
		if entity != nil {
			ega.currentFile.AddEntity(entity)
		}

	case "method_declaration":
		entity := ega.extractMethod(node, parent)
		if entity != nil {
			ega.currentFile.AddEntity(entity)
			// Track method by receiver type
			receiverType := ega.extractReceiverType(node)
			if receiverType != "" {
				ega.methodRegistry[receiverType] = append(ega.methodRegistry[receiverType], entity)
			}
		}

	case "type_declaration":
		ega.extractEnhancedTypeDeclarations(node)

	case "import_declaration":
		ega.extractEnhancedImports(node)

	case "var_declaration", "const_declaration":
		ega.extractVariableDeclarations(node)
	}

	// Recursively process child nodes
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		ega.extractEntities(child, parent)
	}
}

// extractFunction extracts function entities with enhanced signature analysis
func (ega *EnhancedGoAnalyzer) extractFunction(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ega.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	id := ega.generateEntityID("function", name, node)
	entity := entities.NewEntity(id, name, entities.EntityTypeFunction, ega.currentFile.Path, node)

	// Enhanced signature extraction
	signature := ega.buildFunctionSignature(node, name)
	entity.Signature = signature

	// Extract parameters with types
	ega.extractParameterInfo(node, entity)

	// Extract return types
	ega.extractReturnTypeInfo(node, entity)

	// Extract body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = ega.getNodeText(bodyNode)
	}

	// Extract comments/documentation
	entity.DocString = ega.extractDocumentation(node)

	return entity
}

// extractMethod extracts method entities with receiver analysis
func (ega *EnhancedGoAnalyzer) extractMethod(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ega.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	id := ega.generateEntityID("method", name, node)
	entity := entities.NewEntity(id, name, entities.EntityTypeMethod, ega.currentFile.Path, node)

	// Extract and analyze receiver
	receiverNode := node.ChildByFieldName("receiver")
	if receiverNode != nil {
		receiverText := ega.getNodeText(receiverNode)
		receiverType := ega.extractReceiverType(node)

		entity.SetProperty("receiver", receiverText)
		entity.SetProperty("receiver_type", receiverType)
		entity.SetProperty("is_pointer_receiver", ega.isPointerReceiver(receiverNode))
	}

	// Enhanced signature extraction
	signature := ega.buildMethodSignature(node, name)
	entity.Signature = signature

	// Extract parameters and return types
	ega.extractParameterInfo(node, entity)
	ega.extractReturnTypeInfo(node, entity)

	// Extract body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = ega.getNodeText(bodyNode)
	}

	// Extract documentation
	entity.DocString = ega.extractDocumentation(node)

	return entity
}

// extractEnhancedTypeDeclarations extracts type declarations with detailed analysis
func (ega *EnhancedGoAnalyzer) extractEnhancedTypeDeclarations(node *ts.Node) {
	ega.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "type_spec" {
			entity := ega.extractTypeSpec(n)
			if entity != nil {
				ega.currentFile.AddEntity(entity)
				ega.typeRegistry[entity.Name] = entity
			}
		}
	})
}

// extractTypeSpec extracts a single type specification with enhanced analysis
func (ega *EnhancedGoAnalyzer) extractTypeSpec(node *ts.Node) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	typeNode := node.ChildByFieldName("type")

	if nameNode == nil || typeNode == nil {
		return nil
	}

	name := ega.getNodeText(nameNode)
	typeText := ega.getNodeText(typeNode)

	// Determine entity type and extract detailed information
	var entityType entities.EntityType
	if typeNode.Kind() == "interface_type" {
		entityType = entities.EntityTypeInterface
		return ega.extractInterfaceDefinition(node, name, typeNode)
	} else if typeNode.Kind() == "struct_type" {
		entityType = entities.EntityTypeStruct
		return ega.extractStructDefinition(node, name, typeNode)
	} else {
		// Type alias or type definition
		entityType = entities.EntityTypeClass
	}

	id := ega.generateEntityID(string(entityType), name, node)
	entity := entities.NewEntity(id, name, entityType, ega.currentFile.Path, node)
	entity.SetProperty("type_definition", typeText)
	entity.DocString = ega.extractDocumentation(node)

	return entity
}

// extractInterfaceDefinition extracts interface definitions with method signatures
func (ega *EnhancedGoAnalyzer) extractInterfaceDefinition(node *ts.Node, name string, typeNode *ts.Node) *entities.Entity {
	id := ega.generateEntityID("interface", name, node)
	entity := entities.NewEntity(id, name, entities.EntityTypeInterface, ega.currentFile.Path, node)

	// Extract interface methods
	interfaceInfo := &InterfaceInfo{
		Entity:  entity,
		Methods: make(map[string]*MethodSignature),
	}

	ega.walkNode(typeNode, func(n *ts.Node) {
		if n.Kind() == "method_spec" {
			methodSig := ega.extractInterfaceMethodSignature(n)
			if methodSig != nil {
				interfaceInfo.Methods[methodSig.Name] = methodSig
				entity.SetProperty("method_"+methodSig.Name, methodSig.Parameters+" "+methodSig.ReturnType)
			}
		}
	})

	ega.interfaceRegistry[name] = interfaceInfo
	entity.DocString = ega.extractDocumentation(node)

	return entity
}

// extractStructDefinition extracts struct definitions with field analysis
func (ega *EnhancedGoAnalyzer) extractStructDefinition(node *ts.Node, name string, typeNode *ts.Node) *entities.Entity {
	id := ega.generateEntityID("struct", name, node)
	entity := entities.NewEntity(id, name, entities.EntityTypeStruct, ega.currentFile.Path, node)

	// Extract struct fields
	ega.walkNode(typeNode, func(n *ts.Node) {
		if n.Kind() == "field_declaration" {
			ega.extractStructField(n, entity)
		}
	})

	// Check for embedded structs
	ega.detectStructEmbedding(typeNode, entity)

	entity.DocString = ega.extractDocumentation(node)
	return entity
}

// extractStructField extracts individual struct fields
func (ega *EnhancedGoAnalyzer) extractStructField(node *ts.Node, structEntity *entities.Entity) {
	// Handle different field declaration patterns
	if node.ChildCount() >= 2 {
		var fieldName, fieldType string

		// Try to extract field name and type
		for i := uint(0); i < node.ChildCount(); i++ {
			child := node.Child(i)
			childText := ega.getNodeText(child)

			if child.Kind() == "field_identifier" {
				fieldName = childText
			} else if child.Kind() == "type_identifier" || child.Kind() == "pointer_type" ||
				child.Kind() == "slice_type" || child.Kind() == "map_type" {
				fieldType = childText
			}
		}

		if fieldName != "" {
			structEntity.SetProperty("field_"+fieldName, fieldType)

			// Check if this is an embedded field (anonymous field)
			if fieldName == fieldType {
				structEntity.SetProperty("embedded_"+fieldType, "true")
			}
		}
	}
}

// detectStructEmbedding detects struct embedding relationships
func (ega *EnhancedGoAnalyzer) detectStructEmbedding(typeNode *ts.Node, entity *entities.Entity) {
	ega.walkNode(typeNode, func(n *ts.Node) {
		if n.Kind() == "field_declaration" && n.ChildCount() == 1 {
			// This might be an embedded field
			embeddedType := ega.getNodeText(n.Child(0))
			if embeddedType != "" {
				// Create embedding relationship
				relID := ega.generateRelationshipID("embeds", entity.ID, embeddedType)
				relationship := entities.NewRelationshipByID(
					relID,
					entities.RelationshipTypeEmbeds,
					entity.ID,
					embeddedType,
					entities.EntityTypeStruct, // Source is a struct that embeds another struct
					entities.EntityTypeStruct, // Target is the embedded struct
				)
				ega.relationships = append(ega.relationships, relationship)
			}
		}
	})
}

// buildTypeRegistry builds a registry of all types for relationship analysis
func (ega *EnhancedGoAnalyzer) buildTypeRegistry() {
	// The registry is built during entity extraction
	// This method can be used for post-processing if needed
}

// extractEnhancedRelationships extracts advanced relationships
func (ega *EnhancedGoAnalyzer) extractEnhancedRelationships(node *ts.Node) {
	ega.walkNode(node, func(n *ts.Node) {
		switch n.Kind() {
		case "call_expression":
			ega.extractCallRelationship(n)
		case "selector_expression":
			ega.extractMethodCallRelationship(n)
		case "type_assertion":
			ega.extractTypeAssertionRelationship(n)
		case "composite_literal":
			ega.extractInstantiationRelationship(n)
		}
	})
}

// detectInterfaceImplementations detects when structs implement interfaces
func (ega *EnhancedGoAnalyzer) detectInterfaceImplementations() {
	for interfaceName, interfaceInfo := range ega.interfaceRegistry {
		for structName, methods := range ega.methodRegistry {
			if ega.implementsInterface(methods, interfaceInfo) {
				// Create implements relationship
				structEntity := ega.typeRegistry[structName]
				if structEntity != nil {
					relID := ega.generateRelationshipID("implements", structEntity.ID, interfaceInfo.Entity.ID)
					relationship := entities.NewRelationshipByID(
						relID,
						entities.RelationshipTypeImplements,
						structEntity.ID,
						interfaceInfo.Entity.ID,
						entities.EntityTypeStruct,    // Source is a struct
						entities.EntityTypeInterface, // Target is an interface
					)
					ega.relationships = append(ega.relationships, relationship)
				}
			}
		}
		_ = interfaceName // unused variable fix
	}
}

// implementsInterface checks if a set of methods implements an interface
func (ega *EnhancedGoAnalyzer) implementsInterface(methods []*entities.Entity, interfaceInfo *InterfaceInfo) bool {
	// Create a set of method signatures from the methods
	methodSigs := make(map[string]bool)
	for _, method := range methods {
		methodSigs[method.Name] = true
	}

	// Check if all interface methods are implemented
	for methodName := range interfaceInfo.Methods {
		if !methodSigs[methodName] {
			return false
		}
	}

	return len(interfaceInfo.Methods) > 0 // Interface must have at least one method
}

// Helper methods for enhanced analysis

func (ega *EnhancedGoAnalyzer) extractReceiverType(methodNode *ts.Node) string {
	receiverNode := methodNode.ChildByFieldName("receiver")
	if receiverNode == nil {
		return ""
	}

	receiverText := ega.getNodeText(receiverNode)

	// Extract just the type name from receiver
	re := regexp.MustCompile(`\(\s*\*?\w+\s+\*?(\w+)\s*\)`)
	matches := re.FindStringSubmatch(receiverText)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: try to extract type from simple patterns
	cleaned := strings.Trim(receiverText, "()")
	parts := strings.Fields(cleaned)
	if len(parts) >= 2 {
		typePart := parts[len(parts)-1]
		return strings.TrimPrefix(typePart, "*")
	}

	return ""
}

func (ega *EnhancedGoAnalyzer) isPointerReceiver(receiverNode *ts.Node) string {
	receiverText := ega.getNodeText(receiverNode)
	if strings.Contains(receiverText, "*") {
		return "true"
	}
	return "false"
}

func (ega *EnhancedGoAnalyzer) buildFunctionSignature(node *ts.Node, name string) string {
	signature := name

	parametersNode := node.ChildByFieldName("parameters")
	if parametersNode != nil {
		signature += ega.getNodeText(parametersNode)
	} else {
		signature += "()"
	}

	resultNode := node.ChildByFieldName("result")
	if resultNode != nil {
		signature += " " + ega.getNodeText(resultNode)
	}

	return signature
}

func (ega *EnhancedGoAnalyzer) buildMethodSignature(node *ts.Node, name string) string {
	signature := ""

	receiverNode := node.ChildByFieldName("receiver")
	if receiverNode != nil {
		signature += ega.getNodeText(receiverNode) + " "
	}

	signature += name

	parametersNode := node.ChildByFieldName("parameters")
	if parametersNode != nil {
		signature += ega.getNodeText(parametersNode)
	} else {
		signature += "()"
	}

	resultNode := node.ChildByFieldName("result")
	if resultNode != nil {
		signature += " " + ega.getNodeText(resultNode)
	}

	return signature
}

// Placeholder methods for additional functionality (to be implemented)

func (ega *EnhancedGoAnalyzer) extractEnhancedImports(node *ts.Node) {
	// Enhanced import extraction with alias tracking
	ega.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "import_spec" {
			pathNode := n.ChildByFieldName("path")
			nameNode := n.ChildByFieldName("name")

			if pathNode != nil {
				importPath := strings.Trim(ega.getNodeText(pathNode), "\"")
				alias := ""

				if nameNode != nil {
					alias = ega.getNodeText(nameNode)
					ega.packageImports[alias] = importPath
				}

				id := ega.generateEntityID("import", importPath, n)
				entity := entities.NewEntity(id, importPath, entities.EntityTypeImport, ega.currentFile.Path, n)
				entity.SetProperty("path", importPath)
				if alias != "" {
					entity.SetProperty("alias", alias)
				}

				ega.currentFile.AddEntity(entity)
			}
		}
	})
}

func (ega *EnhancedGoAnalyzer) extractVariableDeclarations(node *ts.Node) {
	// Extract variable and constant declarations
	// Implementation would go here
}

func (ega *EnhancedGoAnalyzer) extractParameterInfo(node *ts.Node, entity *entities.Entity) {
	// Extract detailed parameter information
	// Implementation would go here
}

func (ega *EnhancedGoAnalyzer) extractReturnTypeInfo(node *ts.Node, entity *entities.Entity) {
	// Extract return type information
	// Implementation would go here
}

func (ega *EnhancedGoAnalyzer) extractDocumentation(node *ts.Node) string {
	// Extract Go documentation comments
	return ""
}

func (ega *EnhancedGoAnalyzer) extractInterfaceMethodSignature(node *ts.Node) *MethodSignature {
	// Extract method signature from interface method spec
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	return &MethodSignature{
		Name:       ega.getNodeText(nameNode),
		Parameters: "", // Would extract parameters
		ReturnType: "", // Would extract return type
	}
}

func (ega *EnhancedGoAnalyzer) extractCallRelationship(callNode *ts.Node) {
	// Enhanced call relationship extraction
	// Implementation similar to basic analyzer but with improvements
}

func (ega *EnhancedGoAnalyzer) extractMethodCallRelationship(node *ts.Node) {
	// Extract method call relationships (receiver.method())
}

func (ega *EnhancedGoAnalyzer) extractTypeAssertionRelationship(node *ts.Node) {
	// Extract type assertion relationships
}

func (ega *EnhancedGoAnalyzer) extractInstantiationRelationship(node *ts.Node) {
	// Extract struct/type instantiation relationships
}

// Common helper methods

func (ega *EnhancedGoAnalyzer) getNodeText(node *ts.Node) string {
	if node == nil {
		return ""
	}
	return node.Utf8Text(ega.currentFile.Content)
}

func (ega *EnhancedGoAnalyzer) generateEntityID(entityType, name string, node *ts.Node) string {
	data := fmt.Sprintf("%s:%s:%s:%d:%d",
		ega.currentFile.Path, entityType, name,
		node.StartPosition().Row, node.StartPosition().Column)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

func (ega *EnhancedGoAnalyzer) generateRelationshipID(relType, source, target string) string {
	data := fmt.Sprintf("%s:%s:%s:%s", ega.currentFile.Path, relType, source, target)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

func (ega *EnhancedGoAnalyzer) walkNode(node *ts.Node, visitor func(*ts.Node)) {
	visitor(node)
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		ega.walkNode(child, visitor)
	}
}
