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

// AdvancedGoAnalyzer provides sophisticated Go language analysis
type AdvancedGoAnalyzer struct {
	parser        *ts.Parser
	language      *ts.Language
	currentFile   *entities.File
	relationships []*entities.Relationship

	// Advanced Go tracking
	channels   map[string]*ChannelInfo
	goroutines map[string]*GoroutineInfo
	generics   map[string]*GenericInfo
	modules    map[string]*ModuleInfo
	packages   map[string]*PackageInfo

	// Concurrency patterns
	concurrencyPatterns []*ConcurrencyPattern
	channelOperations   []*ChannelOperation
}

// ChannelInfo represents channel usage information
type ChannelInfo struct {
	Name        string
	Type        string
	Direction   string // "send", "receive", "bidirectional"
	BufferSize  string
	Declaration *entities.Entity
	Operations  []*ChannelOperation
}

// GoroutineInfo represents goroutine usage information
type GoroutineInfo struct {
	Function     string
	LaunchSite   *entities.Entity
	ChannelUsage []string
	Context      string
}

// GenericInfo represents generic type information
type GenericInfo struct {
	Name        string
	Constraints []string
	TypeParams  []string
	Usage       []*entities.Entity
}

// ModuleInfo represents Go module information
type ModuleInfo struct {
	Name         string
	Version      string
	Dependencies []string
	Path         string
}

// PackageInfo represents Go package information
type PackageInfo struct {
	Name     string
	Path     string
	Imports  []string
	Exports  []*entities.Entity
	Internal bool
}

// ConcurrencyPattern represents detected concurrency patterns
type ConcurrencyPattern struct {
	Type        string // "worker_pool", "pipeline", "fan_out", "fan_in", "producer_consumer"
	Entities    []*entities.Entity
	Channels    []string
	Description string
}

// ChannelOperation represents channel send/receive operations
type ChannelOperation struct {
	Channel   string
	Operation string // "send", "receive", "close", "select"
	Location  *entities.Entity
	Value     string
}

// NewAdvancedGoAnalyzer creates a new advanced Go analyzer
func NewAdvancedGoAnalyzer() *AdvancedGoAnalyzer {
	parser := ts.NewParser()
	language := ts.NewLanguage(golang.Language())
	parser.SetLanguage(language)

	return &AdvancedGoAnalyzer{
		parser:              parser,
		language:            language,
		relationships:       make([]*entities.Relationship, 0),
		channels:            make(map[string]*ChannelInfo),
		goroutines:          make(map[string]*GoroutineInfo),
		generics:            make(map[string]*GenericInfo),
		modules:             make(map[string]*ModuleInfo),
		packages:            make(map[string]*PackageInfo),
		concurrencyPatterns: make([]*ConcurrencyPattern, 0),
		channelOperations:   make([]*ChannelOperation, 0),
	}
}

// AnalyzeFile analyzes a Go file with advanced feature detection
func (aga *AdvancedGoAnalyzer) AnalyzeFile(filePath string, content []byte) (*entities.File, []*entities.Relationship, error) {
	// Parse the file
	tree := aga.parser.Parse(content, nil)
	if tree == nil {
		return nil, nil, fmt.Errorf("failed to parse file %s", filePath)
	}

	// Create File entity
	file := entities.NewFile(filePath, "go", tree, content)
	aga.currentFile = file
	aga.relationships = make([]*entities.Relationship, 0)

	rootNode := tree.RootNode()

	// Phase 1: Basic entity extraction
	aga.extractBasicEntities(rootNode, nil)

	// Phase 2: Advanced Go feature detection
	aga.detectChannels(rootNode)
	aga.detectGoroutines(rootNode)
	aga.detectGenerics(rootNode)
	aga.detectPackageInfo(rootNode)

	// Phase 3: Concurrency pattern analysis
	aga.analyzeConcurrencyPatterns()

	// Phase 4: Build advanced relationships
	aga.buildAdvancedRelationships()

	return file, aga.relationships, nil
}

// extractBasicEntities extracts basic Go entities
func (aga *AdvancedGoAnalyzer) extractBasicEntities(node *ts.Node, parent *entities.Entity) {
	nodeType := node.Kind()

	switch nodeType {
	case "function_declaration":
		entity := aga.extractFunction(node, parent)
		if entity != nil {
			aga.currentFile.AddEntity(entity)
		}

	case "method_declaration":
		entity := aga.extractMethod(node, parent)
		if entity != nil {
			aga.currentFile.AddEntity(entity)
		}

	case "type_declaration":
		aga.extractTypeDeclarations(node)

	case "import_declaration":
		aga.extractImports(node)

	case "var_declaration", "const_declaration":
		aga.extractVariableDeclarations(node)
	}

	// Recursively process child nodes
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		aga.extractBasicEntities(child, parent)
	}
}

