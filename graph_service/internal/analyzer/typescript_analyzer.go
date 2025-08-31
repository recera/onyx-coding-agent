package analyzer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/onyx/onyx-tui/graph_service/internal/entities"

	ts "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

// TypeScriptAnalyzer analyzes TypeScript source code and extracts entities and relationships
type TypeScriptAnalyzer struct {
	parser        *ts.Parser
	language      *ts.Language
	currentFile   *entities.File
	relationships []*entities.Relationship

	// Phase 2: Advanced tracking
	generics   map[string]*TypeScriptGenericInfo
	decorators map[string]*TypeScriptDecoratorInfo
	modules    map[string]*TypeScriptModuleInfo
	components map[string]*ComponentInfo

	// Phase 3: Framework Integration tracking
	endpoints  map[string]*TypeScriptEndpointInfo
	routes     map[string]*TypeScriptRouteInfo
	apiCalls   map[string]*TypeScriptAPICallInfo
	models     map[string]*TypeScriptModelInfo
	middleware map[string]*TypeScriptMiddlewareInfo

	// Test Coverage tracking
	testFramework    string
	testSuites       map[string]*TypeScriptTestSuiteInfo
	testCases        map[string]*TypeScriptTestCaseInfo
	assertions       map[string]*TypeScriptAssertionInfo
	mocks            map[string]*TypeScriptMockInfo
	testFixtures     map[string]*TypeScriptFixtureInfo
	testHooks        map[string]*TypeScriptTestHookInfo
	testCoverage     map[string][]string // test ID -> covered entity IDs
	componentTests   map[string]*TypeScriptComponentTestInfo
}

// TypeScriptGenericInfo represents advanced generic type information
type TypeScriptGenericInfo struct {
	Name        string
	Constraints []string
	TypeParams  []string
	Default     string
	Usage       []*entities.Entity
}

// TypeScriptDecoratorInfo represents decorator usage information
type TypeScriptDecoratorInfo struct {
	Name      string
	Target    string // "class", "method", "property", "parameter"
	Arguments []string
	Factory   bool
	Metadata  map[string]interface{}
}

// TypeScriptModuleInfo represents TypeScript module information
type TypeScriptModuleInfo struct {
	Name            string
	Path            string
	Exports         []string
	TypeOnlyExports []string
	ReExports       []string
	DynamicImports  []string
	IsDeclaration   bool
}

// ComponentInfo represents framework component information
type ComponentInfo struct {
	Name           string   // Component name
	Type           string   // "react", "vue", "angular"
	Props          []string // Component props
	Events         []string // Component events
	Hooks          []string // React hooks used
	Services       []string // Injected services
	JSXElements    []string // JSX elements rendered
	PropsInterface string   // Props interface/type
}

// Phase 3: Framework Integration types

// TypeScriptEndpointInfo represents API endpoint information
type TypeScriptEndpointInfo struct {
	Path         string   // Endpoint path (/api/users)
	Method       string   // HTTP method (GET, POST, etc.)
	Handler      string   // Handler function name
	Middleware   []string // Middleware functions
	Parameters   []string // Route parameters
	RequestType  string   // Request body type
	ResponseType string   // Response type
}

// TypeScriptRouteInfo represents routing information
type TypeScriptRouteInfo struct {
	Path      string   // Route path
	Component string   // Component for this route
	Guards    []string // Route guards
	Resolvers []string // Route resolvers
	Children  []string // Child routes
}

// TypeScriptAPICallInfo represents API call information
type TypeScriptAPICallInfo struct {
	URL        string   // API URL or pattern
	Method     string   // HTTP method
	CallSite   string   // Where the call is made
	Parameters []string // Call parameters
	Response   string   // Expected response type
	Library    string   // Library used (fetch, axios, etc.)
}

// TypeScriptModelInfo represents data model information
type TypeScriptModelInfo struct {
	Name      string   // Model name
	Fields    []string // Model fields
	Relations []string // Model relationships
	Table     string   // Database table name
	ORM       string   // ORM type (TypeORM, Prisma, etc.)
}

// TypeScriptMiddlewareInfo represents middleware information
type TypeScriptMiddlewareInfo struct {
	Name     string   // Middleware name
	Type     string   // Middleware type (auth, logging, etc.)
	Order    int      // Execution order
	Routes   []string // Routes using this middleware
	Function string   // Middleware function
}

// Test Coverage Types

// TypeScriptTestSuiteInfo represents a test suite (describe block)
type TypeScriptTestSuiteInfo struct {
	ID          string
	Name        string
	FilePath    string
	Type        string // "describe", "suite", "context"
	TestCases   []string
	NestedSuites []string
	SetupHooks  []string
	TeardownHooks []string
	StartLine   int
	EndLine     int
}

// TypeScriptTestCaseInfo represents an individual test case
type TypeScriptTestCaseInfo struct {
	ID              string
	Name            string
	SuiteID         string
	Type            string // "it", "test", "specify"
	TestType        string // "unit", "integration", "e2e", "component"
	Async           bool
	Timeout         int
	Skip            bool
	Only            bool // .only modifier
	Assertions      []string
	Mocks           []string
	TestedEntities  []string
	StartLine       int
	EndLine         int
}

// TypeScriptAssertionInfo represents a test assertion
type TypeScriptAssertionInfo struct {
	ID         string
	TestCaseID string
	Type       string // "expect", "assert", "should"
	Method     string // "toBe", "toEqual", "toMatch", etc.
	Expected   string
	Actual     string
	IsNegated  bool   // .not modifier
	Line       int
}

// TypeScriptMockInfo represents a mock or spy
type TypeScriptMockInfo struct {
	ID           string
	TestCaseID   string
	Type         string // "mock", "spy", "stub", "fake"
	Target       string // What is being mocked
	Module       string // Module being mocked
	Method       string // Method being mocked
	Framework    string // "jest", "sinon", "vitest"
	ReturnValue  string
	Implementation string
	Line         int
}

// TypeScriptFixtureInfo represents test data fixtures
type TypeScriptFixtureInfo struct {
	ID         string
	Name       string
	Type       string // "data", "component", "state", "props"
	Scope      string // "test", "suite", "file", "global"
	Data       string
	UsedBy     []string // Test case IDs using this fixture
	Line       int
}

// TypeScriptTestHookInfo represents test lifecycle hooks
type TypeScriptTestHookInfo struct {
	ID         string
	Type       string // "beforeEach", "afterEach", "beforeAll", "afterAll"
	Scope      string // Suite ID or "global"
	Async      bool
	Body       string
	Line       int
}

// TypeScriptComponentTestInfo represents component-specific test information
type TypeScriptComponentTestInfo struct {
	ID              string
	TestCaseID      string
	ComponentName   string
	RenderMethod    string // "render", "shallow", "mount"
	Props           map[string]interface{}
	UserInteractions []string // "click", "type", "hover", etc.
	Queries         []string // "getByRole", "findByText", etc.
	Framework       string // "testing-library", "enzyme"
}

// NewTypeScriptAnalyzer creates a new TypeScript analyzer
func NewTypeScriptAnalyzer() *TypeScriptAnalyzer {
	parser := ts.NewParser()
	language := ts.NewLanguage(tree_sitter_typescript.LanguageTypescript())
	parser.SetLanguage(language)

	return &TypeScriptAnalyzer{
		parser:        parser,
		language:      language,
		relationships: make([]*entities.Relationship, 0),
		generics:      make(map[string]*TypeScriptGenericInfo),
		decorators:    make(map[string]*TypeScriptDecoratorInfo),
		modules:       make(map[string]*TypeScriptModuleInfo),
		components:    make(map[string]*ComponentInfo),
		endpoints:     make(map[string]*TypeScriptEndpointInfo),
		routes:        make(map[string]*TypeScriptRouteInfo),
		apiCalls:      make(map[string]*TypeScriptAPICallInfo),
		models:        make(map[string]*TypeScriptModelInfo),
		middleware:    make(map[string]*TypeScriptMiddlewareInfo),
		// Test coverage maps
		testSuites:     make(map[string]*TypeScriptTestSuiteInfo),
		testCases:      make(map[string]*TypeScriptTestCaseInfo),
		assertions:     make(map[string]*TypeScriptAssertionInfo),
		mocks:          make(map[string]*TypeScriptMockInfo),
		testFixtures:   make(map[string]*TypeScriptFixtureInfo),
		testHooks:      make(map[string]*TypeScriptTestHookInfo),
		testCoverage:   make(map[string][]string),
		componentTests: make(map[string]*TypeScriptComponentTestInfo),
	}
}

// AnalyzeFile analyzes a TypeScript file and returns the File entity with all extracted entities
func (ta *TypeScriptAnalyzer) AnalyzeFile(filePath string, content []byte) (*entities.File, []*entities.Relationship, error) {
	// Parse the file
	tree := ta.parser.ParseCtx(context.Background(), content, nil)
	if tree == nil {
		return nil, nil, fmt.Errorf("failed to parse file %s", filePath)
	}

	// Create File entity
	file := entities.NewFile(filePath, "typescript", tree, content)
	ta.currentFile = file
	ta.relationships = make([]*entities.Relationship, 0)

	// Extract entities from the parse tree
	rootNode := tree.RootNode()
	ta.extractEntities(rootNode, nil)

	// Phase 2: Detect declaration files
	ta.detectDeclarationFiles()

	// Test Coverage: Detect if this is a test file and enhance test entities
	if ta.isTestFile(filePath, content) {
		ta.detectTestFramework(content)
		ta.extractTestSuites(rootNode)
		ta.enhanceTestEntities()
		ta.extractTestRelationships(rootNode)
	}

	// Extract relationships (function calls, imports, inheritance, etc.)
	ta.extractRelationships(rootNode)

	// Phase 2: Build advanced relationships
	ta.buildAdvancedRelationships()

	// Test Coverage: Build test-specific relationships
	if ta.isTestFile(filePath, content) {
		ta.buildTestRelationships()
	}

	return file, ta.relationships, nil
}

// extractEntities recursively extracts entities from the parse tree
func (ta *TypeScriptAnalyzer) extractEntities(node *ts.Node, parent *entities.Entity) {
	nodeType := node.Kind()

	switch nodeType {
	case "function_declaration":
		entity := ta.extractFunction(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
			if parent != nil {
				parent.AddChild(entity)
			}
		}

	case "method_definition":
		entity := ta.extractMethod(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
			if parent != nil {
				parent.AddChild(entity)
			}
		}

	case "class_declaration":
		entity := ta.extractClass(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
			if parent != nil {
				parent.AddChild(entity)
			}
			// Recursively process the class body for methods
			for i := uint(0); i < node.ChildCount(); i++ {
				child := node.Child(i)
				ta.extractEntities(child, entity)
			}
			return
		}

	case "interface_declaration":
		entity := ta.extractInterface(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
			if parent != nil {
				parent.AddChild(entity)
			}
		}

	case "type_alias_declaration":
		entity := ta.extractTypeAlias(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
		}

	case "enum_declaration":
		entity := ta.extractEnum(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
		}

	case "import_statement":
		entity := ta.extractImport(node)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
		}

	case "export_statement":
		entity := ta.extractExport(node)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
		}

	case "variable_declaration":
		entitiesList := ta.extractVariables(node, parent)
		for _, entity := range entitiesList {
			ta.currentFile.AddEntity(entity)
		}

	// Phase 2: Advanced TypeScript features
	case "decorator":
		entity := ta.extractDecorator(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
		}

	case "namespace_declaration":
		entity := ta.extractNamespace(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
			// Process namespace body
			for i := uint(0); i < node.ChildCount(); i++ {
				child := node.Child(i)
				ta.extractEntities(child, entity)
			}
			return
		}

	case "module_declaration":
		entity := ta.extractModuleDeclaration(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
		}

	case "ambient_declaration":
		entity := ta.extractAmbientDeclaration(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
		}

	// Phase 3: Framework Integration features
	case "jsx_element", "jsx_self_closing_element":
		entity := ta.extractJSXElement(node, parent)
		if entity != nil {
			ta.currentFile.AddEntity(entity)
		}

	case "call_expression":
		// Check for API calls, middleware usage, etc.
		ta.analyzeCallExpression(node, parent)
		// Also check for test-related calls (describe, it, test, etc.)
		if ta.currentFile != nil && ta.isTestFile(ta.currentFile.Path, ta.currentFile.Content) {
			ta.extractTestCallExpression(node, parent)
		}
	}

	// Phase 2: Enhanced decorator detection
	ta.extractDecoratorsFromNode(node, parent)

	// Phase 2: Enhanced component detection
	ta.detectFrameworkPatterns(node, parent)

	// Phase 3: Enhanced framework detection
	ta.detectPhase3FrameworkPatterns(node, parent)

	// Phase 2: Enhanced generic analysis
	ta.analyzeAdvancedGenerics(node, parent)

	// Phase 3: API and endpoint analysis
	ta.analyzeAPIPatterns(node, parent)

	// Recursively process child nodes
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		ta.extractEntities(child, parent)
	}
}

// extractFunction extracts a function entity with TypeScript-specific features
func (ta *TypeScriptAnalyzer) extractFunction(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ta.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ta.generateEntityID("function", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeFunction, ta.currentFile.Path, node)

	// Extract type parameters (generics)
	typeParamsNode := node.ChildByFieldName("type_parameters")
	if typeParamsNode != nil {
		entity.SetProperty("type_parameters", ta.getNodeText(typeParamsNode))
		entity.AddSymbol("type_parameters", typeParamsNode)
	}

	// Extract signature (parameters)
	parametersNode := node.ChildByFieldName("parameters")
	if parametersNode != nil {
		entity.Signature = ta.getNodeText(parametersNode)
		ta.extractParameterTypes(parametersNode, entity)
	} else {
		entity.Signature = "()"
	}

	// Extract return type annotation
	returnTypeNode := node.ChildByFieldName("return_type")
	if returnTypeNode != nil {
		entity.SetProperty("return_type", ta.getNodeText(returnTypeNode))
		entity.AddSymbol("return_type", returnTypeNode)
	}

	// Extract body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = ta.getNodeText(bodyNode)
	}

	return entity
}

// extractMethod extracts a method entity from class/interface
func (ta *TypeScriptAnalyzer) extractMethod(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ta.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ta.generateEntityID("method", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeMethod, ta.currentFile.Path, node)

	// Extract type parameters (generics)
	typeParamsNode := node.ChildByFieldName("type_parameters")
	if typeParamsNode != nil {
		entity.SetProperty("type_parameters", ta.getNodeText(typeParamsNode))
	}

	// Extract signature
	parametersNode := node.ChildByFieldName("parameters")
	if parametersNode != nil {
		entity.Signature = ta.getNodeText(parametersNode)
		ta.extractParameterTypes(parametersNode, entity)
	} else {
		entity.Signature = "()"
	}

	// Extract return type
	returnTypeNode := node.ChildByFieldName("return_type")
	if returnTypeNode != nil {
		entity.SetProperty("return_type", ta.getNodeText(returnTypeNode))
	}

	// Extract body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = ta.getNodeText(bodyNode)
	}

	return entity
}

