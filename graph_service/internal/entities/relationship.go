package entities

import "fmt"

// Relationship represents a connection between two entities with comprehensive
// type metadata for proper database storage and query generation.
//
// Enhanced with type information to support KuzuDB's type-aware relationship
// creation, enabling proper resolution of cross-type relationships and
// validation against schema constraints.
type Relationship struct {
	ID         string                 // Unique identifier for this relationship
	Type       RelationshipType       // Type of relationship
	Source     *Entity                // Source entity (if loaded)
	Target     *Entity                // Target entity (if loaded)
	SourceID   string                 // Source entity ID (for database storage)
	TargetID   string                 // Target entity ID (for database storage)
	Properties map[string]interface{} // Additional properties
	Location   *Location              // Where the relationship occurs in code
	
	// Enhanced type metadata for proper database storage
	SourceType EntityType             // Type of the source entity
	TargetType EntityType             // Type of the target entity
	
	// Resolution metadata for debugging and validation
	ResolutionContext *RelationshipResolutionContext // Context used for entity resolution
	IsResolved        bool                           // Whether both entities were successfully resolved
	ResolutionErrors  []string                       // Any errors encountered during resolution
}

// RelationshipResolutionContext provides detailed context about how a relationship
// was resolved, including the resolution strategy used and any fallback attempts.
type RelationshipResolutionContext struct {
	// SourceResolution describes how the source entity was resolved
	SourceResolution *EntityResolutionResult
	
	// TargetResolution describes how the target entity was resolved
	TargetResolution *EntityResolutionResult
	
	// ResolutionStrategy indicates the strategy used for resolution
	ResolutionStrategy string
	
	// AnalyzerContext provides context from the analyzer that created the relationship
	AnalyzerContext string
	
	// FilePath is the file where the relationship was discovered
	FilePath string
}

// EntityResolutionResult provides detailed information about entity resolution
type EntityResolutionResult struct {
	// OriginalName is the name as it appeared in the source code
	OriginalName string
	
	// ResolvedEntity is the entity that was resolved (nil if resolution failed)
	ResolvedEntity *Entity
	
	// ResolutionMethod describes how the entity was resolved
	ResolutionMethod string
	
	// SearchedScopes lists all scopes that were searched during resolution
	SearchedScopes []string
	
	// CandidateEntities lists all potential matches found during resolution
	CandidateEntities []*Entity
	
	// ResolutionError contains any error that occurred during resolution
	ResolutionError string
}

// RelationshipType represents the type of relationship between entities
type RelationshipType string