// detectChannels detects channel declarations and operations
func (aga *AdvancedGoAnalyzer) detectChannels(node *ts.Node) {
	aga.walkNode(node, func(n *ts.Node) {
		// Detect channel type declarations
		if n.Kind() == "channel_type" {
			aga.extractChannelType(n)
		}

		// Detect make(chan ...) calls
		if n.Kind() == "call_expression" {
			aga.detectChannelMake(n)
		}

		// Detect channel operations
		if n.Kind() == "send_statement" {
			aga.extractChannelSend(n)
		}

		if n.Kind() == "receive_expression" {
			aga.extractChannelReceive(n)
		}

		// Detect select statements
		if n.Kind() == "select_statement" {
			aga.extractSelectStatement(n)
		}
	})
}

// extractChannelType extracts channel type information
func (aga *AdvancedGoAnalyzer) extractChannelType(node *ts.Node) {
	channelText := aga.getNodeText(node)

	// Determine channel direction
	direction := "bidirectional"
	if strings.Contains(channelText, "<-chan") {
		direction = "receive"
	} else if strings.Contains(channelText, "chan<-") {
		direction = "send"
	}

	// Extract element type
	elementType := ""
	re := regexp.MustCompile(`chan\s*(<-\s*)?(\w+)`)
	matches := re.FindStringSubmatch(channelText)
	if len(matches) > 2 {
		elementType = matches[2]
	}

	channelInfo := &ChannelInfo{
		Type:      elementType,
		Direction: direction,
	}

	// Store channel info (key would be generated from context)
	key := fmt.Sprintf("chan_%d_%d", node.StartPosition().Row, node.StartPosition().Column)
	aga.channels[key] = channelInfo
}

// detectChannelMake detects make(chan ...) calls
func (aga *AdvancedGoAnalyzer) detectChannelMake(node *ts.Node) {
	funcNode := node.ChildByFieldName("function")
	if funcNode == nil {
		return
	}

	funcText := aga.getNodeText(funcNode)
	if funcText != "make" {
		return
	}

	// Check if it's making a channel
	args := node.ChildByFieldName("arguments")
	if args == nil {
		return
	}

	argsText := aga.getNodeText(args)
	if !strings.Contains(argsText, "chan") {
		return
	}

	// Extract channel information
	channelInfo := &ChannelInfo{
		BufferSize: "0", // Default unbuffered
	}

	// Try to extract buffer size if buffered
	re := regexp.MustCompile(`chan\s+\w+\s*,\s*(\d+)`)
	matches := re.FindStringSubmatch(argsText)
	if len(matches) > 1 {
		channelInfo.BufferSize = matches[1]
	}

	key := fmt.Sprintf("make_chan_%d_%d", node.StartPosition().Row, node.StartPosition().Column)
	aga.channels[key] = channelInfo
}

// extractChannelSend extracts channel send operations
func (aga *AdvancedGoAnalyzer) extractChannelSend(node *ts.Node) {
	channelNode := node.ChildByFieldName("channel")
	valueNode := node.ChildByFieldName("value")

	if channelNode == nil {
		return
	}

	channelName := aga.getNodeText(channelNode)
	value := ""
	if valueNode != nil {
		value = aga.getNodeText(valueNode)
	}

	operation := &ChannelOperation{
		Channel:   channelName,
		Operation: "send",
		Value:     value,
	}

	aga.channelOperations = append(aga.channelOperations, operation)
}

// extractChannelReceive extracts channel receive operations
func (aga *AdvancedGoAnalyzer) extractChannelReceive(node *ts.Node) {
	channelNode := node.ChildByFieldName("operand")
	if channelNode == nil {
		return
	}

	channelName := aga.getNodeText(channelNode)

	operation := &ChannelOperation{
		Channel:   channelName,
		Operation: "receive",
	}

	aga.channelOperations = append(aga.channelOperations, operation)
}

// extractSelectStatement extracts select statement information
func (aga *AdvancedGoAnalyzer) extractSelectStatement(node *ts.Node) {
	// Walk through communication clauses
	aga.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "communication_case" {
			// Extract channel operations in select cases
			aga.walkNode(n, func(commNode *ts.Node) {
				if commNode.Kind() == "send_statement" {
					aga.extractChannelSend(commNode)
				} else if commNode.Kind() == "receive_expression" {
					aga.extractChannelReceive(commNode)
				}
			})
		}
	})
}