// extractClass extracts a class entity with TypeScript features
func (ta *TypeScriptAnalyzer) extractClass(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ta.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ta.generateEntityID("class", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeClass, ta.currentFile.Path, node)

	// Extract type parameters (generics)
	typeParamsNode := node.ChildByFieldName("type_parameters")
	if typeParamsNode != nil {
		entity.SetProperty("type_parameters", ta.getNodeText(typeParamsNode))
	}

	// Extract heritage clause (extends/implements)
	heritageNode := node.ChildByFieldName("heritage_clause")
	if heritageNode != nil {
		ta.extractHeritageClause(heritageNode, entity)
	}

	// Extract class body to get properties and methods
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = ta.getNodeText(bodyNode)
		ta.extractClassMembers(bodyNode, entity)
	}

	return entity
}

// extractInterface extracts an interface entity
func (ta *TypeScriptAnalyzer) extractInterface(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ta.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ta.generateEntityID("interface", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeInterface, ta.currentFile.Path, node)

	// Extract type parameters (generics)
	typeParamsNode := node.ChildByFieldName("type_parameters")
	if typeParamsNode != nil {
		entity.SetProperty("type_parameters", ta.getNodeText(typeParamsNode))
	}

	// Extract heritage clause (extends)
	heritageNode := node.ChildByFieldName("heritage_clause")
	if heritageNode != nil {
		ta.extractHeritageClause(heritageNode, entity)
	}

	// Extract interface body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = ta.getNodeText(bodyNode)
		ta.extractInterfaceMembers(bodyNode, entity)
	}

	return entity
}

// extractTypeAlias extracts a type alias entity
func (ta *TypeScriptAnalyzer) extractTypeAlias(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ta.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ta.generateEntityID("type", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeType, ta.currentFile.Path, node)

	// Extract type parameters (generics)
	typeParamsNode := node.ChildByFieldName("type_parameters")
	if typeParamsNode != nil {
		entity.SetProperty("type_parameters", ta.getNodeText(typeParamsNode))
	}

	// Extract type definition
	typeNode := node.ChildByFieldName("value")
	if typeNode != nil {
		entity.SetProperty("type_definition", ta.getNodeText(typeNode))
	}

	return entity
}

// extractEnum extracts an enum entity
func (ta *TypeScriptAnalyzer) extractEnum(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ta.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ta.generateEntityID("enum", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeEnum, ta.currentFile.Path, node)

	// Extract enum body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = ta.getNodeText(bodyNode)
		ta.extractEnumMembers(bodyNode, entity)
	}

	return entity
}

// extractImport extracts import statements
func (ta *TypeScriptAnalyzer) extractImport(node *ts.Node) *entities.Entity {
	// Extract import source
	sourceNode := node.ChildByFieldName("source")
	if sourceNode == nil {
		return nil
	}

	source := ta.getNodeText(sourceNode)
	source = strings.Trim(source, "\"'")

	// Generate unique ID
	id := ta.generateEntityID("import", source, node)

	entity := entities.NewEntity(id, source, entities.EntityTypeImport, ta.currentFile.Path, node)

	// Phase 2: Check for type-only imports
	isTypeOnly := ta.isTypeOnlyImport(node)
	entity.SetProperty("type_only", isTypeOnly)

	// Extract import clause (what's being imported)
	clauseNode := node.ChildByFieldName("import_clause")
	if clauseNode != nil {
		importClause := ta.getNodeText(clauseNode)
		entity.SetProperty("imports", importClause)

		// Phase 2: Analyze import patterns
		ta.analyzeImportPatterns(clauseNode, entity, source)
	}

	// Phase 2: Store module info
	moduleInfo := &TypeScriptModuleInfo{
		Name: source,
		Path: ta.currentFile.Path,
	}

	if isTypeOnly {
		moduleInfo.TypeOnlyExports = append(moduleInfo.TypeOnlyExports, source)
	}

	key := fmt.Sprintf("import_%s", source)
	ta.modules[key] = moduleInfo

	return entity
}

// extractExport extracts export statements with Phase 2 enhancements
func (ta *TypeScriptAnalyzer) extractExport(node *ts.Node) *entities.Entity {
	// Try to get the exported name
	var name string
	var exportType string

	// Handle different export patterns
	if declarationNode := node.ChildByFieldName("declaration"); declarationNode != nil {
		if nameNode := declarationNode.ChildByFieldName("name"); nameNode != nil {
			name = ta.getNodeText(nameNode)
			exportType = declarationNode.Kind()
		}
	} else if sourceNode := node.ChildByFieldName("source"); sourceNode != nil {
		name = ta.getNodeText(sourceNode)
		name = strings.Trim(name, "\"'")
		exportType = "re-export"
	}

	if name == "" {
		name = "export"
	}

	// Generate unique ID
	id := ta.generateEntityID("export", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeExport, ta.currentFile.Path, node)
	entity.SetProperty("export_type", exportType)

	// Phase 2: Enhanced export analysis
	ta.enhanceExportAnalysis(node, entity)

	return entity
}

// extractVariables extracts variable declarations
func (ta *TypeScriptAnalyzer) extractVariables(node *ts.Node, parent *entities.Entity) []*entities.Entity {
	var entityList []*entities.Entity

	ta.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "variable_declarator" {
			nameNode := n.ChildByFieldName("name")
			if nameNode != nil {
				name := ta.getNodeText(nameNode)
				if name != "" {
					id := ta.generateEntityID("variable", name, n)
					entity := entities.NewEntity(id, name, entities.EntityTypeVariable, ta.currentFile.Path, n)

					// Extract type annotation
					typeNode := n.ChildByFieldName("type")
					if typeNode != nil {
						entity.SetProperty("type", ta.getNodeText(typeNode))
					}

					// Extract initializer
					initNode := n.ChildByFieldName("value")
					if initNode != nil {
						entity.SetProperty("initializer", ta.getNodeText(initNode))
					}

					entityList = append(entityList, entity)
				}
			}
		}
	})

	return entityList
}

// Helper methods for extraction

// extractParameterTypes extracts parameter type information
func (ta *TypeScriptAnalyzer) extractParameterTypes(parametersNode *ts.Node, entity *entities.Entity) {
	var params []string
	ta.walkNode(parametersNode, func(n *ts.Node) {
		if n.Kind() == "required_parameter" || n.Kind() == "optional_parameter" {
			nameNode := n.ChildByFieldName("pattern")
			typeNode := n.ChildByFieldName("type")

			if nameNode != nil {
				paramName := ta.getNodeText(nameNode)
				if typeNode != nil {
					paramType := ta.getNodeText(typeNode)
					params = append(params, fmt.Sprintf("%s: %s", paramName, paramType))
				} else {
					params = append(params, paramName)
				}
			}
		}
	})

	if len(params) > 0 {
		entity.SetProperty("parameters", strings.Join(params, ", "))
	}
}

// extractHeritageClause extracts extends/implements clauses
func (ta *TypeScriptAnalyzer) extractHeritageClause(heritageNode *ts.Node, entity *entities.Entity) {
	ta.walkNode(heritageNode, func(n *ts.Node) {
		if n.Kind() == "extends_clause" {
			typesNode := n.ChildByFieldName("value")
			if typesNode != nil {
				entity.SetProperty("extends", ta.getNodeText(typesNode))
			}
		} else if n.Kind() == "implements_clause" {
			typesNode := n.ChildByFieldName("value")
			if typesNode != nil {
				entity.SetProperty("implements", ta.getNodeText(typesNode))
			}
		}
	})
}

// extractClassMembers extracts class properties and methods
func (ta *TypeScriptAnalyzer) extractClassMembers(bodyNode *ts.Node, classEntity *entities.Entity) {
	ta.walkNode(bodyNode, func(n *ts.Node) {
		switch n.Kind() {
		case "property_definition":
			if nameNode := n.ChildByFieldName("name"); nameNode != nil {
				propName := ta.getNodeText(nameNode)
				id := ta.generateEntityID("property", propName, n)
				prop := entities.NewEntity(id, propName, entities.EntityTypeProperty, ta.currentFile.Path, n)

				// Extract type
				if typeNode := n.ChildByFieldName("type"); typeNode != nil {
					prop.SetProperty("type", ta.getNodeText(typeNode))
				}

				ta.currentFile.AddEntity(prop)
				classEntity.AddChild(prop)
			}
		}
	})
}

// extractInterfaceMembers extracts interface properties and methods
func (ta *TypeScriptAnalyzer) extractInterfaceMembers(bodyNode *ts.Node, interfaceEntity *entities.Entity) {
	ta.walkNode(bodyNode, func(n *ts.Node) {
		switch n.Kind() {
		case "property_signature":
			if nameNode := n.ChildByFieldName("name"); nameNode != nil {
				propName := ta.getNodeText(nameNode)
				id := ta.generateEntityID("property", propName, n)
				prop := entities.NewEntity(id, propName, entities.EntityTypeProperty, ta.currentFile.Path, n)

				// Extract type
				if typeNode := n.ChildByFieldName("type"); typeNode != nil {
					prop.SetProperty("type", ta.getNodeText(typeNode))
				}

				ta.currentFile.AddEntity(prop)
				interfaceEntity.AddChild(prop)
			}
		case "method_signature":
			if nameNode := n.ChildByFieldName("name"); nameNode != nil {
				methodName := ta.getNodeText(nameNode)
				id := ta.generateEntityID("method", methodName, n)
				method := entities.NewEntity(id, methodName, entities.EntityTypeMethod, ta.currentFile.Path, n)

				// Extract signature
				if paramsNode := n.ChildByFieldName("parameters"); paramsNode != nil {
					method.Signature = ta.getNodeText(paramsNode)
				}

				// Extract return type
				if returnTypeNode := n.ChildByFieldName("return_type"); returnTypeNode != nil {
					method.SetProperty("return_type", ta.getNodeText(returnTypeNode))
				}

				ta.currentFile.AddEntity(method)
				interfaceEntity.AddChild(method)
			}
		}
	})
}

// extractEnumMembers extracts enum values
func (ta *TypeScriptAnalyzer) extractEnumMembers(bodyNode *ts.Node, enumEntity *entities.Entity) {
	var members []string

	ta.walkNode(bodyNode, func(n *ts.Node) {
		if n.Kind() == "property_identifier" || n.Kind() == "enum_assignment" {
			memberText := ta.getNodeText(n)
			if memberText != "" {
				members = append(members, memberText)
			}
		}
	})

	if len(members) > 0 {
		enumEntity.SetProperty("members", strings.Join(members, ", "))
	}
}

// extractRelationships extracts relationships between entities
func (ta *TypeScriptAnalyzer) extractRelationships(node *ts.Node) {
	ta.walkNode(node, func(n *ts.Node) {
		switch n.Kind() {
		case "call_expression":
			ta.extractCallRelationship(n)
		case "class_declaration":
			ta.extractInheritanceRelationships(n)
		case "interface_declaration":
			ta.extractInterfaceRelationships(n)
		}
	})
}

// extractInheritanceRelationships extracts class inheritance relationships
func (ta *TypeScriptAnalyzer) extractInheritanceRelationships(classNode *ts.Node) {
	nameNode := classNode.ChildByFieldName("name")
	if nameNode == nil {
		return
	}

	className := ta.getNodeText(nameNode)
	sourceEntity := ta.findEntityByName(className)
	if sourceEntity == nil {
		return
	}

	heritageNode := classNode.ChildByFieldName("heritage_clause")
	if heritageNode == nil {
		return
	}

	ta.walkNode(heritageNode, func(n *ts.Node) {
		if n.Kind() == "extends_clause" {
			if typesNode := n.ChildByFieldName("value"); typesNode != nil {
				baseClassName := ta.getNodeText(typesNode)
				if baseClassName != "" {
					targetEntity := ta.findEntityByName(baseClassName)
					if targetEntity != nil {
						relID := ta.generateRelationshipID("extends", className, baseClassName)
						rel := entities.NewRelationship(relID, entities.RelationshipTypeInherits, sourceEntity, targetEntity)
						ta.relationships = append(ta.relationships, rel)
					}
				}
			}
		} else if n.Kind() == "implements_clause" {
			if typesNode := n.ChildByFieldName("value"); typesNode != nil {
				interfaceName := ta.getNodeText(typesNode)
				if interfaceName != "" {
					targetEntity := ta.findEntityByName(interfaceName)
					if targetEntity != nil {
						relID := ta.generateRelationshipID("implements", className, interfaceName)
						rel := entities.NewRelationship(relID, entities.RelationshipTypeImplements, sourceEntity, targetEntity)
						ta.relationships = append(ta.relationships, rel)
					}
				}
			}
		}
	})
}

// extractInterfaceRelationships extracts interface inheritance relationships
func (ta *TypeScriptAnalyzer) extractInterfaceRelationships(interfaceNode *ts.Node) {
	nameNode := interfaceNode.ChildByFieldName("name")
	if nameNode == nil {
		return
	}

	interfaceName := ta.getNodeText(nameNode)
	sourceEntity := ta.findEntityByName(interfaceName)
	if sourceEntity == nil {
		return
	}

	heritageNode := interfaceNode.ChildByFieldName("heritage_clause")
	if heritageNode == nil {
		return
	}

	ta.walkNode(heritageNode, func(n *ts.Node) {
		if n.Kind() == "extends_clause" {
			if typesNode := n.ChildByFieldName("value"); typesNode != nil {
				baseInterfaceName := ta.getNodeText(typesNode)
				if baseInterfaceName != "" {
					targetEntity := ta.findEntityByName(baseInterfaceName)
					if targetEntity != nil {
						relID := ta.generateRelationshipID("extends", interfaceName, baseInterfaceName)
						rel := entities.NewRelationship(relID, entities.RelationshipTypeInherits, sourceEntity, targetEntity)
						ta.relationships = append(ta.relationships, rel)
					}
				}
			}
		}
	})
}