const (
	RelationshipTypeCalls      RelationshipType = "CALLS"      // Function/method calls another function/method
	RelationshipTypeContains   RelationshipType = "CONTAINS"   // File contains function/class, class contains method
	RelationshipTypeImports    RelationshipType = "IMPORTS"    // File imports another file/module
	RelationshipTypeInherits   RelationshipType = "INHERITS"   // Class inherits from another class
	RelationshipTypeReferences RelationshipType = "REFERENCES" // Entity references another entity (variables, etc.)
	RelationshipTypeDefines    RelationshipType = "DEFINES"    // Entity defines another entity
	RelationshipTypeUses       RelationshipType = "USES"       // Entity uses another entity (instantiation, etc.)
	RelationshipTypeEmbeds     RelationshipType = "EMBEDS"     // Struct embeds another struct (Go embedding)
	RelationshipTypeImplements RelationshipType = "IMPLEMENTS" // Type implements an interface

	// Phase 2: Advanced TypeScript relationships
	RelationshipTypeDecorates     RelationshipType = "DECORATES"      // Decorator decorates class/method/property
	RelationshipTypeConstrains    RelationshipType = "CONSTRAINS"     // Generic type parameter constraints
	RelationshipTypeReExports     RelationshipType = "RE_EXPORTS"     // Module re-exports another module
	RelationshipTypeDynamicImport RelationshipType = "DYNAMIC_IMPORT" // Dynamic import relationships
	RelationshipTypeTypeOnly      RelationshipType = "TYPE_ONLY"      // Type-only import relationships
	RelationshipTypeProvides      RelationshipType = "PROVIDES"       // Service provides functionality
	RelationshipTypeInjects       RelationshipType = "INJECTS"        // Dependency injection relationships

	// Phase 3: Framework Integration relationships
	RelationshipTypeHasProps        RelationshipType = "HAS_PROPS"        // Component has props
	RelationshipTypeRendersJSX      RelationshipType = "RENDERS_JSX"      // Component renders JSX elements
	RelationshipTypeHandlesRoute    RelationshipType = "HANDLES_ROUTE"    // Controller/function handles route
	RelationshipTypeCallsAPI        RelationshipType = "CALLS_API"        // Function/component calls API endpoint
	RelationshipTypeUsesMiddleware  RelationshipType = "USES_MIDDLEWARE"  // Route uses middleware
	RelationshipTypeExposesEndpoint RelationshipType = "EXPOSES_ENDPOINT" // Controller exposes API endpoint
	RelationshipTypeAcceptsProps    RelationshipType = "ACCEPTS_PROPS"    // Component accepts specific props
	RelationshipTypeConsumesService RelationshipType = "CONSUMES_SERVICE" // Component consumes service

	// Test Coverage relationships
	RelationshipTypeTests           RelationshipType = "TESTS"           // Test function tests a production function/class/method
	RelationshipTypeCovers          RelationshipType = "COVERS"          // Test covers execution of production code
	RelationshipTypeMocks           RelationshipType = "MOCKS"           // Test mocks a dependency or external service
	RelationshipTypeSetupFor        RelationshipType = "SETUP_FOR"       // Test fixture/setup function prepares for test
	RelationshipTypeTeardownFor     RelationshipType = "TEARDOWN_FOR"    // Test teardown function cleans up after test
	RelationshipTypeAsserts         RelationshipType = "ASSERTS"         // Test assertion checks a specific condition
	RelationshipTypeVerifies        RelationshipType = "VERIFIES"        // Test verifies behavior of production code
	RelationshipTypeSpies           RelationshipType = "SPIES"           // Test spies on function calls or behavior
	RelationshipTypeStubs           RelationshipType = "STUBS"           // Test stubs out external dependencies
	RelationshipTypeFixtures        RelationshipType = "FIXTURES"        // Test uses test data fixtures
	RelationshipTypeRunsTest        RelationshipType = "RUNS_TEST"       // Test suite runs individual test cases
	RelationshipTypeGroupsTests     RelationshipType = "GROUPS_TESTS"    // Test suite groups related test cases
	RelationshipTypeSkips           RelationshipType = "SKIPS"           // Test conditionally skips other tests
	RelationshipTypeDepends         RelationshipType = "DEPENDS"         // Test depends on another test or setup
)

// Location represents a position in source code
type Location struct {
	FilePath  string // Path to the file
	StartByte uint32 // Start position in bytes
	EndByte   uint32 // End position in bytes
	Line      uint32 // Line number (if available)
	Column    uint32 // Column number (if available)
}

// NewRelationship creates a new relationship between two entities with full type metadata
func NewRelationship(id string, relType RelationshipType, source, target *Entity) *Relationship {
	relationship := &Relationship{
		ID:         id,
		Type:       relType,
		Source:     source,
		Target:     target,
		SourceID:   source.ID,
		TargetID:   target.ID,
		Properties: make(map[string]interface{}),
		SourceType: source.Type,
		TargetType: target.Type,
		IsResolved: true,
		ResolutionErrors: []string{},
	}

	// Create resolution context
	relationship.ResolutionContext = &RelationshipResolutionContext{
		SourceResolution: &EntityResolutionResult{
			OriginalName:   source.Name,
			ResolvedEntity: source,
			ResolutionMethod: "direct_entity_reference",
			SearchedScopes: []string{"direct"},
			CandidateEntities: []*Entity{source},
		},
		TargetResolution: &EntityResolutionResult{
			OriginalName:   target.Name,
			ResolvedEntity: target,
			ResolutionMethod: "direct_entity_reference",
			SearchedScopes: []string{"direct"},
			CandidateEntities: []*Entity{target},
		},
		ResolutionStrategy: "direct_entity_references",
		FilePath: source.FilePath,
	}

	return relationship
}

// NewRelationshipByID creates a new relationship using entity IDs with comprehensive type information
// This method should be used when entities are not directly available but type information is known
func NewRelationshipByID(id string, relType RelationshipType, sourceID, targetID string, sourceType, targetType EntityType) *Relationship {
	return &Relationship{
		ID:         id,
		Type:       relType,
		SourceID:   sourceID,
		TargetID:   targetID,
		Properties: make(map[string]interface{}),
		SourceType: sourceType,
		TargetType: targetType,
		IsResolved: false, // Entities not loaded, only IDs available
		ResolutionErrors: []string{},
	}
}