// detectGoroutines detects goroutine launches
func (aga *AdvancedGoAnalyzer) detectGoroutines(node *ts.Node) {
	aga.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "go_statement" {
			aga.extractGoroutine(n)
		}
	})
}

// extractGoroutine extracts goroutine information
func (aga *AdvancedGoAnalyzer) extractGoroutine(node *ts.Node) {
	callNode := node.ChildByFieldName("expression")
	if callNode == nil {
		return
	}

	// Extract function being called
	var functionName string
	if callNode.Kind() == "call_expression" {
		funcNode := callNode.ChildByFieldName("function")
		if funcNode != nil {
			functionName = aga.getNodeText(funcNode)
		}
	} else {
		functionName = aga.getNodeText(callNode)
	}

	goroutineInfo := &GoroutineInfo{
		Function: functionName,
		Context:  "async",
	}

	key := fmt.Sprintf("goroutine_%d_%d", node.StartPosition().Row, node.StartPosition().Column)
	aga.goroutines[key] = goroutineInfo

	// Create relationship for goroutine launch
	relID := aga.generateRelationshipID("launches", "goroutine", functionName)
	relationship := entities.NewRelationshipByID(
		relID,
		entities.RelationshipTypeUses,
		"caller", // Would be resolved to actual calling function
		functionName,
		entities.EntityTypeFunction, // Source is a function that launches the goroutine
		entities.EntityTypeFunction, // Target is the function being launched as a goroutine
	)
	relationship.SetProperty("concurrency", "goroutine")
	relationship.SetProperty("async", true)

	aga.relationships = append(aga.relationships, relationship)
}

// detectGenerics detects generic type usage
func (aga *AdvancedGoAnalyzer) detectGenerics(node *ts.Node) {
	aga.walkNode(node, func(n *ts.Node) {
		// Detect type parameters in function/type declarations
		if n.Kind() == "type_parameter_list" {
			aga.extractTypeParameters(n)
		}

		// Detect type constraints
		if n.Kind() == "type_constraint" {
			aga.extractTypeConstraint(n)
		}

		// Detect generic instantiation
		if n.Kind() == "type_instantiation" {
			aga.extractGenericInstantiation(n)
		}
	})
}

// extractTypeParameters extracts generic type parameters
func (aga *AdvancedGoAnalyzer) extractTypeParameters(node *ts.Node) {
	params := []string{}
	constraints := []string{}

	aga.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "type_parameter" {
			paramText := aga.getNodeText(n)
			params = append(params, paramText)

			// Extract constraints if any
			if strings.Contains(paramText, "~") || strings.Contains(paramText, "|") {
				constraints = append(constraints, paramText)
			}
		}
	})

	if len(params) > 0 {
		genericInfo := &GenericInfo{
			TypeParams:  params,
			Constraints: constraints,
		}

		key := fmt.Sprintf("generic_%d_%d", node.StartPosition().Row, node.StartPosition().Column)
		aga.generics[key] = genericInfo
	}
}

// extractTypeConstraint extracts type constraint information
func (aga *AdvancedGoAnalyzer) extractTypeConstraint(node *ts.Node) {
	constraintText := aga.getNodeText(node)

	// This would be more sophisticated in a full implementation
	// For now, just store the raw constraint text
	key := fmt.Sprintf("constraint_%d_%d", node.StartPosition().Row, node.StartPosition().Column)
	if genericInfo, exists := aga.generics[key]; exists {
		genericInfo.Constraints = append(genericInfo.Constraints, constraintText)
	}
}

// extractGenericInstantiation extracts generic type instantiation
func (aga *AdvancedGoAnalyzer) extractGenericInstantiation(node *ts.Node) {
	instantiationText := aga.getNodeText(node)

	// Create relationship for generic usage
	relID := aga.generateRelationshipID("instantiates", "generic", instantiationText)
	relationship := entities.NewRelationshipByID(
		relID,
		entities.RelationshipTypeUses,
		"caller", // Would be resolved to actual using entity
		instantiationText,
		entities.EntityTypeFunction, // Source is a function using the generic type
		entities.EntityTypeInterface, // Target is likely a generic type or interface
	)
	relationship.SetProperty("generic_instantiation", true)

	aga.relationships = append(aga.relationships, relationship)
}

// detectPackageInfo detects package-level information
func (aga *AdvancedGoAnalyzer) detectPackageInfo(node *ts.Node) {
	// Extract package declaration
	aga.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "package_clause" {
			aga.extractPackageClause(n)
		}
	})
}