// extractCallRelationship extracts function call relationships
func (ta *TypeScriptAnalyzer) extractCallRelationship(callNode *ts.Node) {
	functionNode := callNode.ChildByFieldName("function")
	if functionNode == nil {
		return
	}

	calledFunctionName := ta.getNodeText(functionNode)
	if calledFunctionName == "" {
		return
	}

	// Find the containing function
	containingFunction := ta.findContainingFunction(callNode)
	if containingFunction == nil {
		return
	}

	targetEntity := ta.findEntityByName(calledFunctionName)
	if targetEntity != nil {
		relID := ta.generateRelationshipID("calls", containingFunction.Name, calledFunctionName)
		rel := entities.NewRelationship(relID, entities.RelationshipTypeCalls, containingFunction, targetEntity)
		ta.relationships = append(ta.relationships, rel)
	}
}

// findContainingFunction finds the function that contains the given node
func (ta *TypeScriptAnalyzer) findContainingFunction(node *ts.Node) *entities.Entity {
	current := node.Parent()
	for current != nil {
		switch current.Kind() {
		case "function_declaration", "method_definition", "arrow_function", "function_expression":
			nameNode := current.ChildByFieldName("name")
			if nameNode != nil {
				name := ta.getNodeText(nameNode)
				// Find the entity in our current file
				for _, entity := range ta.currentFile.Entities {
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

// findEntityByName finds an entity by name in the current file
func (ta *TypeScriptAnalyzer) findEntityByName(name string) *entities.Entity {
	for _, entity := range ta.currentFile.Entities {
		if entity.Name == name {
			return entity
		}
	}
	return nil
}

// Utility methods

// getNodeText extracts text content from a tree-sitter node
func (ta *TypeScriptAnalyzer) getNodeText(node *ts.Node) string {
	if node == nil {
		return ""
	}
	return string(ta.currentFile.Content[node.StartByte():node.EndByte()])
}

// generateEntityID generates a unique ID for an entity
func (ta *TypeScriptAnalyzer) generateEntityID(entityType, name string, node *ts.Node) string {
	content := fmt.Sprintf("%s:%s:%s:%d:%d", ta.currentFile.Path, entityType, name, node.StartByte(), node.EndByte())
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:8])
}

// generateRelationshipID generates a unique ID for a relationship
func (ta *TypeScriptAnalyzer) generateRelationshipID(relType, source, target string) string {
	content := fmt.Sprintf("%s:%s:%s:%s", ta.currentFile.Path, relType, source, target)
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:8])
}

// walkNode recursively walks through all child nodes
func (ta *TypeScriptAnalyzer) walkNode(node *ts.Node, visitor func(*ts.Node)) {
	visitor(node)
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		ta.walkNode(child, visitor)
	}
}

// Phase 2: Advanced TypeScript Analysis Methods

// extractDecorator extracts decorator entities
func (ta *TypeScriptAnalyzer) extractDecorator(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ta.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ta.generateEntityID("decorator", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeDecorator, ta.currentFile.Path, node)

	// Extract decorator arguments
	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode != nil {
		entity.SetProperty("arguments", ta.getNodeText(argumentsNode))
	}

	// Store decorator info for relationship building
	decoratorInfo := &TypeScriptDecoratorInfo{
		Name:      name,
		Arguments: []string{},
		Factory:   argumentsNode != nil,
	}

	if argumentsNode != nil {
		ta.walkNode(argumentsNode, func(n *ts.Node) {
			if n.Kind() == "string" || n.Kind() == "number" || n.Kind() == "identifier" {
				decoratorInfo.Arguments = append(decoratorInfo.Arguments, ta.getNodeText(n))
			}
		})
	}

	key := fmt.Sprintf("decorator_%d_%d", node.StartPosition().Row, node.StartPosition().Column)
	ta.decorators[key] = decoratorInfo

	return entity
}

// extractNamespace extracts namespace entities
func (ta *TypeScriptAnalyzer) extractNamespace(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ta.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ta.generateEntityID("namespace", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeNamespace, ta.currentFile.Path, node)

	// Extract namespace body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode != nil {
		entity.Body = ta.getNodeText(bodyNode)
	}

	return entity
}

// extractModuleDeclaration extracts module declaration entities
func (ta *TypeScriptAnalyzer) extractModuleDeclaration(node *ts.Node, parent *entities.Entity) *entities.Entity {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := ta.getNodeText(nameNode)
	if name == "" {
		return nil
	}

	// Generate unique ID
	id := ta.generateEntityID("module", name, node)

	entity := entities.NewEntity(id, name, entities.EntityTypeModule, ta.currentFile.Path, node)

	// Check if it's an ambient module declaration
	isAmbient := ta.isAmbientDeclaration(node)
	entity.SetProperty("ambient", isAmbient)

	// Store module info
	moduleInfo := &TypeScriptModuleInfo{
		Name:          name,
		Path:          ta.currentFile.Path,
		IsDeclaration: isAmbient,
	}

	key := fmt.Sprintf("module_%s", name)
	ta.modules[key] = moduleInfo

	return entity
}

// extractAmbientDeclaration extracts ambient declarations
func (ta *TypeScriptAnalyzer) extractAmbientDeclaration(node *ts.Node, parent *entities.Entity) *entities.Entity {
	// Ambient declarations can contain various types
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		switch child.Kind() {
		case "function_declaration", "class_declaration", "interface_declaration", "variable_declaration":
			return ta.extractEntitiesFromAmbient(child, parent)
		}
	}
	return nil
}

// extractEntitiesFromAmbient extracts entities from ambient declarations
func (ta *TypeScriptAnalyzer) extractEntitiesFromAmbient(node *ts.Node, parent *entities.Entity) *entities.Entity {
	switch node.Kind() {
	case "function_declaration":
		entity := ta.extractFunction(node, parent)
		if entity != nil {
			entity.SetProperty("ambient", true)
		}
		return entity
	case "class_declaration":
		entity := ta.extractClass(node, parent)
		if entity != nil {
			entity.SetProperty("ambient", true)
		}
		return entity
	case "interface_declaration":
		entity := ta.extractInterface(node, parent)
		if entity != nil {
			entity.SetProperty("ambient", true)
		}
		return entity
	case "variable_declaration":
		entities := ta.extractVariables(node, parent)
		for _, entity := range entities {
			entity.SetProperty("ambient", true)
		}
		if len(entities) > 0 {
			return entities[0]
		}
	}
	return nil
}

// extractDecoratorsFromNode extracts decorators that precede a node
func (ta *TypeScriptAnalyzer) extractDecoratorsFromNode(node *ts.Node, parent *entities.Entity) {
	current := node.PrevSibling()
	targetEntity := parent

	// Find the target entity this decorator applies to
	if node.Kind() == "class_declaration" || node.Kind() == "function_declaration" || node.Kind() == "method_definition" {
		if nameNode := node.ChildByFieldName("name"); nameNode != nil {
			name := ta.getNodeText(nameNode)
			targetEntity = ta.findEntityByName(name)
		}
	}

	for current != nil && current.Kind() == "decorator" {
		decoratorEntity := ta.extractDecorator(current, parent)
		if decoratorEntity != nil && targetEntity != nil {
			// Create decorator relationship
			relID := ta.generateRelationshipID("decorates", decoratorEntity.Name, targetEntity.Name)
			rel := entities.NewRelationship(relID, entities.RelationshipTypeDecorates, decoratorEntity, targetEntity)
			ta.relationships = append(ta.relationships, rel)
		}
		current = current.PrevSibling()
	}
}

// detectFrameworkPatterns detects React, Vue, Angular patterns
func (ta *TypeScriptAnalyzer) detectFrameworkPatterns(node *ts.Node, parent *entities.Entity) {
	// Detect React components
	ta.detectReactPatterns(node, parent)

	// Detect Angular patterns
	ta.detectAngularPatterns(node, parent)

	// Detect Vue patterns
	ta.detectVuePatterns(node, parent)
}

// detectReactPatterns detects React component and hook patterns
func (ta *TypeScriptAnalyzer) detectReactPatterns(node *ts.Node, parent *entities.Entity) {
	// Detect React components (classes extending React.Component)
	if node.Kind() == "class_declaration" {
		heritageNode := node.ChildByFieldName("heritage_clause")
		if heritageNode != nil {
			ta.walkNode(heritageNode, func(n *ts.Node) {
				if n.Kind() == "extends_clause" {
					if typesNode := n.ChildByFieldName("value"); typesNode != nil {
						baseClass := ta.getNodeText(typesNode)
						if strings.Contains(baseClass, "React.Component") || strings.Contains(baseClass, "Component") {
							if nameNode := node.ChildByFieldName("name"); nameNode != nil {
								name := ta.getNodeText(nameNode)
								ta.createComponentEntity(name, "react", node)
							}
						}
					}
				}
			})
		}
	}

	// Detect React functional components and hooks
	if node.Kind() == "function_declaration" || node.Kind() == "arrow_function" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			name := ta.getNodeText(nameNode)

			// Check if it's a hook (starts with "use")
			if strings.HasPrefix(name, "use") && len(name) > 3 {
				ta.createHookEntity(name, node)
			}

			// Check if it returns JSX (simplified detection)
			bodyNode := node.ChildByFieldName("body")
			if bodyNode != nil {
				bodyText := ta.getNodeText(bodyNode)
				if strings.Contains(bodyText, "jsx") || strings.Contains(bodyText, "<") {
					ta.createComponentEntity(name, "react", node)
				}
			}
		}
	}
}

// detectAngularPatterns detects Angular component and service patterns
func (ta *TypeScriptAnalyzer) detectAngularPatterns(node *ts.Node, parent *entities.Entity) {
	// Look for Angular decorators
	if node.Kind() == "class_declaration" {
		current := node.PrevSibling()
		for current != nil && current.Kind() == "decorator" {
			decoratorName := ta.getNodeText(current)

			if nameNode := node.ChildByFieldName("name"); nameNode != nil {
				className := ta.getNodeText(nameNode)

				if strings.Contains(decoratorName, "@Component") {
					ta.createComponentEntity(className, "angular", node)
				} else if strings.Contains(decoratorName, "@Injectable") {
					ta.createServiceEntity(className, node)
				}
			}
			current = current.PrevSibling()
		}
	}
}

// detectVuePatterns detects Vue component patterns
func (ta *TypeScriptAnalyzer) detectVuePatterns(node *ts.Node, parent *entities.Entity) {
	// Detect Vue.extend or defineComponent patterns
	if node.Kind() == "call_expression" {
		functionNode := node.ChildByFieldName("function")
		if functionNode != nil {
			functionName := ta.getNodeText(functionNode)
			if strings.Contains(functionName, "Vue.extend") || strings.Contains(functionName, "defineComponent") {
				// This is likely a Vue component
				if parent != nil {
					ta.createComponentEntity(parent.Name, "vue", node)
				}
			}
		}
	}
}

// analyzeAdvancedGenerics analyzes generic type constraints and relationships
func (ta *TypeScriptAnalyzer) analyzeAdvancedGenerics(node *ts.Node, parent *entities.Entity) {
	// Enhanced generic analysis beyond basic type parameters
	ta.walkNode(node, func(n *ts.Node) {
		switch n.Kind() {
		case "type_parameters":
			ta.extractAdvancedTypeParameters(n, parent)
		case "conditional_type":
			ta.extractConditionalType(n, parent)
		case "mapped_type":
			ta.extractMappedType(n, parent)
		case "infer_type":
			ta.extractInferType(n, parent)
		}
	})
}

// extractAdvancedTypeParameters extracts generic type parameters with constraints
func (ta *TypeScriptAnalyzer) extractAdvancedTypeParameters(node *ts.Node, parent *entities.Entity) {
	ta.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "type_parameter" {
			nameNode := n.ChildByFieldName("name")
			if nameNode == nil {
				return
			}

			paramName := ta.getNodeText(nameNode)

			genericInfo := &TypeScriptGenericInfo{
				Name:       paramName,
				TypeParams: []string{paramName},
			}

			// Extract constraint
			constraintNode := n.ChildByFieldName("constraint")
			if constraintNode != nil {
				constraint := ta.getNodeText(constraintNode)
				genericInfo.Constraints = []string{constraint}

				// Create constraint relationship
				if parent != nil {
					relID := ta.generateRelationshipID("constrains", paramName, constraint)
					rel := entities.NewRelationshipByID(relID, entities.RelationshipTypeConstrains, parent.ID, constraint, parent.Type, entities.EntityTypeType)
					ta.relationships = append(ta.relationships, rel)
				}
			}

			// Extract default type
			defaultNode := n.ChildByFieldName("default_type")
			if defaultNode != nil {
				genericInfo.Default = ta.getNodeText(defaultNode)
			}

			key := fmt.Sprintf("generic_%s_%d", paramName, n.StartPosition().Row)
			ta.generics[key] = genericInfo
		}
	})
}

// extractConditionalType extracts conditional type information
func (ta *TypeScriptAnalyzer) extractConditionalType(node *ts.Node, parent *entities.Entity) {
	// Create entity for conditional type
	id := ta.generateEntityID("conditional_type", "conditional", node)
	entity := entities.NewEntity(id, "conditional_type", entities.EntityTypeType, ta.currentFile.Path, node)
	entity.SetProperty("conditional", true)
	entity.Body = ta.getNodeText(node)

	if parent != nil {
		parent.AddChild(entity)
	}
	ta.currentFile.AddEntity(entity)
}

// extractMappedType extracts mapped type information
func (ta *TypeScriptAnalyzer) extractMappedType(node *ts.Node, parent *entities.Entity) {
	// Create entity for mapped type
	id := ta.generateEntityID("mapped_type", "mapped", node)
	entity := entities.NewEntity(id, "mapped_type", entities.EntityTypeType, ta.currentFile.Path, node)
	entity.SetProperty("mapped", true)
	entity.Body = ta.getNodeText(node)

	if parent != nil {
		parent.AddChild(entity)
	}
	ta.currentFile.AddEntity(entity)
}

// extractInferType extracts infer type information
func (ta *TypeScriptAnalyzer) extractInferType(node *ts.Node, parent *entities.Entity) {
	// Extract infer type parameter
	typeNode := node.ChildByFieldName("type_parameter")
	if typeNode != nil {
		paramName := ta.getNodeText(typeNode)

		// Create generic info for inferred type
		genericInfo := &TypeScriptGenericInfo{
			Name:       paramName,
			TypeParams: []string{paramName},
		}

		key := fmt.Sprintf("infer_%s_%d", paramName, node.StartPosition().Row)
		ta.generics[key] = genericInfo
	}
}

// Helper methods for creating specialized entities

// createComponentEntity creates a component entity
func (ta *TypeScriptAnalyzer) createComponentEntity(name, framework string, node *ts.Node) {
	id := ta.generateEntityID("component", name, node)
	entity := entities.NewEntity(id, name, entities.EntityTypeComponent, ta.currentFile.Path, node)
	entity.SetProperty("framework", framework)

	componentInfo := &ComponentInfo{
		Name: name,
		Type: framework,
	}

	// Extract props for React/Vue components
	if framework == "react" || framework == "vue" {
		ta.extractComponentProps(node, componentInfo)
	}

	key := fmt.Sprintf("component_%s", name)
	ta.components[key] = componentInfo

	ta.currentFile.AddEntity(entity)
}