// NewRelationshipWithResolution creates a new relationship with detailed resolution information
// This method should be used by the EntityRegistry when resolving relationships
func NewRelationshipWithResolution(id string, relType RelationshipType, sourceID, targetID string, 
	sourceType, targetType EntityType, resolutionContext *RelationshipResolutionContext) *Relationship {
	
	isResolved := resolutionContext.SourceResolution.ResolvedEntity != nil && 
		resolutionContext.TargetResolution.ResolvedEntity != nil

	var resolutionErrors []string
	if resolutionContext.SourceResolution.ResolutionError != "" {
		resolutionErrors = append(resolutionErrors, "Source: "+resolutionContext.SourceResolution.ResolutionError)
	}
	if resolutionContext.TargetResolution.ResolutionError != "" {
		resolutionErrors = append(resolutionErrors, "Target: "+resolutionContext.TargetResolution.ResolutionError)
	}

	return &Relationship{
		ID:         id,
		Type:       relType,
		Source:     resolutionContext.SourceResolution.ResolvedEntity,
		Target:     resolutionContext.TargetResolution.ResolvedEntity,
		SourceID:   sourceID,
		TargetID:   targetID,
		Properties: make(map[string]interface{}),
		SourceType: sourceType,
		TargetType: targetType,
		ResolutionContext: resolutionContext,
		IsResolved: isResolved,
		ResolutionErrors: resolutionErrors,
	}
}

// SetProperty sets a property on the relationship
func (r *Relationship) SetProperty(key string, value interface{}) {
	r.Properties[key] = value
}

// GetProperty gets a property from the relationship
func (r *Relationship) GetProperty(key string) interface{} {
	return r.Properties[key]
}

// SetLocation sets the location where this relationship occurs
func (r *Relationship) SetLocation(filePath string, startByte, endByte uint32) {
	r.Location = &Location{
		FilePath:  filePath,
		StartByte: startByte,
		EndByte:   endByte,
	}
}

// IsValid checks if the relationship has valid source and target
func (r *Relationship) IsValid() bool {
	return r.SourceID != "" && r.TargetID != "" && r.Type != ""
}

// String returns a string representation of the relationship
func (r *Relationship) String() string {
	sourceName := r.SourceID
	if r.Source != nil {
		sourceName = r.Source.Name
	}

	targetName := r.TargetID
	if r.Target != nil {
		targetName = r.Target.Name
	}

	return sourceName + " " + string(r.Type) + " " + targetName
}

// Test-related relationship methods

// IsTestRelationship returns true if this relationship is test-related
func (r *Relationship) IsTestRelationship() bool {
	testRelationships := map[RelationshipType]bool{
		RelationshipTypeTests:       true,
		RelationshipTypeCovers:      true,
		RelationshipTypeMocks:       true,
		RelationshipTypeSetupFor:    true,
		RelationshipTypeTeardownFor: true,
		RelationshipTypeAsserts:     true,
		RelationshipTypeVerifies:    true,
		RelationshipTypeSpies:       true,
		RelationshipTypeStubs:       true,
		RelationshipTypeFixtures:    true,
		RelationshipTypeRunsTest:    true,
		RelationshipTypeGroupsTests: true,
		RelationshipTypeSkips:       true,
		RelationshipTypeDepends:     true,
	}
	return testRelationships[r.Type]
}

// GetConfidenceScore returns the confidence score for test relationships
func (r *Relationship) GetConfidenceScore() float64 {
	if score := r.GetProperty("confidence_score"); score != nil {
		if f, ok := score.(float64); ok {
			return f
		}
	}
	return 0.0
}

// SetConfidenceScore sets the confidence score for test relationships
func (r *Relationship) SetConfidenceScore(score float64) {
	r.SetProperty("confidence_score", score)
}

// GetCoverageType returns the type of coverage this relationship represents
func (r *Relationship) GetCoverageType() string {
	if coverageType := r.GetProperty("coverage_type"); coverageType != nil {
		if str, ok := coverageType.(string); ok {
			return str
		}
	}
	return ""
}

// SetCoverageType sets the type of coverage (direct, indirect, partial)
func (r *Relationship) SetCoverageType(coverageType string) {
	r.SetProperty("coverage_type", coverageType)
}