// extractPackageClause extracts package information
func (aga *AdvancedGoAnalyzer) extractPackageClause(node *ts.Node) {
	packageName := ""

	aga.walkNode(node, func(n *ts.Node) {
		if n.Kind() == "package_identifier" {
			packageName = aga.getNodeText(n)
		}
	})

	if packageName != "" {
		packageInfo := &PackageInfo{
			Name: packageName,
			Path: aga.currentFile.Path,
		}

		aga.packages[packageName] = packageInfo
	}
}

// analyzeConcurrencyPatterns analyzes common Go concurrency patterns
func (aga *AdvancedGoAnalyzer) analyzeConcurrencyPatterns() {
	// Detect worker pool pattern
	aga.detectWorkerPoolPattern()

	// Detect pipeline pattern
	aga.detectPipelinePattern()

	// Detect fan-out/fan-in patterns
	aga.detectFanOutFanInPattern()
}

// detectWorkerPoolPattern detects worker pool concurrency pattern
func (aga *AdvancedGoAnalyzer) detectWorkerPoolPattern() {
	// Look for multiple goroutines reading from the same channel
	channelUsage := make(map[string]int)

	for _, op := range aga.channelOperations {
		if op.Operation == "receive" {
			channelUsage[op.Channel]++
		}
	}

	for channel, count := range channelUsage {
		if count > 1 {
			// Potential worker pool pattern
			pattern := &ConcurrencyPattern{
				Type:        "worker_pool",
				Channels:    []string{channel},
				Description: fmt.Sprintf("Multiple workers reading from channel %s", channel),
			}
			aga.concurrencyPatterns = append(aga.concurrencyPatterns, pattern)
		}
	}
}

// detectPipelinePattern detects pipeline concurrency pattern
func (aga *AdvancedGoAnalyzer) detectPipelinePattern() {
	// Look for chains of channels (output of one becomes input of another)
	// This is a simplified detection

	sendChannels := make(map[string]bool)
	receiveChannels := make(map[string]bool)

	for _, op := range aga.channelOperations {
		if op.Operation == "send" {
			sendChannels[op.Channel] = true
		} else if op.Operation == "receive" {
			receiveChannels[op.Channel] = true
		}
	}

	// If we have channels that are both sent to and received from
	pipelineChannels := []string{}
	for channel := range sendChannels {
		if receiveChannels[channel] {
			pipelineChannels = append(pipelineChannels, channel)
		}
	}

	if len(pipelineChannels) > 1 {
		pattern := &ConcurrencyPattern{
			Type:        "pipeline",
			Channels:    pipelineChannels,
			Description: "Pipeline pattern detected with multiple stage channels",
		}
		aga.concurrencyPatterns = append(aga.concurrencyPatterns, pattern)
	}
}

// detectFanOutFanInPattern detects fan-out/fan-in patterns
func (aga *AdvancedGoAnalyzer) detectFanOutFanInPattern() {
	// Look for one source feeding multiple channels (fan-out)
	// or multiple channels feeding one destination (fan-in)

	// This would require more sophisticated analysis in a full implementation
	// For now, just detect if we have multiple channels being used

	if len(aga.channels) > 2 && len(aga.goroutines) > 1 {
		channelNames := []string{}
		for key := range aga.channels {
			channelNames = append(channelNames, key)
		}

		pattern := &ConcurrencyPattern{
			Type:        "fan_out_fan_in",
			Channels:    channelNames,
			Description: "Complex concurrency pattern with multiple channels and goroutines",
		}
		aga.concurrencyPatterns = append(aga.concurrencyPatterns, pattern)
	}
}

// buildAdvancedRelationships creates advanced Go-specific relationships
func (aga *AdvancedGoAnalyzer) buildAdvancedRelationships() {
	// Create channel communication relationships
	aga.buildChannelRelationships()

	// Create goroutine relationships
	aga.buildGoroutineRelationships()

	// Create generic relationships
	aga.buildGenericRelationships()
}

// buildChannelRelationships creates relationships for channel operations
func (aga *AdvancedGoAnalyzer) buildChannelRelationships() {
	for _, op := range aga.channelOperations {
		relType := entities.RelationshipTypeUses
		if op.Operation == "send" {
			relType = entities.RelationshipTypeDefines
		}

		relID := aga.generateRelationshipID("channel_op", op.Channel, op.Operation)
		relationship := entities.NewRelationshipByID(
			relID,
			relType,
			"entity", // Would be resolved to actual entity
			op.Channel,
			entities.EntityTypeFunction, // Source is a function using the channel
			entities.EntityTypeVariable, // Target is the channel variable
		)
		relationship.SetProperty("channel_operation", op.Operation)
		relationship.SetProperty("channel_name", op.Channel)
		if op.Value != "" {
			relationship.SetProperty("value", op.Value)
		}

		aga.relationships = append(aga.relationships, relationship)
	}
}