// createHookEntity creates a React hook entity
func (ta *TypeScriptAnalyzer) createHookEntity(name string, node *ts.Node) {
	id := ta.generateEntityID("hook", name, node)
	entity := entities.NewEntity(id, name, entities.EntityTypeHook, ta.currentFile.Path, node)
	entity.SetProperty("type", "react_hook")

	ta.currentFile.AddEntity(entity)
}

// createServiceEntity creates a service entity
func (ta *TypeScriptAnalyzer) createServiceEntity(name string, node *ts.Node) {
	id := ta.generateEntityID("service", name, node)
	entity := entities.NewEntity(id, name, entities.EntityTypeService, ta.currentFile.Path, node)

	ta.currentFile.AddEntity(entity)
}

// extractComponentProps extracts component props
func (ta *TypeScriptAnalyzer) extractComponentProps(node *ts.Node, componentInfo *ComponentInfo) {
	// This is a simplified implementation
	// In a full implementation, this would analyze prop types, interfaces, etc.
	ta.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "property_signature" {
			nameNode := n.ChildByFieldName("name")
			if nameNode != nil {
				propName := ta.getNodeText(nameNode)
				componentInfo.Props = append(componentInfo.Props, propName)
			}
		}
	})
}

// isAmbientDeclaration checks if a node is an ambient declaration
func (ta *TypeScriptAnalyzer) isAmbientDeclaration(node *ts.Node) bool {
	// Check if the node has 'declare' keyword
	current := node.PrevSibling()
	for current != nil {
		if current.Kind() == "declare" {
			return true
		}
		current = current.PrevSibling()
	}
	return false
}

// Phase 2: Advanced Import/Export Analysis Methods

// isTypeOnlyImport checks if an import is type-only
func (ta *TypeScriptAnalyzer) isTypeOnlyImport(node *ts.Node) bool {
	// Check for "import type" syntax
	hasTypeKeyword := false
	ta.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "type" {
			hasTypeKeyword = true
		}
	})

	if hasTypeKeyword {
		return true
	}

	// Check the import clause for type-only patterns
	clauseNode := node.ChildByFieldName("import_clause")
	if clauseNode != nil {
		clauseText := ta.getNodeText(clauseNode)
		if strings.Contains(clauseText, "type ") {
			return true
		}
	}

	return false
}

// analyzeImportPatterns analyzes different import patterns
func (ta *TypeScriptAnalyzer) analyzeImportPatterns(clauseNode *ts.Node, entity *entities.Entity, source string) {
	clauseText := ta.getNodeText(clauseNode)

	// Detect import patterns
	if strings.Contains(clauseText, "* as ") {
		entity.SetProperty("import_type", "namespace")
	} else if strings.Contains(clauseText, "{") {
		entity.SetProperty("import_type", "named")
		ta.extractNamedImports(clauseNode, entity)
	} else {
		entity.SetProperty("import_type", "default")
	}

	// Check for dynamic imports in the file
	ta.checkForDynamicImports(source)
}

// extractNamedImports extracts individual named imports
func (ta *TypeScriptAnalyzer) extractNamedImports(clauseNode *ts.Node, entity *entities.Entity) {
	var namedImports []string

	ta.walkNode(clauseNode, func(n *ts.Node) {
		if n.Kind() == "import_specifier" {
			nameNode := n.ChildByFieldName("name")
			if nameNode != nil {
				namedImports = append(namedImports, ta.getNodeText(nameNode))
			}
		}
	})

	if len(namedImports) > 0 {
		entity.SetProperty("named_imports", strings.Join(namedImports, ", "))
	}
}

// checkForDynamicImports checks for dynamic import() expressions
func (ta *TypeScriptAnalyzer) checkForDynamicImports(source string) {
	// This would be called during relationship extraction to find dynamic imports
	// For now, we'll store this information for later processing
	if moduleInfo, exists := ta.modules[fmt.Sprintf("import_%s", source)]; exists {
		moduleInfo.DynamicImports = append(moduleInfo.DynamicImports, source)
	}
}

// enhanceExportAnalysis enhances export analysis for Phase 2
func (ta *TypeScriptAnalyzer) enhanceExportAnalysis(node *ts.Node, entity *entities.Entity) {
	// Check for re-exports
	if sourceNode := node.ChildByFieldName("source"); sourceNode != nil {
		source := ta.getNodeText(sourceNode)
		source = strings.Trim(source, "\"'")
		entity.SetProperty("re_export_source", source)

		// Create re-export relationship
		relID := ta.generateRelationshipID("re_exports", entity.Name, source)
		rel := entities.NewRelationshipByID(relID, entities.RelationshipTypeReExports, entity.ID, source, entities.EntityTypeExport, entities.EntityTypeModule)
		ta.relationships = append(ta.relationships, rel)

		// Store module info
		if moduleInfo, exists := ta.modules[fmt.Sprintf("export_%s", entity.Name)]; exists {
			moduleInfo.ReExports = append(moduleInfo.ReExports, source)
		} else {
			moduleInfo := &TypeScriptModuleInfo{
				Name:      entity.Name,
				Path:      ta.currentFile.Path,
				ReExports: []string{source},
			}
			ta.modules[fmt.Sprintf("export_%s", entity.Name)] = moduleInfo
		}
	}

	// Check for type-only exports
	if ta.isTypeOnlyExport(node) {
		entity.SetProperty("type_only", true)
	}
}

// isTypeOnlyExport checks if an export is type-only
func (ta *TypeScriptAnalyzer) isTypeOnlyExport(node *ts.Node) bool {
	nodeText := ta.getNodeText(node)
	return strings.Contains(nodeText, "export type")
}

// detectDeclarationFiles detects .d.ts declaration file patterns
func (ta *TypeScriptAnalyzer) detectDeclarationFiles() {
	if strings.HasSuffix(ta.currentFile.Path, ".d.ts") {
		// Mark all entities in this file as ambient/declaration
		for _, entity := range ta.currentFile.Entities {
			entity.SetProperty("declaration", true)
			entity.SetProperty("ambient", true)
		}

		// Create module info for declaration file
		moduleName := strings.TrimSuffix(ta.currentFile.Path, ".d.ts")
		moduleInfo := &TypeScriptModuleInfo{
			Name:          moduleName,
			Path:          ta.currentFile.Path,
			IsDeclaration: true,
		}

		ta.modules[fmt.Sprintf("declaration_%s", moduleName)] = moduleInfo
	}
}

// Phase 2: Advanced Relationship Building

// buildAdvancedRelationships builds Phase 2 and Phase 3 relationships
func (ta *TypeScriptAnalyzer) buildAdvancedRelationships() {
	// Phase 2: Build decorator relationships
	ta.buildDecoratorRelationships()

	// Phase 2: Build generic constraint relationships
	ta.buildGenericConstraintRelationships()

	// Phase 2: Build component relationships
	ta.buildComponentRelationships()

	// Phase 2: Build module relationships
	ta.buildModuleRelationships()

	// Phase 3: Build framework integration relationships
	ta.buildFrameworkRelationships()
	ta.buildAPIRelationships()
	ta.buildJSXRelationships()
}

// buildDecoratorRelationships creates relationships for decorators
func (ta *TypeScriptAnalyzer) buildDecoratorRelationships() {
	for _, decoratorInfo := range ta.decorators {
		decoratorEntity := ta.findEntityByName(decoratorInfo.Name)
		if decoratorEntity == nil {
			continue
		}

		// Find what this decorator decorates
		for _, entity := range ta.currentFile.Entities {
			if ta.entityHasDecorator(entity, decoratorInfo.Name) {
				relID := ta.generateRelationshipID("decorates", decoratorInfo.Name, entity.Name)
				rel := entities.NewRelationship(relID, entities.RelationshipTypeDecorates, decoratorEntity, entity)
				ta.relationships = append(ta.relationships, rel)
			}
		}
	}
}

// buildGenericConstraintRelationships creates constraint relationships
func (ta *TypeScriptAnalyzer) buildGenericConstraintRelationships() {
	for _, genericInfo := range ta.generics {
		for _, constraint := range genericInfo.Constraints {
			constraintEntity := ta.findEntityByName(constraint)
			if constraintEntity != nil {
				relID := ta.generateRelationshipID("constrains", genericInfo.Name, constraint)
				rel := entities.NewRelationshipByID(relID, entities.RelationshipTypeConstrains, genericInfo.Name, constraint, entities.EntityTypeGeneric, entities.EntityTypeType)
				ta.relationships = append(ta.relationships, rel)
			}
		}
	}
}

// buildComponentRelationships creates component-specific relationships
func (ta *TypeScriptAnalyzer) buildComponentRelationships() {
	for _, componentInfo := range ta.components {
		componentEntity := ta.findEntityByName(componentInfo.Name)
		if componentEntity == nil {
			continue
		}

		// Create relationships for props
		for _, prop := range componentInfo.Props {
			propEntity := ta.findEntityByName(prop)
			if propEntity != nil {
				relID := ta.generateRelationshipID("uses", componentInfo.Name, prop)
				rel := entities.NewRelationship(relID, entities.RelationshipTypeUses, componentEntity, propEntity)
				ta.relationships = append(ta.relationships, rel)
			}
		}

		// Create relationships for services (Angular)
		for _, service := range componentInfo.Services {
			serviceEntity := ta.findEntityByName(service)
			if serviceEntity != nil {
				relID := ta.generateRelationshipID("injects", componentInfo.Name, service)
				rel := entities.NewRelationship(relID, entities.RelationshipTypeInjects, componentEntity, serviceEntity)
				ta.relationships = append(ta.relationships, rel)
			}
		}
	}
}

// buildModuleRelationships creates module-specific relationships
func (ta *TypeScriptAnalyzer) buildModuleRelationships() {
	for _, moduleInfo := range ta.modules {
		// Create relationships for re-exports
		for _, reExport := range moduleInfo.ReExports {
			moduleEntity := ta.findEntityByName(moduleInfo.Name)
			targetEntity := ta.findEntityByName(reExport)
			if moduleEntity != nil && targetEntity != nil {
				relID := ta.generateRelationshipID("re_exports", moduleInfo.Name, reExport)
				rel := entities.NewRelationship(relID, entities.RelationshipTypeReExports, moduleEntity, targetEntity)
				ta.relationships = append(ta.relationships, rel)
			}
		}

		// Create relationships for dynamic imports
		for _, dynamicImport := range moduleInfo.DynamicImports {
			moduleEntity := ta.findEntityByName(moduleInfo.Name)
			if moduleEntity != nil {
				relID := ta.generateRelationshipID("dynamic_import", moduleInfo.Name, dynamicImport)
				rel := entities.NewRelationshipByID(relID, entities.RelationshipTypeDynamicImport, moduleEntity.ID, dynamicImport, entities.EntityTypeModule, entities.EntityTypeModule)
				ta.relationships = append(ta.relationships, rel)
			}
		}
	}
}

// Helper methods

// entityHasDecorator checks if an entity has a specific decorator
func (ta *TypeScriptAnalyzer) entityHasDecorator(entity *entities.Entity, decoratorName string) bool {
	decorators, exists := entity.Properties["decorators"]
	if !exists {
		return false
	}

	if decoratorList, ok := decorators.([]string); ok {
		for _, decorator := range decoratorList {
			if strings.Contains(decorator, decoratorName) {
				return true
			}
		}
	}

	return false
}

// Phase 3: Framework Integration Methods

// extractJSXElement extracts JSX elements and components
func (ta *TypeScriptAnalyzer) extractJSXElement(node *ts.Node, parent *entities.Entity) *entities.Entity {
	var elementName string

	// Extract JSX element name
	if node.Kind() == "jsx_element" {
		if openingElement := node.ChildByFieldName("opening_element"); openingElement != nil {
			if nameNode := openingElement.ChildByFieldName("name"); nameNode != nil {
				elementName = ta.getNodeText(nameNode)
			}
		}
	} else if node.Kind() == "jsx_self_closing_element" {
		if nameNode := node.ChildByFieldName("name"); nameNode != nil {
			elementName = ta.getNodeText(nameNode)
		}
	}

	if elementName == "" {
		return nil
	}

	// Generate unique ID
	id := ta.generateEntityID("jsx_element", elementName, node)
	entity := entities.NewEntity(id, elementName, entities.EntityTypeJSXElement, ta.currentFile.Path, node)

	// Extract JSX attributes (props)
	ta.extractJSXAttributes(node, entity)

	// Store JSX element info for component relationships
	if parent != nil && parent.Type == entities.EntityTypeComponent {
		if componentInfo, exists := ta.components[parent.Name]; exists {
			componentInfo.JSXElements = append(componentInfo.JSXElements, elementName)
		}
	}

	return entity
}

// extractJSXAttributes extracts JSX element attributes (props)
func (ta *TypeScriptAnalyzer) extractJSXAttributes(node *ts.Node, jsxEntity *entities.Entity) {
	var attributes []string

	ta.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "jsx_attribute" {
			if nameNode := n.ChildByFieldName("name"); nameNode != nil {
				attrName := ta.getNodeText(nameNode)
				attributes = append(attributes, attrName)

				// Create prop entity
				propID := ta.generateEntityID("prop", attrName, n)
				propEntity := entities.NewEntity(propID, attrName, entities.EntityTypeProp, ta.currentFile.Path, n)

				// Extract prop value
				if valueNode := n.ChildByFieldName("value"); valueNode != nil {
					propEntity.SetProperty("value", ta.getNodeText(valueNode))
				}

				ta.currentFile.AddEntity(propEntity)
				jsxEntity.AddChild(propEntity)
			}
		}
	})

	if len(attributes) > 0 {
		jsxEntity.SetProperty("attributes", strings.Join(attributes, ", "))
	}
}

// analyzeCallExpression analyzes function calls for API patterns and framework usage
func (ta *TypeScriptAnalyzer) analyzeCallExpression(node *ts.Node, parent *entities.Entity) {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return
	}

	functionName := ta.getNodeText(functionNode)

	// Detect API calls
	if ta.isAPICall(functionName) {
		ta.extractAPICall(node, parent, functionName)
	}

	// Detect Express.js patterns
	if ta.isExpressPattern(functionName) {
		ta.extractExpressPattern(node, parent, functionName)
	}

	// Detect middleware usage
	if ta.isMiddlewareCall(functionName) {
		ta.extractMiddlewareUsage(node, parent, functionName)
	}
}