// GetTestContext returns contextual information about the test relationship
func (r *Relationship) GetTestContext() string {
	if context := r.GetProperty("test_context"); context != nil {
		if str, ok := context.(string); ok {
			return str
		}
	}
	return ""
}

// SetTestContext sets contextual information about the test relationship
func (r *Relationship) SetTestContext(context string) {
	r.SetProperty("test_context", context)
}

// GetAssertionType returns the type of assertion for assertion relationships
func (r *Relationship) GetAssertionType() string {
	if assertType := r.GetProperty("assertion_type"); assertType != nil {
		if str, ok := assertType.(string); ok {
			return str
		}
	}
	return ""
}

// SetAssertionType sets the type of assertion (equals, notNull, throws, etc.)
func (r *Relationship) SetAssertionType(assertType string) {
	r.SetProperty("assertion_type", assertType)
}

// GetMockType returns the type of mock for mock relationships
func (r *Relationship) GetMockType() string {
	if mockType := r.GetProperty("mock_type"); mockType != nil {
		if str, ok := mockType.(string); ok {
			return str
		}
	}
	return ""
}

// SetMockType sets the type of mock (spy, stub, fake, partial)
func (r *Relationship) SetMockType(mockType string) {
	r.SetProperty("mock_type", mockType)
}

// IsDirectTestRelationship returns true if this is a direct test relationship
func (r *Relationship) IsDirectTestRelationship() bool {
	return r.Type == RelationshipTypeTests || r.Type == RelationshipTypeVerifies
}

// IsCoverageRelationship returns true if this is a coverage relationship
func (r *Relationship) IsCoverageRelationship() bool {
	return r.Type == RelationshipTypeCovers
}

// IsMockingRelationship returns true if this is a mocking relationship
func (r *Relationship) IsMockingRelationship() bool {
	return r.Type == RelationshipTypeMocks || r.Type == RelationshipTypeSpies || r.Type == RelationshipTypeStubs
}

// IsSetupTeardownRelationship returns true if this is a setup/teardown relationship
func (r *Relationship) IsSetupTeardownRelationship() bool {
	return r.Type == RelationshipTypeSetupFor || r.Type == RelationshipTypeTeardownFor
}

// Enhanced validation and type checking methods

// IsValidForSchema validates if this relationship is allowed by the database schema
// This method checks if the source and target entity types are compatible with
// the relationship type according to the KuzuDB schema constraints.
func (r *Relationship) IsValidForSchema() bool {
	if r.SourceType == "" || r.TargetType == "" {
		return false
	}

	// Define valid relationship constraints based on the KuzuDB schema
	validConstraints := map[RelationshipType][]struct {
		SourceType EntityType
		TargetType EntityType
	}{
		RelationshipTypeCalls: {
			{EntityTypeFunction, EntityTypeFunction},
			{EntityTypeMethod, EntityTypeFunction},
			{EntityTypeFunction, EntityTypeMethod},
			{EntityTypeMethod, EntityTypeMethod},
			{EntityTypeTestFunction, EntityTypeFunction},
			{EntityTypeTestFunction, EntityTypeMethod},
			{EntityTypeTestCase, EntityTypeFunction},
			{EntityTypeTestCase, EntityTypeMethod},
		},
		RelationshipTypeContains: {
			{EntityTypeFile, EntityTypeFunction},
			{EntityTypeFile, EntityTypeClass},
			{EntityTypeFile, EntityTypeMethod},
			{EntityTypeFile, EntityTypeStruct},
			{EntityTypeFile, EntityTypeInterface},
			{EntityTypeFile, EntityTypeImport},
			{EntityTypeFile, EntityTypeVariable},
			{EntityTypeFile, EntityTypeTestFunction},
			{EntityTypeFile, EntityTypeTestCase},
			{EntityTypeFile, EntityTypeTestSuite},
			{EntityTypeFile, EntityTypeAssertion},
			{EntityTypeFile, EntityTypeMock},
			{EntityTypeFile, EntityTypeFixture},
		},
		RelationshipTypeImports: {
			{EntityTypeFile, EntityTypeFile},
		},
		RelationshipTypeInherits: {
			{EntityTypeClass, EntityTypeClass},
		},
		RelationshipTypeEmbeds: {
			{EntityTypeStruct, EntityTypeStruct},
		},
		RelationshipTypeImplements: {
			{EntityTypeStruct, EntityTypeInterface},
		},
		RelationshipTypeDefines: {
			{EntityTypeStruct, EntityTypeMethod},
			{EntityTypeInterface, EntityTypeMethod},
		},
		RelationshipTypeUses: {
			{EntityTypeFunction, EntityTypeStruct},
			{EntityTypeMethod, EntityTypeStruct},
			{EntityTypeFunction, EntityTypeInterface},
			{EntityTypeMethod, EntityTypeInterface},
		},
		// Test coverage relationships
		RelationshipTypeTests: {
			{EntityTypeTestFunction, EntityTypeFunction},
			{EntityTypeTestFunction, EntityTypeMethod},
			{EntityTypeTestFunction, EntityTypeClass},
			{EntityTypeTestCase, EntityTypeFunction},
			{EntityTypeTestCase, EntityTypeMethod},
			{EntityTypeTestCase, EntityTypeClass},
		},
		RelationshipTypeCovers: {
			{EntityTypeTestFunction, EntityTypeFunction},
			{EntityTypeTestFunction, EntityTypeMethod},
			{EntityTypeTestCase, EntityTypeFunction},
			{EntityTypeTestCase, EntityTypeMethod},
		},
		RelationshipTypeMocks: {
			{EntityTypeTestFunction, EntityTypeFunction},
			{EntityTypeTestFunction, EntityTypeMethod},
			{EntityTypeTestCase, EntityTypeFunction},
			{EntityTypeTestCase, EntityTypeMethod},
			{EntityTypeMock, EntityTypeFunction},
			{EntityTypeMock, EntityTypeMethod},
		},
	}

	constraints, exists := validConstraints[r.Type]
	if !exists {
		// If relationship type is not in our constraints map, assume it's valid
		// This allows for future relationship types without breaking existing code
		return true
	}

	// Check if the source/target type combination is valid
	for _, constraint := range constraints {
		if constraint.SourceType == r.SourceType && constraint.TargetType == r.TargetType {
			return true
		}
	}

	return false
}