// buildGoroutineRelationships creates relationships for goroutine usage
func (aga *AdvancedGoAnalyzer) buildGoroutineRelationships() {
	for _, goroutine := range aga.goroutines {
		relID := aga.generateRelationshipID("goroutine", "launch", goroutine.Function)
		relationship := entities.NewRelationshipByID(
			relID,
			entities.RelationshipTypeUses,
			"launcher", // Would be resolved to actual launching entity
			goroutine.Function,
			entities.EntityTypeFunction, // Source is a function that launches the goroutine
			entities.EntityTypeFunction, // Target is the function being launched as a goroutine
		)
		relationship.SetProperty("concurrency", "goroutine")
		relationship.SetProperty("async_execution", true)

		aga.relationships = append(aga.relationships, relationship)
	}
}

// buildGenericRelationships creates relationships for generic usage
func (aga *AdvancedGoAnalyzer) buildGenericRelationships() {
	for _, generic := range aga.generics {
		for _, param := range generic.TypeParams {
			relID := aga.generateRelationshipID("generic", "param", param)
			relationship := entities.NewRelationshipByID(
				relID,
				entities.RelationshipTypeDefines,
				"generic_type", // Would be resolved to actual generic type
				param,
				entities.EntityTypeInterface, // Source is a generic type or interface
				entities.EntityTypeType, // Target is a type parameter
			)
			relationship.SetProperty("generic_parameter", true)
			relationship.SetProperty("constraints", strings.Join(generic.Constraints, ", "))

			aga.relationships = append(aga.relationships, relationship)
		}
	}
}

// Helper methods (simplified implementations)

func (aga *AdvancedGoAnalyzer) extractFunction(node *ts.Node, parent *entities.Entity) *entities.Entity {
	// Simplified implementation - would use enhanced extraction
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := aga.getNodeText(nameNode)
	id := aga.generateEntityID("function", name, node)

	return entities.NewEntity(id, name, entities.EntityTypeFunction, aga.currentFile.Path, node)
}

func (aga *AdvancedGoAnalyzer) extractMethod(node *ts.Node, parent *entities.Entity) *entities.Entity {
	// Simplified implementation
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := aga.getNodeText(nameNode)
	id := aga.generateEntityID("method", name, node)

	return entities.NewEntity(id, name, entities.EntityTypeMethod, aga.currentFile.Path, node)
}

func (aga *AdvancedGoAnalyzer) extractTypeDeclarations(node *ts.Node) {
	// Simplified implementation
}

func (aga *AdvancedGoAnalyzer) extractImports(node *ts.Node) {
	// Simplified implementation
}

func (aga *AdvancedGoAnalyzer) extractVariableDeclarations(node *ts.Node) {
	// Simplified implementation
}

func (aga *AdvancedGoAnalyzer) getNodeText(node *ts.Node) string {
	if node == nil {
		return ""
	}
	return node.Utf8Text(aga.currentFile.Content)
}

func (aga *AdvancedGoAnalyzer) generateEntityID(entityType, name string, node *ts.Node) string {
	data := fmt.Sprintf("%s:%s:%s:%d:%d",
		aga.currentFile.Path, entityType, name,
		node.StartPosition().Row, node.StartPosition().Column)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

func (aga *AdvancedGoAnalyzer) generateRelationshipID(relType, source, target string) string {
	data := fmt.Sprintf("%s:%s:%s:%s", aga.currentFile.Path, relType, source, target)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

func (aga *AdvancedGoAnalyzer) walkNode(node *ts.Node, visitor func(*ts.Node)) {
	visitor(node)
	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		aga.walkNode(child, visitor)
	}
}

// GetAdvancedAnalysisResults returns the advanced analysis results
func (aga *AdvancedGoAnalyzer) GetAdvancedAnalysisResults() *AdvancedGoAnalysisResults {
	return &AdvancedGoAnalysisResults{
		Channels:            aga.channels,
		Goroutines:          aga.goroutines,
		Generics:            aga.generics,
		Packages:            aga.packages,
		ConcurrencyPatterns: aga.concurrencyPatterns,
		ChannelOperations:   aga.channelOperations,
	}
}

// AdvancedGoAnalysisResults contains the results of advanced Go analysis
type AdvancedGoAnalysisResults struct {
	Channels            map[string]*ChannelInfo
	Goroutines          map[string]*GoroutineInfo
	Generics            map[string]*GenericInfo
	Packages            map[string]*PackageInfo
	ConcurrencyPatterns []*ConcurrencyPattern
	ChannelOperations   []*ChannelOperation
}