// isAPICall checks if a function call is an API call
func (ta *TypeScriptAnalyzer) isAPICall(functionName string) bool {
	apiPatterns := []string{
		"fetch", "axios", "http", "$.get", "$.post", "$.ajax",
		"this.http", "httpClient", "api.get", "api.post",
	}

	for _, pattern := range apiPatterns {
		if strings.Contains(functionName, pattern) {
			return true
		}
	}

	return false
}

// extractAPICall extracts API call information
func (ta *TypeScriptAnalyzer) extractAPICall(node *ts.Node, parent *entities.Entity, functionName string) {
	// Generate unique ID for API call
	id := ta.generateEntityID("api_call", functionName, node)
	entity := entities.NewEntity(id, functionName, entities.EntityTypeAPICall, ta.currentFile.Path, node)

	// Extract API call details
	apiCallInfo := &TypeScriptAPICallInfo{
		CallSite: ta.currentFile.Path,
		Library:  ta.getAPILibrary(functionName),
	}

	// Extract URL and method from arguments
	ta.extractAPICallArguments(node, apiCallInfo)

	entity.SetProperty("url", apiCallInfo.URL)
	entity.SetProperty("method", apiCallInfo.Method)
	entity.SetProperty("library", apiCallInfo.Library)

	key := fmt.Sprintf("api_call_%d_%d", node.StartPosition().Row, node.StartPosition().Column)
	ta.apiCalls[key] = apiCallInfo

	ta.currentFile.AddEntity(entity)

	// Create relationship if called from a function/component
	if parent != nil {
		relID := ta.generateRelationshipID("calls_api", parent.Name, apiCallInfo.URL)
		rel := entities.NewRelationship(relID, entities.RelationshipTypeCallsAPI, parent, entity)
		ta.relationships = append(ta.relationships, rel)
	}
}

// extractAPICallArguments extracts URL and method from API call arguments
func (ta *TypeScriptAnalyzer) extractAPICallArguments(node *ts.Node, apiCallInfo *TypeScriptAPICallInfo) {
	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode == nil {
		return
	}

	var args []string
	ta.walkNode(argumentsNode, func(n *ts.Node) {
		if n.Kind() == "string" || n.Kind() == "template_string" {
			argValue := ta.getNodeText(n)
			args = append(args, argValue)
		}
	})

	if len(args) > 0 {
		// First argument is usually the URL
		apiCallInfo.URL = strings.Trim(args[0], "\"'`")

		// Try to detect HTTP method
		if strings.Contains(apiCallInfo.URL, "/api/") {
			// Default to GET, but could be enhanced with more analysis
			apiCallInfo.Method = "GET"
		}
	}
}

// getAPILibrary determines which API library is being used
func (ta *TypeScriptAnalyzer) getAPILibrary(functionName string) string {
	if strings.Contains(functionName, "fetch") {
		return "fetch"
	} else if strings.Contains(functionName, "axios") {
		return "axios"
	} else if strings.Contains(functionName, "http") {
		return "http"
	} else if strings.Contains(functionName, "$") {
		return "jquery"
	}
	return "unknown"
}

// isExpressPattern checks if a call is an Express.js pattern
func (ta *TypeScriptAnalyzer) isExpressPattern(functionName string) bool {
	expressPatterns := []string{
		"app.get", "app.post", "app.put", "app.delete", "app.patch",
		"router.get", "router.post", "router.put", "router.delete",
		"express.Router", "app.use",
	}

	for _, pattern := range expressPatterns {
		if strings.Contains(functionName, pattern) {
			return true
		}
	}

	return false
}

// extractExpressPattern extracts Express.js route and endpoint patterns
func (ta *TypeScriptAnalyzer) extractExpressPattern(node *ts.Node, parent *entities.Entity, functionName string) {
	// Extract route information
	if strings.Contains(functionName, ".get") || strings.Contains(functionName, ".post") ||
		strings.Contains(functionName, ".put") || strings.Contains(functionName, ".delete") {
		ta.extractExpressRoute(node, parent, functionName)
	}

	// Extract middleware usage
	if strings.Contains(functionName, "app.use") {
		ta.extractExpressMiddleware(node, parent)
	}
}

// extractExpressRoute extracts Express.js route definitions
func (ta *TypeScriptAnalyzer) extractExpressRoute(node *ts.Node, parent *entities.Entity, functionName string) {
	// Determine HTTP method
	method := "GET"
	if strings.Contains(functionName, ".post") {
		method = "POST"
	} else if strings.Contains(functionName, ".put") {
		method = "PUT"
	} else if strings.Contains(functionName, ".delete") {
		method = "DELETE"
	}

	// Extract route path from arguments
	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode == nil {
		return
	}

	var routePath string
	var handlerName string

	// Get first argument (route path)
	ta.walkNode(argumentsNode, func(n *ts.Node) {
		if n.Kind() == "string" && routePath == "" {
			routePath = strings.Trim(ta.getNodeText(n), "\"'")
		} else if n.Kind() == "identifier" || n.Kind() == "arrow_function" {
			handlerName = ta.getNodeText(n)
		}
	})

	if routePath != "" {
		// Create endpoint entity
		endpointID := ta.generateEntityID("endpoint", fmt.Sprintf("%s:%s", method, routePath), node)
		endpointEntity := entities.NewEntity(endpointID, routePath, entities.EntityTypeEndpoint, ta.currentFile.Path, node)

		endpointEntity.SetProperty("method", method)
		endpointEntity.SetProperty("path", routePath)
		endpointEntity.SetProperty("handler", handlerName)

		// Store endpoint info
		endpointInfo := &TypeScriptEndpointInfo{
			Path:    routePath,
			Method:  method,
			Handler: handlerName,
		}

		key := fmt.Sprintf("endpoint_%s_%s", method, routePath)
		ta.endpoints[key] = endpointInfo

		ta.currentFile.AddEntity(endpointEntity)

		// Create relationship to handler
		if parent != nil {
			relID := ta.generateRelationshipID("exposes_endpoint", parent.Name, routePath)
			rel := entities.NewRelationship(relID, entities.RelationshipTypeExposesEndpoint, parent, endpointEntity)
			ta.relationships = append(ta.relationships, rel)
		}
	}
}

// extractExpressMiddleware extracts Express.js middleware usage
func (ta *TypeScriptAnalyzer) extractExpressMiddleware(node *ts.Node, parent *entities.Entity) {
	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode == nil {
		return
	}

	var middlewareName string
	ta.walkNode(argumentsNode, func(n *ts.Node) {
		if n.Kind() == "identifier" {
			middlewareName = ta.getNodeText(n)
		}
	})

	if middlewareName != "" {
		// Create middleware entity
		middlewareID := ta.generateEntityID("middleware", middlewareName, node)
		middlewareEntity := entities.NewEntity(middlewareID, middlewareName, entities.EntityTypeMiddleware, ta.currentFile.Path, node)

		// Store middleware info
		middlewareInfo := &TypeScriptMiddlewareInfo{
			Name:     middlewareName,
			Function: middlewareName,
		}

		key := fmt.Sprintf("middleware_%s", middlewareName)
		ta.middleware[key] = middlewareInfo

		ta.currentFile.AddEntity(middlewareEntity)
	}
}

// isMiddlewareCall checks if a call is middleware usage
func (ta *TypeScriptAnalyzer) isMiddlewareCall(functionName string) bool {
	middlewarePatterns := []string{
		"use", "middleware", "guard", "interceptor",
	}

	for _, pattern := range middlewarePatterns {
		if strings.Contains(functionName, pattern) {
			return true
		}
	}

	return false
}

// extractMiddlewareUsage extracts middleware usage patterns
func (ta *TypeScriptAnalyzer) extractMiddlewareUsage(node *ts.Node, parent *entities.Entity, functionName string) {
	// This could be enhanced to detect specific middleware patterns
	// For now, we'll treat it as a general middleware usage
	if parent != nil {
		middlewareID := ta.generateEntityID("middleware_usage", functionName, node)
		entity := entities.NewEntity(middlewareID, functionName, entities.EntityTypeMiddleware, ta.currentFile.Path, node)
		ta.currentFile.AddEntity(entity)
	}
}

// detectPhase3FrameworkPatterns detects enhanced framework patterns for Phase 3
func (ta *TypeScriptAnalyzer) detectPhase3FrameworkPatterns(node *ts.Node, parent *entities.Entity) {
	// Enhanced React patterns
	ta.detectEnhancedReactPatterns(node, parent)

	// Enhanced Angular patterns
	ta.detectEnhancedAngularPatterns(node, parent)

	// Enhanced Vue patterns
	ta.detectEnhancedVuePatterns(node, parent)

	// Express.js patterns
	ta.detectExpressPatterns(node, parent)
}

// detectEnhancedReactPatterns detects advanced React patterns
func (ta *TypeScriptAnalyzer) detectEnhancedReactPatterns(node *ts.Node, parent *entities.Entity) {
	// Detect React component props interface
	if node.Kind() == "interface_declaration" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			interfaceName := ta.getNodeText(nameNode)
			if strings.Contains(interfaceName, "Props") {
				ta.linkPropsToComponent(interfaceName, node)
			}
		}
	}

	// Detect React Context usage
	if node.Kind() == "call_expression" {
		functionNode := node.ChildByFieldName("function")
		if functionNode != nil {
			functionName := ta.getNodeText(functionNode)
			if strings.Contains(functionName, "createContext") || strings.Contains(functionName, "useContext") {
				ta.extractReactContext(node, parent)
			}
		}
	}
}

// linkPropsToComponent links props interfaces to components
func (ta *TypeScriptAnalyzer) linkPropsToComponent(propsInterfaceName string, node *ts.Node) {
	// Find component that might use this props interface
	componentName := strings.Replace(propsInterfaceName, "Props", "", 1)

	if componentInfo, exists := ta.components[componentName]; exists {
		componentInfo.PropsInterface = propsInterfaceName

		// Create relationship
		propsEntity := ta.findEntityByName(propsInterfaceName)
		componentEntity := ta.findEntityByName(componentName)

		if propsEntity != nil && componentEntity != nil {
			relID := ta.generateRelationshipID("has_props", componentName, propsInterfaceName)
			rel := entities.NewRelationship(relID, entities.RelationshipTypeHasProps, componentEntity, propsEntity)
			ta.relationships = append(ta.relationships, rel)
		}
	}
}

// extractReactContext extracts React Context usage
func (ta *TypeScriptAnalyzer) extractReactContext(node *ts.Node, parent *entities.Entity) {
	// Create context entity
	contextID := ta.generateEntityID("context", "context", node)
	contextEntity := entities.NewEntity(contextID, "context", entities.EntityTypeService, ta.currentFile.Path, node)
	contextEntity.SetProperty("type", "react_context")

	ta.currentFile.AddEntity(contextEntity)
}

// detectEnhancedAngularPatterns detects advanced Angular patterns
func (ta *TypeScriptAnalyzer) detectEnhancedAngularPatterns(node *ts.Node, parent *entities.Entity) {
	// Detect Angular modules
	if node.Kind() == "class_declaration" {
		current := node.PrevSibling()
		for current != nil && current.Kind() == "decorator" {
			decoratorText := ta.getNodeText(current)
			if strings.Contains(decoratorText, "@NgModule") {
				ta.extractAngularModule(node, current)
			}
			current = current.PrevSibling()
		}
	}

	// Detect Angular routes
	ta.detectAngularRoutes(node, parent)
}

// extractAngularModule extracts Angular module information
func (ta *TypeScriptAnalyzer) extractAngularModule(classNode *ts.Node, decoratorNode *ts.Node) {
	nameNode := classNode.ChildByFieldName("name")
	if nameNode == nil {
		return
	}

	moduleName := ta.getNodeText(nameNode)

	// Create module entity
	moduleID := ta.generateEntityID("angular_module", moduleName, classNode)
	moduleEntity := entities.NewEntity(moduleID, moduleName, entities.EntityTypeModule, ta.currentFile.Path, classNode)
	moduleEntity.SetProperty("framework", "angular")
	moduleEntity.SetProperty("type", "ngmodule")

	ta.currentFile.AddEntity(moduleEntity)
}

// detectAngularRoutes detects Angular routing patterns
func (ta *TypeScriptAnalyzer) detectAngularRoutes(node *ts.Node, parent *entities.Entity) {
	// Look for route configuration objects
	if node.Kind() == "object" {
		nodeText := ta.getNodeText(node)
		if strings.Contains(nodeText, "path") && strings.Contains(nodeText, "component") {
			ta.extractAngularRoute(node, parent)
		}
	}
}

// extractAngularRoute extracts Angular route configuration
func (ta *TypeScriptAnalyzer) extractAngularRoute(node *ts.Node, parent *entities.Entity) {
	routeInfo := &TypeScriptRouteInfo{}

	ta.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "property_assignment" {
			keyNode := n.ChildByFieldName("name")
			valueNode := n.ChildByFieldName("value")

			if keyNode != nil && valueNode != nil {
				key := ta.getNodeText(keyNode)
				value := strings.Trim(ta.getNodeText(valueNode), "\"'")

				switch key {
				case "path":
					routeInfo.Path = value
				case "component":
					routeInfo.Component = value
				}
			}
		}
	})

	if routeInfo.Path != "" {
		// Create route entity
		routeID := ta.generateEntityID("route", routeInfo.Path, node)
		routeEntity := entities.NewEntity(routeID, routeInfo.Path, entities.EntityTypeRoute, ta.currentFile.Path, node)
		routeEntity.SetProperty("component", routeInfo.Component)

		key := fmt.Sprintf("route_%s", routeInfo.Path)
		ta.routes[key] = routeInfo

		ta.currentFile.AddEntity(routeEntity)
	}
}

// detectEnhancedVuePatterns detects advanced Vue patterns
func (ta *TypeScriptAnalyzer) detectEnhancedVuePatterns(node *ts.Node, parent *entities.Entity) {
	// Detect Vue 3 Composition API patterns
	if node.Kind() == "call_expression" {
		functionNode := node.ChildByFieldName("function")
		if functionNode != nil {
			functionName := ta.getNodeText(functionNode)
			if strings.Contains(functionName, "defineComponent") {
				ta.extractVue3Component(node, parent)
			}
		}
	}
}

// extractVue3Component extracts Vue 3 component with Composition API
func (ta *TypeScriptAnalyzer) extractVue3Component(node *ts.Node, parent *entities.Entity) {
	// Create Vue component entity
	componentID := ta.generateEntityID("vue_component", "vue_component", node)
	componentEntity := entities.NewEntity(componentID, "vue_component", entities.EntityTypeComponent, ta.currentFile.Path, node)
	componentEntity.SetProperty("framework", "vue")
	componentEntity.SetProperty("api", "composition")

	ta.currentFile.AddEntity(componentEntity)
}