// GetSourceTypeLabel returns the KuzuDB node label for the source entity type
func (r *Relationship) GetSourceTypeLabel() string {
	return string(r.SourceType)
}

// GetTargetTypeLabel returns the KuzuDB node label for the target entity type
func (r *Relationship) GetTargetTypeLabel() string {
	return string(r.TargetType)
}

// HasResolutionErrors returns true if there were errors during entity resolution
func (r *Relationship) HasResolutionErrors() bool {
	return len(r.ResolutionErrors) > 0
}

// GetResolutionSummary returns a human-readable summary of the resolution process
func (r *Relationship) GetResolutionSummary() string {
	if r.ResolutionContext == nil {
		return "No resolution context available"
	}

	summary := fmt.Sprintf("Strategy: %s", r.ResolutionContext.ResolutionStrategy)
	
	if r.IsResolved {
		summary += " (RESOLVED)"
	} else {
		summary += " (UNRESOLVED)"
	}

	if len(r.ResolutionErrors) > 0 {
		summary += fmt.Sprintf(" - Errors: %v", r.ResolutionErrors)
	}

	return summary
}

// Clone creates a deep copy of the relationship
func (r *Relationship) Clone() *Relationship {
	clone := &Relationship{
		ID:         r.ID,
		Type:       r.Type,
		Source:     r.Source,
		Target:     r.Target,
		SourceID:   r.SourceID,
		TargetID:   r.TargetID,
		SourceType: r.SourceType,
		TargetType: r.TargetType,
		IsResolved: r.IsResolved,
		Properties: make(map[string]interface{}),
	}

	// Deep copy properties
	for k, v := range r.Properties {
		clone.Properties[k] = v
	}

	// Deep copy resolution errors
	clone.ResolutionErrors = make([]string, len(r.ResolutionErrors))
	copy(clone.ResolutionErrors, r.ResolutionErrors)

	// Deep copy location if it exists
	if r.Location != nil {
		clone.Location = &Location{
			FilePath:  r.Location.FilePath,
			StartByte: r.Location.StartByte,
			EndByte:   r.Location.EndByte,
			Line:      r.Location.Line,
			Column:    r.Location.Column,
		}
	}

	// Note: ResolutionContext is not deep copied to avoid circular references
	// and because it's primarily for debugging
	clone.ResolutionContext = r.ResolutionContext

	return clone
}