// detectExpressPatterns detects Express.js application patterns
func (ta *TypeScriptAnalyzer) detectExpressPatterns(node *ts.Node, parent *entities.Entity) {
	// This is handled in analyzeCallExpression for Express route patterns
	// Additional Express-specific patterns could be added here
}

// analyzeAPIPatterns analyzes API communication patterns
func (ta *TypeScriptAnalyzer) analyzeAPIPatterns(node *ts.Node, parent *entities.Entity) {
	// Detect HTTP client imports
	ta.detectHTTPClientImports(node, parent)

	// Detect API service patterns
	ta.detectAPIServicePatterns(node, parent)

	// Detect GraphQL patterns
	ta.detectGraphQLPatterns(node, parent)
}

// detectHTTPClientImports detects HTTP client library imports
func (ta *TypeScriptAnalyzer) detectHTTPClientImports(node *ts.Node, parent *entities.Entity) {
	if node.Kind() == "import_statement" {
		sourceNode := node.ChildByFieldName("source")
		if sourceNode != nil {
			source := strings.Trim(ta.getNodeText(sourceNode), "\"'")
			httpLibraries := []string{"axios", "fetch", "@angular/common/http", "node-fetch"}

			for _, lib := range httpLibraries {
				if strings.Contains(source, lib) {
					// Mark this file as using HTTP client (tracked via entities)
					// Note: HTTP client usage will be tracked through API call entities
					break
				}
			}
		}
	}
}

// detectAPIServicePatterns detects API service class patterns
func (ta *TypeScriptAnalyzer) detectAPIServicePatterns(node *ts.Node, parent *entities.Entity) {
	if node.Kind() == "class_declaration" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			className := ta.getNodeText(nameNode)
			if strings.Contains(className, "Service") || strings.Contains(className, "API") {
				ta.enhanceServiceEntity(className, node)
			}
		}
	}
}

// enhanceServiceEntity enhances service entities with API information
func (ta *TypeScriptAnalyzer) enhanceServiceEntity(serviceName string, node *ts.Node) {
	// Find existing service entity or create new one
	serviceEntity := ta.findEntityByName(serviceName)
	if serviceEntity == nil {
		serviceID := ta.generateEntityID("service", serviceName, node)
		serviceEntity = entities.NewEntity(serviceID, serviceName, entities.EntityTypeService, ta.currentFile.Path, node)
		ta.currentFile.AddEntity(serviceEntity)
	}

	serviceEntity.SetProperty("api_service", true)

	// Analyze service methods for API endpoints
	ta.analyzeServiceMethods(node, serviceEntity)
}

// analyzeServiceMethods analyzes service methods for API endpoints
func (ta *TypeScriptAnalyzer) analyzeServiceMethods(classNode *ts.Node, serviceEntity *entities.Entity) {
	ta.walkNode(classNode, func(n *ts.Node) {
		if n.Kind() == "method_definition" {
			methodNameNode := n.ChildByFieldName("name")
			if methodNameNode != nil {
				methodName := ta.getNodeText(methodNameNode)

				// Check if method makes API calls
				bodyNode := n.ChildByFieldName("body")
				if bodyNode != nil {
					bodyText := ta.getNodeText(bodyNode)
					if strings.Contains(bodyText, "http") || strings.Contains(bodyText, "fetch") || strings.Contains(bodyText, "axios") {
						// This method makes API calls
						serviceEntity.SetProperty(fmt.Sprintf("method_%s_api", methodName), true)
					}
				}
			}
		}
	})
}

// detectGraphQLPatterns detects GraphQL usage patterns
func (ta *TypeScriptAnalyzer) detectGraphQLPatterns(node *ts.Node, parent *entities.Entity) {
	if node.Kind() == "template_string" {
		templateText := ta.getNodeText(node)
		if strings.Contains(templateText, "query") || strings.Contains(templateText, "mutation") || strings.Contains(templateText, "subscription") {
			ta.extractGraphQLQuery(node, parent)
		}
	}
}

// extractGraphQLQuery extracts GraphQL query information
func (ta *TypeScriptAnalyzer) extractGraphQLQuery(node *ts.Node, parent *entities.Entity) {
	queryText := ta.getNodeText(node)

	// Create GraphQL query entity
	queryID := ta.generateEntityID("graphql_query", "query", node)
	queryEntity := entities.NewEntity(queryID, "graphql_query", entities.EntityTypeAPICall, ta.currentFile.Path, node)
	queryEntity.SetProperty("type", "graphql")
	queryEntity.SetProperty("query", queryText)

	ta.currentFile.AddEntity(queryEntity)
}

// Phase 3: Relationship Building Methods

// buildFrameworkRelationships builds framework-specific relationships
func (ta *TypeScriptAnalyzer) buildFrameworkRelationships() {
	// Build component-prop relationships
	for _, componentInfo := range ta.components {
		componentEntity := ta.findEntityByName(componentInfo.Name)
		if componentEntity == nil {
			continue
		}

		// Link component to its props interface
		if componentInfo.PropsInterface != "" {
			propsEntity := ta.findEntityByName(componentInfo.PropsInterface)
			if propsEntity != nil {
				relID := ta.generateRelationshipID("has_props", componentInfo.Name, componentInfo.PropsInterface)
				rel := entities.NewRelationship(relID, entities.RelationshipTypeHasProps, componentEntity, propsEntity)
				ta.relationships = append(ta.relationships, rel)
			}
		}

		// Link component to JSX elements it renders
		for _, jsxElement := range componentInfo.JSXElements {
			jsxEntity := ta.findEntityByName(jsxElement)
			if jsxEntity != nil {
				relID := ta.generateRelationshipID("renders_jsx", componentInfo.Name, jsxElement)
				rel := entities.NewRelationship(relID, entities.RelationshipTypeRendersJSX, componentEntity, jsxEntity)
				ta.relationships = append(ta.relationships, rel)
			}
		}

		// Link component to services it consumes
		for _, service := range componentInfo.Services {
			serviceEntity := ta.findEntityByName(service)
			if serviceEntity != nil {
				relID := ta.generateRelationshipID("consumes_service", componentInfo.Name, service)
				rel := entities.NewRelationship(relID, entities.RelationshipTypeConsumesService, componentEntity, serviceEntity)
				ta.relationships = append(ta.relationships, rel)
			}
		}
	}

	// Build route-component relationships
	for _, routeInfo := range ta.routes {
		routeEntity := ta.findEntityByName(routeInfo.Path)
		componentEntity := ta.findEntityByName(routeInfo.Component)

		if routeEntity != nil && componentEntity != nil {
			relID := ta.generateRelationshipID("handles_route", routeInfo.Component, routeInfo.Path)
			rel := entities.NewRelationship(relID, entities.RelationshipTypeHandlesRoute, componentEntity, routeEntity)
			ta.relationships = append(ta.relationships, rel)
		}
	}
}

// buildAPIRelationships builds API communication relationships
func (ta *TypeScriptAnalyzer) buildAPIRelationships() {
	// For Phase 3, we focus on detecting API calls and their patterns
	// External endpoint linking would be handled in Phase 4 cross-language integration

	// Build relationships between API call entities and the functions/services that make them
	for _, entity := range ta.currentFile.Entities {
		if entity.Type == entities.EntityTypeAPICall {
			// Find the function or service that contains this API call
			for _, potentialCaller := range ta.currentFile.Entities {
				if (potentialCaller.Type == entities.EntityTypeFunction ||
					potentialCaller.Type == entities.EntityTypeService ||
					strings.Contains(potentialCaller.Name, "Service")) &&
					ta.isEntityContainedIn(entity, potentialCaller) {

					relID := ta.generateRelationshipID("calls_api", potentialCaller.Name, entity.Name)
					rel := entities.NewRelationship(relID, entities.RelationshipTypeCallsAPI, potentialCaller, entity)
					ta.relationships = append(ta.relationships, rel)
				}
			}
		}
	}

	// Build endpoint-handler relationships
	for _, endpointInfo := range ta.endpoints {
		endpointEntity := ta.findEntityByName(endpointInfo.Path)
		handlerEntity := ta.findEntityByName(endpointInfo.Handler)

		if endpointEntity != nil && handlerEntity != nil {
			relID := ta.generateRelationshipID("handles_endpoint", endpointInfo.Handler, endpointInfo.Path)
			rel := entities.NewRelationship(relID, entities.RelationshipTypeHandlesRoute, handlerEntity, endpointEntity)
			ta.relationships = append(ta.relationships, rel)
		}

		// Build middleware relationships
		for _, middlewareName := range endpointInfo.Middleware {
			middlewareEntity := ta.findEntityByName(middlewareName)
			if endpointEntity != nil && middlewareEntity != nil {
				relID := ta.generateRelationshipID("uses_middleware", endpointInfo.Path, middlewareName)
				rel := entities.NewRelationship(relID, entities.RelationshipTypeUsesMiddleware, endpointEntity, middlewareEntity)
				ta.relationships = append(ta.relationships, rel)
			}
		}
	}
}

// buildJSXRelationships builds JSX-specific relationships
func (ta *TypeScriptAnalyzer) buildJSXRelationships() {
	// Find JSX elements and their parent components
	for _, entity := range ta.currentFile.Entities {
		if entity.Type == entities.EntityTypeJSXElement {
			// Find the component that renders this JSX element
			for _, componentEntity := range ta.currentFile.Entities {
				if componentEntity.Type == entities.EntityTypeComponent {
					// Check if this component renders the JSX element
					if componentInfo, exists := ta.components[componentEntity.Name]; exists {
						for _, jsxElement := range componentInfo.JSXElements {
							if jsxElement == entity.Name {
								relID := ta.generateRelationshipID("renders_jsx", componentEntity.Name, entity.Name)
								rel := entities.NewRelationship(relID, entities.RelationshipTypeRendersJSX, componentEntity, entity)
								ta.relationships = append(ta.relationships, rel)
							}
						}
					}
				}
			}

			// Build prop relationships
			for _, childEntity := range entity.Children {
				if childEntity.Type == entities.EntityTypeProp {
					relID := ta.generateRelationshipID("accepts_props", entity.Name, childEntity.Name)
					rel := entities.NewRelationship(relID, entities.RelationshipTypeAcceptsProps, entity, childEntity)
					ta.relationships = append(ta.relationships, rel)
				}
			}
		}
	}
}

// Helper method to check if one entity is contained within another
func (ta *TypeScriptAnalyzer) isEntityContainedIn(child, parent *entities.Entity) bool {
	if child.Node == nil || parent.Node == nil {
		return false
	}

	// Check if child's byte range is within parent's byte range
	return child.StartByte >= parent.StartByte && child.EndByte <= parent.EndByte
}

// ============================================================================
// Test Coverage Implementation
// ============================================================================

// isTestFile determines if a file is a test file based on name patterns and content
func (ta *TypeScriptAnalyzer) isTestFile(filePath string, content []byte) bool {
	if filePath == "" {
		return false
	}

	// Check file name patterns
	testPatterns := []string{
		".test.ts", ".test.tsx", ".test.js", ".test.jsx",
		".spec.ts", ".spec.tsx", ".spec.js", ".spec.jsx",
		"__tests__/", "__test__/", "tests/", "test/",
	}

	for _, pattern := range testPatterns {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}

	// Check content for test patterns if file name doesn't match
	contentStr := string(content)
	testIndicators := []string{
		"describe(", "describe.only(", "describe.skip(",
		"it(", "it.only(", "it.skip(",
		"test(", "test.only(", "test.skip(",
		"suite(", "suite.only(", "suite.skip(",
		"@jest/globals", "from 'jest'", "from \"jest\"",
		"from 'mocha'", "from \"mocha\"",
		"from 'vitest'", "from \"vitest\"",
		"from 'jasmine'", "from \"jasmine\"",
		"@testing-library",
	}

	for _, indicator := range testIndicators {
		if strings.Contains(contentStr, indicator) {
			return true
		}
	}

	return false
}

// detectTestFramework determines which testing framework is being used
func (ta *TypeScriptAnalyzer) detectTestFramework(content []byte) {
	contentStr := string(content)

	// Check for Jest
	jestIndicators := []string{
		"from '@jest/globals'", "from 'jest'",
		"jest.fn()", "jest.mock(", "jest.spyOn(",
		"expect.assertions(", "expect.hasAssertions(",
		"toMatchSnapshot()", "toMatchInlineSnapshot(",
	}
	for _, indicator := range jestIndicators {
		if strings.Contains(contentStr, indicator) {
			ta.testFramework = "jest"
			return
		}
	}

	// Check for Vitest
	vitestIndicators := []string{
		"from 'vitest'", "from \"vitest\"",
		"vi.fn()", "vi.mock(", "vi.spyOn(",
		"import.meta.vitest",
	}
	for _, indicator := range vitestIndicators {
		if strings.Contains(contentStr, indicator) {
			ta.testFramework = "vitest"
			return
		}
	}

	// Check for Mocha
	mochaIndicators := []string{
		"from 'mocha'", "from \"mocha\"",
		"suite(", "suite.only(", "suite.skip(",
		"setup(", "teardown(", "suiteSetup(", "suiteTeardown(",
	}
	for _, indicator := range mochaIndicators {
		if strings.Contains(contentStr, indicator) {
			ta.testFramework = "mocha"
			return
		}
	}

	// Check for Jasmine
	jasmineIndicators := []string{
		"from 'jasmine'", "from \"jasmine\"",
		"jasmine.createSpy(", "jasmine.createSpyObj(",
		"spyOn(", "spyOnProperty(",
	}
	for _, indicator := range jasmineIndicators {
		if strings.Contains(contentStr, indicator) {
			ta.testFramework = "jasmine"
			return
		}
	}

	// Check for Testing Library
	if strings.Contains(contentStr, "@testing-library") {
		if strings.Contains(contentStr, "@testing-library/react") {
			ta.testFramework = "testing-library-react"
		} else {
			ta.testFramework = "testing-library"
		}
		return
	}

	// Default to jest if we see common test patterns
	if strings.Contains(contentStr, "describe(") || strings.Contains(contentStr, "it(") || strings.Contains(contentStr, "test(") {
		ta.testFramework = "jest"
		return
	}

	ta.testFramework = "unknown"
}

// extractTestSuites extracts test suites (describe blocks) from the AST
func (ta *TypeScriptAnalyzer) extractTestSuites(node *ts.Node) {
	ta.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "call_expression" {
			ta.extractTestSuiteFromCall(n, nil)
		}
	})
}

// extractTestSuiteFromCall extracts a test suite from a call expression
func (ta *TypeScriptAnalyzer) extractTestSuiteFromCall(node *ts.Node, parentSuite *TypeScriptTestSuiteInfo) *TypeScriptTestSuiteInfo {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	functionName := ta.getNodeText(functionNode)
	
	// Check if this is a describe/suite call
	if !ta.isTestSuiteCall(functionName) {
		return nil
	}

	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode == nil {
		return nil
	}

	// Get suite name from first argument
	var suiteName string
	for i := uint(0); i < argumentsNode.ChildCount(); i++ {
		arg := argumentsNode.Child(i)
		if arg.Kind() == "string" || arg.Kind() == "template_string" {
			suiteName = ta.getNodeText(arg)
			break
		}
	}

	if suiteName == "" {
		return nil
	}

	// Create suite info
	suiteID := ta.generateTestID("suite", suiteName, int(node.StartByte()))
	suite := &TypeScriptTestSuiteInfo{
		ID:        suiteID,
		Name:      suiteName,
		FilePath:  ta.currentFile.Path,
		Type:      ta.normalizeSuiteType(functionName),
		TestCases: []string{},
		NestedSuites: []string{},
		SetupHooks: []string{},
		TeardownHooks: []string{},
		StartLine: int(node.StartPosition().Row),
		EndLine:   int(node.EndPosition().Row),
	}

	ta.testSuites[suiteID] = suite

	// If this is a nested suite, add it to parent
	if parentSuite != nil {
		parentSuite.NestedSuites = append(parentSuite.NestedSuites, suiteID)
	}

	// Extract test cases and nested suites from the callback function
	for i := uint(0); i < argumentsNode.ChildCount(); i++ {
		arg := argumentsNode.Child(i)
		if arg.Kind() == "arrow_function" || arg.Kind() == "function" {
			bodyNode := arg.ChildByFieldName("body")
			if bodyNode != nil {
				ta.extractTestsFromBlock(bodyNode, suite)
			}
		}
	}

	return suite
}

// extractTestsFromBlock extracts test cases and nested suites from a block
func (ta *TypeScriptAnalyzer) extractTestsFromBlock(node *ts.Node, suite *TypeScriptTestSuiteInfo) {
	ta.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "call_expression" {
			functionNode := n.ChildByFieldName("function")
			if functionNode != nil {
				functionName := ta.getNodeText(functionNode)
				
				// Check for nested describe/suite
				if ta.isTestSuiteCall(functionName) {
					nestedSuite := ta.extractTestSuiteFromCall(n, suite)
					if nestedSuite != nil && suite != nil {
						suite.NestedSuites = append(suite.NestedSuites, nestedSuite.ID)
					}
				}
				
				// Check for test case (it/test)
				if ta.isTestCaseCall(functionName) {
					testCase := ta.extractTestCaseFromCall(n, suite)
					if testCase != nil && suite != nil {
						suite.TestCases = append(suite.TestCases, testCase.ID)
					}
				}
				
				// Check for hooks
				if ta.isTestHookCall(functionName) {
					hook := ta.extractTestHookFromCall(n, suite)
					if hook != nil && suite != nil {
						if strings.Contains(hook.Type, "before") {
							suite.SetupHooks = append(suite.SetupHooks, hook.ID)
						} else {
							suite.TeardownHooks = append(suite.TeardownHooks, hook.ID)
						}
					}
				}
			}
		}
	})
}

// extractTestCaseFromCall extracts a test case from a call expression
func (ta *TypeScriptAnalyzer) extractTestCaseFromCall(node *ts.Node, suite *TypeScriptTestSuiteInfo) *TypeScriptTestCaseInfo {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	functionName := ta.getNodeText(functionNode)
	if !ta.isTestCaseCall(functionName) {
		return nil
	}

	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode == nil {
		return nil
	}

	// Get test name from first argument
	var testName string
	for i := uint(0); i < argumentsNode.ChildCount(); i++ {
		arg := argumentsNode.Child(i)
		if arg.Kind() == "string" || arg.Kind() == "template_string" {
			testName = ta.getNodeText(arg)
			break
		}
	}

	if testName == "" {
		return nil
	}

	// Create test case info
	testID := ta.generateTestID("test", testName, int(node.StartByte()))
	testCase := &TypeScriptTestCaseInfo{
		ID:       testID,
		Name:     testName,
		SuiteID:  "",
		Type:     ta.normalizeTestType(functionName),
		TestType: ta.determineTestType(testName),
		Async:    ta.isAsyncTest(node),
		Skip:     strings.Contains(functionName, ".skip"),
		Only:     strings.Contains(functionName, ".only"),
		Assertions: []string{},
		Mocks:      []string{},
		TestedEntities: []string{},
		StartLine: int(node.StartPosition().Row),
		EndLine:   int(node.EndPosition().Row),
	}

	if suite != nil {
		testCase.SuiteID = suite.ID
	}

	ta.testCases[testID] = testCase

	// Extract assertions and mocks from test body
	for i := uint(0); i < argumentsNode.ChildCount(); i++ {
		arg := argumentsNode.Child(i)
		if arg.Kind() == "arrow_function" || arg.Kind() == "function" {
			bodyNode := arg.ChildByFieldName("body")
			if bodyNode != nil {
				ta.extractTestCaseDetails(bodyNode, testCase)
			}
		}
	}

	return testCase
}

// extractTestCaseDetails extracts assertions, mocks, and tested entities from test body
func (ta *TypeScriptAnalyzer) extractTestCaseDetails(node *ts.Node, testCase *TypeScriptTestCaseInfo) {
	ta.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "call_expression" {
			functionNode := n.ChildByFieldName("function")
			if functionNode != nil {
				// Check for assertions
				if ta.isAssertionCall(n) {
					assertion := ta.extractAssertion(n, testCase)
					if assertion != nil {
						testCase.Assertions = append(testCase.Assertions, assertion.ID)
					}
				}
				
				// Check for mocks
				if ta.isMockCall(n) {
					mock := ta.extractMock(n, testCase)
					if mock != nil {
						testCase.Mocks = append(testCase.Mocks, mock.ID)
					}
				}
				
				// Check for component renders (React testing)
				if ta.isComponentRender(n) {
					componentTest := ta.extractComponentTest(n, testCase)
					if componentTest != nil {
						ta.componentTests[componentTest.ID] = componentTest
					}
				}
				
				// Track function calls to determine what's being tested
				calledFunction := ta.extractCalledFunction(n)
				if calledFunction != "" && !ta.isTestUtilityCall(calledFunction) {
					testCase.TestedEntities = append(testCase.TestedEntities, calledFunction)
				}
			}
		}
	})
}

// extractAssertion extracts assertion information from a call expression
func (ta *TypeScriptAnalyzer) extractAssertion(node *ts.Node, testCase *TypeScriptTestCaseInfo) *TypeScriptAssertionInfo {
	assertionID := ta.generateTestID("assertion", "", int(node.StartByte()))
	
	assertion := &TypeScriptAssertionInfo{
		ID:         assertionID,
		TestCaseID: testCase.ID,
		Line:       int(node.StartPosition().Row),
	}

	// Analyze the assertion pattern
	functionNode := node.ChildByFieldName("function")
	if functionNode != nil {
		if functionNode.Kind() == "member_expression" {
			objectNode := functionNode.ChildByFieldName("object")
			propertyNode := functionNode.ChildByFieldName("property")
			
			if objectNode != nil {
				objectText := ta.getNodeText(objectNode)
				if strings.HasPrefix(objectText, "expect") {
					assertion.Type = "expect"
					
					// Check for .not modifier
					if strings.Contains(objectText, ".not") {
						assertion.IsNegated = true
					}
				} else if strings.HasPrefix(objectText, "assert") {
					assertion.Type = "assert"
				}
			}
			
			if propertyNode != nil {
				assertion.Method = ta.getNodeText(propertyNode)
			}
		} else {
			functionText := ta.getNodeText(functionNode)
			if strings.HasPrefix(functionText, "expect") {
				assertion.Type = "expect"
			} else if strings.HasPrefix(functionText, "assert") {
				assertion.Type = "assert"
			}
		}
	}

	// Extract expected and actual values
	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode != nil && argumentsNode.ChildCount() > 0 {
		assertion.Expected = ta.getNodeText(argumentsNode.Child(0))
		if argumentsNode.ChildCount() > 1 {
			assertion.Actual = ta.getNodeText(argumentsNode.Child(1))
		}
	}

	ta.assertions[assertionID] = assertion
	return assertion
}

// extractMock extracts mock information from a call expression
func (ta *TypeScriptAnalyzer) extractMock(node *ts.Node, testCase *TypeScriptTestCaseInfo) *TypeScriptMockInfo {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	functionText := ta.getNodeText(functionNode)
	mockID := ta.generateTestID("mock", functionText, int(node.StartByte()))
	
	mock := &TypeScriptMockInfo{
		ID:         mockID,
		TestCaseID: testCase.ID,
		Line:       int(node.StartPosition().Row),
		Framework:  ta.testFramework,
	}

	// Determine mock type and target
	if strings.Contains(functionText, "jest.fn") || strings.Contains(functionText, "vi.fn") {
		mock.Type = "function"
	} else if strings.Contains(functionText, "jest.mock") || strings.Contains(functionText, "vi.mock") {
		mock.Type = "module"
		// Extract module name from arguments
		argumentsNode := node.ChildByFieldName("arguments")
		if argumentsNode != nil && argumentsNode.ChildCount() > 0 {
			mock.Module = ta.getNodeText(argumentsNode.Child(0))
		}
	} else if strings.Contains(functionText, "spyOn") || strings.Contains(functionText, "jest.spyOn") || strings.Contains(functionText, "vi.spyOn") {
		mock.Type = "spy"
		// Extract target object and method
		argumentsNode := node.ChildByFieldName("arguments")
		if argumentsNode != nil && argumentsNode.ChildCount() > 1 {
			mock.Target = ta.getNodeText(argumentsNode.Child(0))
			mock.Method = ta.getNodeText(argumentsNode.Child(1))
		}
	} else if strings.Contains(functionText, "sinon.stub") {
		mock.Type = "stub"
		mock.Framework = "sinon"
	} else if strings.Contains(functionText, "sinon.spy") {
		mock.Type = "spy"
		mock.Framework = "sinon"
	}

	ta.mocks[mockID] = mock
	return mock
}

// extractComponentTest extracts component test information (React Testing Library, etc.)
func (ta *TypeScriptAnalyzer) extractComponentTest(node *ts.Node, testCase *TypeScriptTestCaseInfo) *TypeScriptComponentTestInfo {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	functionText := ta.getNodeText(functionNode)
	if !strings.Contains(functionText, "render") && !strings.Contains(functionText, "mount") && !strings.Contains(functionText, "shallow") {
		return nil
	}

	testID := ta.generateTestID("component_test", "", int(node.StartByte()))
	componentTest := &TypeScriptComponentTestInfo{
		ID:         testID,
		TestCaseID: testCase.ID,
		Props:      make(map[string]interface{}),
		UserInteractions: []string{},
		Queries:    []string{},
	}

	// Determine render method
	if strings.Contains(functionText, "render") {
		componentTest.RenderMethod = "render"
		componentTest.Framework = "testing-library"
	} else if strings.Contains(functionText, "mount") {
		componentTest.RenderMethod = "mount"
		componentTest.Framework = "enzyme"
	} else if strings.Contains(functionText, "shallow") {
		componentTest.RenderMethod = "shallow"
		componentTest.Framework = "enzyme"
	}

	// Extract component name and props
	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode != nil && argumentsNode.ChildCount() > 0 {
		firstArg := argumentsNode.Child(0)
		if firstArg.Kind() == "jsx_element" || firstArg.Kind() == "jsx_self_closing_element" {
			// Extract component name
			openingElement := firstArg.ChildByFieldName("opening_element")
			if openingElement == nil {
				openingElement = firstArg // For self-closing elements
			}
			if openingElement != nil {
				nameNode := openingElement.ChildByFieldName("name")
				if nameNode != nil {
					componentTest.ComponentName = ta.getNodeText(nameNode)
				}
			}
			
			// Extract props
			ta.extractJSXProps(firstArg, componentTest)
		}
	}

	return componentTest
}

// extractJSXProps extracts props from a JSX element
func (ta *TypeScriptAnalyzer) extractJSXProps(node *ts.Node, componentTest *TypeScriptComponentTestInfo) {
	ta.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "jsx_attribute" {
			nameNode := n.ChildByFieldName("name")
			valueNode := n.ChildByFieldName("value")
			
			if nameNode != nil && valueNode != nil {
				propName := ta.getNodeText(nameNode)
				propValue := ta.getNodeText(valueNode)
				componentTest.Props[propName] = propValue
			}
		}
	})
}

// extractTestHookFromCall extracts test hook (beforeEach, afterEach, etc.)
func (ta *TypeScriptAnalyzer) extractTestHookFromCall(node *ts.Node, suite *TypeScriptTestSuiteInfo) *TypeScriptTestHookInfo {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	functionName := ta.getNodeText(functionNode)
	if !ta.isTestHookCall(functionName) {
		return nil
	}

	hookID := ta.generateTestID("hook", functionName, int(node.StartByte()))
	hook := &TypeScriptTestHookInfo{
		ID:    hookID,
		Type:  functionName,
		Scope: "global",
		Async: ta.isAsyncTest(node),
		Line:  int(node.StartPosition().Row),
	}

	if suite != nil {
		hook.Scope = suite.ID
	}

	// Extract hook body
	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode != nil {
		for i := uint(0); i < argumentsNode.ChildCount(); i++ {
			arg := argumentsNode.Child(i)
			if arg.Kind() == "arrow_function" || arg.Kind() == "function" {
				bodyNode := arg.ChildByFieldName("body")
				if bodyNode != nil {
					hook.Body = ta.getNodeText(bodyNode)
				}
			}
		}
	}

	ta.testHooks[hookID] = hook
	return hook
}

// extractTestCallExpression handles test-related call expressions during entity extraction
func (ta *TypeScriptAnalyzer) extractTestCallExpression(node *ts.Node, parent *entities.Entity) {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return
	}

	functionName := ta.getNodeText(functionNode)
	
	// Create test entities based on the call type
	if ta.isTestSuiteCall(functionName) {
		suite := ta.extractTestSuiteFromCall(node, nil)
		if suite != nil {
			// Create a TestSuite entity
			entity := ta.createTestSuiteEntity(suite)
			if entity != nil {
				ta.currentFile.AddEntity(entity)
				if parent != nil {
					parent.AddChild(entity)
				}
			}
		}
	} else if ta.isTestCaseCall(functionName) {
		testCase := ta.extractTestCaseFromCall(node, nil)
		if testCase != nil {
			// Create a TestFunction entity
			entity := ta.createTestFunctionEntity(testCase)
			if entity != nil {
				ta.currentFile.AddEntity(entity)
				if parent != nil {
					parent.AddChild(entity)
				}
			}
		}
	}
}

// enhanceTestEntities enhances function entities with test-specific information
func (ta *TypeScriptAnalyzer) enhanceTestEntities() {
	// Enhance test functions with framework and assertion information
	for _, testCase := range ta.testCases {
		// Find corresponding entity
		for _, entity := range ta.currentFile.Entities {
			if ta.entityMatchesTest(entity, testCase) {
				// Convert to test function type
				entity.Type = entities.EntityTypeTestFunction
				entity.SetTestFramework(ta.testFramework)
				entity.SetTestType(testCase.TestType)
				entity.SetAssertionCount(len(testCase.Assertions))
				
				// Set test target based on tested entities
				if len(testCase.TestedEntities) > 0 {
					entity.SetTestTarget(testCase.TestedEntities[0])
				}
			}
		}
	}
}

// createTestSuiteEntity creates an entity from test suite info
func (ta *TypeScriptAnalyzer) createTestSuiteEntity(suite *TypeScriptTestSuiteInfo) *entities.Entity {
	entity := &entities.Entity{
		ID:       suite.ID,
		Name:     suite.Name,
		Type:     entities.EntityTypeTestSuite,
		FilePath: suite.FilePath,
		Signature: fmt.Sprintf("%s('%s')", suite.Type, suite.Name),
		Properties: make(map[string]interface{}),
	}
	
	entity.SetProperty("suite_type", suite.Type)
	entity.SetProperty("test_count", len(suite.TestCases))
	entity.SetProperty("nested_suites", len(suite.NestedSuites))
	entity.SetTestFramework(ta.testFramework)
	
	return entity
}

// createTestFunctionEntity creates an entity from test case info
func (ta *TypeScriptAnalyzer) createTestFunctionEntity(testCase *TypeScriptTestCaseInfo) *entities.Entity {
	entity := &entities.Entity{
		ID:       testCase.ID,
		Name:     testCase.Name,
		Type:     entities.EntityTypeTestFunction,
		FilePath: ta.currentFile.Path,
		Signature: fmt.Sprintf("%s('%s')", testCase.Type, testCase.Name),
		Properties: make(map[string]interface{}),
	}
	
	entity.SetTestFramework(ta.testFramework)
	entity.SetTestType(testCase.TestType)
	entity.SetAssertionCount(len(testCase.Assertions))
	
	if len(testCase.TestedEntities) > 0 {
		entity.SetTestTarget(testCase.TestedEntities[0])
	}
	
	entity.SetProperty("async", testCase.Async)
	entity.SetProperty("skip", testCase.Skip)
	entity.SetProperty("only", testCase.Only)
	entity.SetProperty("mock_count", len(testCase.Mocks))
	
	return entity
}

// extractTestRelationships extracts test-specific relationships
func (ta *TypeScriptAnalyzer) extractTestRelationships(node *ts.Node) {
	// This method processes the AST to find relationships between tests and code
	for _, testCase := range ta.testCases {
		// Create TESTS relationships for tested entities
		for _, testedEntity := range testCase.TestedEntities {
			// Find the target entity
			targetEntity := ta.findEntityByName(testedEntity)
			if targetEntity != nil {
				relID := ta.generateTestID("rel", "tests", int(testCase.StartLine))
				rel := entities.NewRelationshipByID(
					relID,
					"TESTS",
					testCase.ID,
					targetEntity.ID,
					entities.EntityTypeTestFunction,
					targetEntity.Type,
				)
				rel.SetProperty("confidence_score", 0.8)
				ta.relationships = append(ta.relationships, rel)
			}
		}
		
		// Create ASSERTS relationships for assertions
		for _, assertionID := range testCase.Assertions {
			relID := ta.generateTestID("rel", "asserts", int(testCase.StartLine))
			rel := entities.NewRelationshipByID(
				relID,
				"ASSERTS",
				testCase.ID,
				assertionID,
				entities.EntityTypeTestFunction,
				entities.EntityTypeAssertion,
			)
			ta.relationships = append(ta.relationships, rel)
		}
		
		// Create USES_MOCK relationships for mocks
		for _, mockID := range testCase.Mocks {
			if mock, exists := ta.mocks[mockID]; exists {
				relID := ta.generateTestID("rel", "mock", int(testCase.StartLine))
				rel := entities.NewRelationshipByID(
					relID,
					"USES_MOCK",
					testCase.ID,
					mockID,
					entities.EntityTypeTestFunction,
					entities.EntityTypeMock,
				)
				rel.SetProperty("mock_type", mock.Type)
				rel.SetProperty("mock_target", mock.Target)
				ta.relationships = append(ta.relationships, rel)
			}
		}
	}
	
	// Create relationships for test suites
	for _, suite := range ta.testSuites {
		// Create CONTAINS relationships for test cases in suites
		for _, testCaseID := range suite.TestCases {
			relID := ta.generateTestID("rel", "contains", int(suite.StartLine))
			rel := entities.NewRelationshipByID(
				relID,
				entities.RelationshipTypeContains,
				suite.ID,
				testCaseID,
				entities.EntityTypeTestSuite,
				entities.EntityTypeTestFunction,
			)
			ta.relationships = append(ta.relationships, rel)
		}
		
		// Create CONTAINS relationships for nested suites
		for _, nestedSuiteID := range suite.NestedSuites {
			relID := ta.generateTestID("rel", "contains", int(suite.StartLine))
			rel := entities.NewRelationshipByID(
				relID,
				entities.RelationshipTypeContains,
				suite.ID,
				nestedSuiteID,
				entities.EntityTypeTestSuite,
				entities.EntityTypeTestSuite,
			)
			ta.relationships = append(ta.relationships, rel)
		}
	}
}

// buildTestRelationships builds additional test relationships after main analysis
func (ta *TypeScriptAnalyzer) buildTestRelationships() {
	// Build coverage relationships
	for testID, coveredEntities := range ta.testCoverage {
		for _, entityID := range coveredEntities {
			rel := entities.NewRelationshipByID(
				testID + "_covers_" + entityID,
				entities.RelationshipType("COVERS"),
				testID,
				entityID,
				entities.EntityTypeTestFunction,
				entities.EntityTypeFunction,
			)
			rel.SetProperty("coverage_type", "direct")
			ta.relationships = append(ta.relationships, rel)
		}
	}
	
	// Build component test relationships
	for _, componentTest := range ta.componentTests {
		if componentTest.ComponentName != "" {
			// Find the component entity
			componentEntity := ta.findEntityByName(componentTest.ComponentName)
			if componentEntity != nil {
				rel := entities.NewRelationshipByID(
					componentTest.TestCaseID + "_tests_component_" + componentEntity.ID,
					entities.RelationshipType("TESTS_COMPONENT"),
					componentTest.TestCaseID,
					componentEntity.ID,
					entities.EntityTypeTestFunction,
					componentEntity.Type,
				)
				rel.SetProperty("render_method", componentTest.RenderMethod)
				rel.SetProperty("framework", componentTest.Framework)
				ta.relationships = append(ta.relationships, rel)
			}
		}
	}
}

// Helper methods for test detection

func (ta *TypeScriptAnalyzer) isTestSuiteCall(functionName string) bool {
	suitePatterns := []string{
		"describe", "describe.only", "describe.skip", "describe.each",
		"suite", "suite.only", "suite.skip",
		"context", "context.only", "context.skip",
	}
	
	for _, pattern := range suitePatterns {
		if functionName == pattern || strings.HasPrefix(functionName, pattern+"(") {
			return true
		}
	}
	return false
}

func (ta *TypeScriptAnalyzer) isTestCaseCall(functionName string) bool {
	testPatterns := []string{
		"it", "it.only", "it.skip", "it.each",
		"test", "test.only", "test.skip", "test.each",
		"specify", "specify.only", "specify.skip",
	}
	
	for _, pattern := range testPatterns {
		if functionName == pattern || strings.HasPrefix(functionName, pattern+"(") {
			return true
		}
	}
	return false
}

func (ta *TypeScriptAnalyzer) isTestHookCall(functionName string) bool {
	hookPatterns := []string{
		"beforeEach", "afterEach", "beforeAll", "afterAll",
		"before", "after", "setup", "teardown",
		"beforeEach.only", "afterEach.only",
	}
	
	for _, pattern := range hookPatterns {
		if functionName == pattern || strings.HasPrefix(functionName, pattern+"(") {
			return true
		}
	}
	return false
}

func (ta *TypeScriptAnalyzer) isAssertionCall(node *ts.Node) bool {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return false
	}
	
	functionText := ta.getNodeText(functionNode)
	assertionPatterns := []string{
		"expect(", "assert", "should",
		".toBe", ".toEqual", ".toMatch", ".toContain",
		".toThrow", ".toBeNull", ".toBeDefined", ".toBeTruthy", ".toBeFalsy",
		".toHaveBeenCalled", ".toHaveBeenCalledWith",
		".toMatchSnapshot", ".toMatchInlineSnapshot",
	}
	
	for _, pattern := range assertionPatterns {
		if strings.Contains(functionText, pattern) {
			return true
		}
	}
	return false
}

func (ta *TypeScriptAnalyzer) isMockCall(node *ts.Node) bool {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return false
	}
	
	functionText := ta.getNodeText(functionNode)
	mockPatterns := []string{
		"jest.fn", "jest.mock", "jest.spyOn", "jest.createMockFromModule",
		"vi.fn", "vi.mock", "vi.spyOn",
		"sinon.stub", "sinon.spy", "sinon.mock",
		"spyOn", "jasmine.createSpy", "jasmine.createSpyObj",
	}
	
	for _, pattern := range mockPatterns {
		if strings.Contains(functionText, pattern) {
			return true
		}
	}
	return false
}

func (ta *TypeScriptAnalyzer) isComponentRender(node *ts.Node) bool {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return false
	}
	
	functionText := ta.getNodeText(functionNode)
	renderPatterns := []string{
		"render(", "mount(", "shallow(",
		"screen.render", "screen.mount",
	}
	
	for _, pattern := range renderPatterns {
		if strings.Contains(functionText, pattern) {
			return true
		}
	}
	return false
}

func (ta *TypeScriptAnalyzer) isAsyncTest(node *ts.Node) bool {
	// Check if the test function is async
	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode != nil {
		for i := uint(0); i < argumentsNode.ChildCount(); i++ {
			arg := argumentsNode.Child(i)
			if arg.Kind() == "arrow_function" || arg.Kind() == "function" {
				// Check for async keyword
				for j := uint(0); j < arg.ChildCount(); j++ {
					child := arg.Child(j)
					if child.Kind() == "async" {
						return true
					}
				}
			}
		}
	}
	return false
}

func (ta *TypeScriptAnalyzer) isTestUtilityCall(functionName string) bool {
	// Functions that are test utilities, not actual code being tested
	utilities := []string{
		"expect", "assert", "should",
		"describe", "it", "test", "suite",
		"beforeEach", "afterEach", "beforeAll", "afterAll",
		"jest", "vi", "sinon",
		"render", "mount", "shallow",
		"screen", "fireEvent", "userEvent", "waitFor",
		"getBy", "queryBy", "findBy",
	}
	
	for _, util := range utilities {
		if strings.HasPrefix(functionName, util) {
			return true
		}
	}
	return false
}

func (ta *TypeScriptAnalyzer) normalizeSuiteType(functionName string) string {
	if strings.HasPrefix(functionName, "describe") {
		return "describe"
	}
	if strings.HasPrefix(functionName, "suite") {
		return "suite"
	}
	if strings.HasPrefix(functionName, "context") {
		return "context"
	}
	return functionName
}

func (ta *TypeScriptAnalyzer) normalizeTestType(functionName string) string {
	if strings.HasPrefix(functionName, "it") {
		return "it"
	}
	if strings.HasPrefix(functionName, "test") {
		return "test"
	}
	if strings.HasPrefix(functionName, "specify") {
		return "specify"
	}
	return functionName
}

func (ta *TypeScriptAnalyzer) determineTestType(testName string) string {
	nameLower := strings.ToLower(testName)
	
	// Check for unit test indicators
	if strings.Contains(nameLower, "unit") ||
		strings.Contains(nameLower, "should return") ||
		strings.Contains(nameLower, "should calculate") ||
		strings.Contains(nameLower, "should transform") {
		return "unit"
	}
	
	// Check for integration test indicators
	if strings.Contains(nameLower, "integration") ||
		strings.Contains(nameLower, "api") ||
		strings.Contains(nameLower, "endpoint") ||
		strings.Contains(nameLower, "database") ||
		strings.Contains(nameLower, "service") {
		return "integration"
	}
	
	// Check for component test indicators
	if strings.Contains(nameLower, "component") ||
		strings.Contains(nameLower, "render") ||
		strings.Contains(nameLower, "display") ||
		strings.Contains(nameLower, "ui") {
		return "component"
	}
	
	// Check for e2e test indicators
	if strings.Contains(nameLower, "e2e") ||
		strings.Contains(nameLower, "end to end") ||
		strings.Contains(nameLower, "flow") ||
		strings.Contains(nameLower, "scenario") {
		return "e2e"
	}
	
	// Check for snapshot test indicators
	if strings.Contains(nameLower, "snapshot") ||
		strings.Contains(nameLower, "match") {
		return "snapshot"
	}
	
	// Default to unit test
	return "unit"
}

func (ta *TypeScriptAnalyzer) extractCalledFunction(node *ts.Node) string {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return ""
	}
	
	functionText := ta.getNodeText(functionNode)
	
	// Skip test utility functions
	if ta.isTestUtilityCall(functionText) {
		return ""
	}
	
	// Extract the actual function name
	if functionNode.Kind() == "identifier" {
		return functionText
	}
	
	if functionNode.Kind() == "member_expression" {
		propertyNode := functionNode.ChildByFieldName("property")
		if propertyNode != nil {
			return ta.getNodeText(propertyNode)
		}
	}
	
	return ""
}

func (ta *TypeScriptAnalyzer) generateTestID(prefix, name string, position int) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s_%s_%d_%s", prefix, name, position, ta.currentFile.Path)))
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(hash[:8]))
}

func (ta *TypeScriptAnalyzer) entityMatchesTest(entity *entities.Entity, testCase *TypeScriptTestCaseInfo) bool {
	// Check if entity name matches test name or if positions overlap
	return entity.Name == testCase.Name ||
		(entity.StartByte >= uint32(testCase.StartLine) && entity.EndByte <= uint32(testCase.EndLine))
}
