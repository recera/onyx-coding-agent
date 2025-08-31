package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/onyx/onyx-tui/graph_service/internal/db"
	"github.com/onyx/onyx-tui/graph_service/internal/entities"
)

// GraphBuilder orchestrates the analysis of code repositories using a comprehensive
// two-phase analysis approach with EntityRegistry for proper relationship resolution.
//
// Phase 1: Entity Discovery and Registration
//   - Analyze all files and extract entities (functions, classes, methods, etc.)
//   - Register all entities in the EntityRegistry for fast lookup
//   - Create preliminary relationships with unresolved references
//
// Phase 2: Relationship Resolution and Validation
//   - Resolve all entity references using the EntityRegistry
//   - Validate relationships against schema constraints
//   - Create fully resolved relationships with type metadata
//   - Store everything in the database
//
// This approach ensures that cross-file references are properly resolved and
// that all relationships are created with correct entity type information.
type GraphBuilder struct {
	// Core components
	database *db.KuzuDatabase
	registry *entities.EntityRegistry

	// Language-specific analyzers
	pythonAnalyzer     *PythonAnalyzer
	goAnalyzer         *GoAnalyzer
	typescriptAnalyzer *TypeScriptAnalyzer

	// Enhanced analyzers for better analysis
	enhancedGoAnalyzer *EnhancedGoAnalyzer
	advancedGoAnalyzer *AdvancedGoAnalyzer

	// Data storage
	files                   map[string]*entities.File
	allEntities             map[string]*entities.Entity
	unresolvedRelationships []*entities.Relationship // Relationships with unresolved references
	resolvedRelationships   []*entities.Relationship // Fully resolved relationships

	// Analysis configuration
	config *GraphBuilderConfig

	// Performance tracking
	stats      *BuildStats
	phaseStats map[string]*PhaseStats
}

// GraphBuilderConfig provides configuration options for the graph builder
type GraphBuilderConfig struct {
	// Analysis options
	EnableCrossFileAnalysis bool
	EnableBuiltinResolution bool
	MaxFileSize             int64 // Maximum file size to analyze (in bytes)
	// Paths/patterns to ignore during static repository walk
	// Matches if substring is present in the path or basename matches filepath.Match
	IgnorePatterns []string

	// Performance options
	EnableParallelAnalysis bool
	MaxConcurrentAnalyzers int

	// Debugging options
	EnableDetailedLogging       bool
	SaveUnresolvedRelationships bool
	GenerateAnalysisReport      bool
}

// DefaultGraphBuilderConfig returns a default configuration
func DefaultGraphBuilderConfig() *GraphBuilderConfig {
	return &GraphBuilderConfig{
		EnableCrossFileAnalysis:     true,
		EnableBuiltinResolution:     true,
		MaxFileSize:                 10 * 1024 * 1024, // 10MB
		EnableParallelAnalysis:      false,            // Disable for now to avoid complexity
		MaxConcurrentAnalyzers:      4,
		EnableDetailedLogging:       false,
		SaveUnresolvedRelationships: true,
		GenerateAnalysisReport:      true,
		IgnorePatterns: []string{
			".git",
			"node_modules",
			"vendor",
			"__pycache__",
			".venv",
			".vscode",
			".idea",
			".DS_Store",
			"dist",
			"build",
			".cache",
			"tmp",
			".goru",
			".db", // avoid folders ending with .db used by test/demo data
		},
	}
}

// PhaseStats tracks performance statistics for each analysis phase
type PhaseStats struct {
	PhaseName      string
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	ItemsProcessed int
	ErrorCount     int
	Details        map[string]interface{}
}

// NewGraphBuilder creates a new graph builder with comprehensive configuration
func NewGraphBuilder(database *db.KuzuDatabase) *GraphBuilder {
	return NewGraphBuilderWithConfig(database, DefaultGraphBuilderConfig())
}

// NewGraphBuilderWithConfig creates a new graph builder with custom configuration
func NewGraphBuilderWithConfig(database *db.KuzuDatabase, config *GraphBuilderConfig) *GraphBuilder {
	return &GraphBuilder{
		database: database,
		registry: entities.NewEntityRegistry(),
		config:   config,

		// Initialize analyzers
		pythonAnalyzer:     NewPythonAnalyzer(),
		goAnalyzer:         NewGoAnalyzer(),
		typescriptAnalyzer: NewTypeScriptAnalyzer(),
		enhancedGoAnalyzer: NewEnhancedGoAnalyzer(),
		advancedGoAnalyzer: NewAdvancedGoAnalyzer(),

		// Initialize data storage
		files:                   make(map[string]*entities.File),
		allEntities:             make(map[string]*entities.Entity),
		unresolvedRelationships: make([]*entities.Relationship, 0),
		resolvedRelationships:   make([]*entities.Relationship, 0),

		// Initialize tracking
		stats:      &BuildStats{},
		phaseStats: make(map[string]*PhaseStats),
	}
}

// BuildStats contains comprehensive statistics about the graph building process
type BuildStats struct {
	// File processing
	FilesProcessed  int
	FilesSkipped    int
	FilesWithErrors int

	// Entity discovery
	EntitiesFound     int
	FunctionsFound    int
	ClassesFound      int
	MethodsFound      int
	StructsFound      int
	InterfacesFound   int
	VariablesFound    int
	ImportsFound      int
	TestEntitiesFound int

	// Relationship processing
	UnresolvedRelationshipsFound int
	RelationshipsResolved        int
	RelationshipsFailed          int
	CrossFileRelationships       int

	// Performance metrics
	TotalAnalysisTime          time.Duration
	EntityRegistrationTime     time.Duration
	RelationshipResolutionTime time.Duration
	DatabaseStorageTime        time.Duration

	// Error tracking
	ErrorsEncountered int
	WarningsGenerated int

	// Registry statistics
	RegistryStats *entities.RegistryStats
}

// BuildGraph analyzes all files in a directory using comprehensive two-phase analysis
func (gb *GraphBuilder) BuildGraph(rootPath string) (*BuildStats, error) {
	startTime := time.Now()

	if gb.config.EnableDetailedLogging {
		fmt.Printf("Starting comprehensive graph analysis of: %s\n", rootPath)
	}

	// Phase 1: Entity Discovery and Registration
	phase1Stats, err := gb.executePhase1(rootPath)
	if err != nil {
		return gb.stats, fmt.Errorf("phase 1 failed: %w", err)
	}
	gb.phaseStats["phase1"] = phase1Stats

	// Phase 2: Relationship Resolution
	phase2Stats, err := gb.executePhase2()
	if err != nil {
		return gb.stats, fmt.Errorf("phase 2 failed: %w", err)
	}
	gb.phaseStats["phase2"] = phase2Stats

	// Phase 3: Database Storage
	phase3Stats, err := gb.executePhase3()
	if err != nil {
		return gb.stats, fmt.Errorf("phase 3 failed: %w", err)
	}
	gb.phaseStats["phase3"] = phase3Stats

	// Finalize statistics
	gb.stats.TotalAnalysisTime = time.Since(startTime)
	gb.stats.EntityRegistrationTime = phase1Stats.Duration
	gb.stats.RelationshipResolutionTime = phase2Stats.Duration
	gb.stats.DatabaseStorageTime = phase3Stats.Duration
	gb.stats.RegistryStats = gb.getRegistryStats()

	if gb.config.EnableDetailedLogging {
		gb.printAnalysisSummary()
	}

	if gb.config.GenerateAnalysisReport {
		gb.generateAnalysisReport()
	}

	return gb.stats, nil
}

// executePhase1 performs entity discovery and registration
func (gb *GraphBuilder) executePhase1(rootPath string) (*PhaseStats, error) {
	phaseStats := &PhaseStats{
		PhaseName: "Entity Discovery and Registration",
		StartTime: time.Now(),
		Details:   make(map[string]interface{}),
	}

	if gb.config.EnableDetailedLogging {
		fmt.Println("Phase 1: Discovering and registering entities...")
	}

	// Walk through all files in the directory
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			gb.stats.ErrorsEncountered++
			phaseStats.ErrorCount++
			return nil // Continue walking
		}

		// Skip directories (and prevent descent) if ignored
		if info.IsDir() {
			if gb.shouldIgnorePath(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip ignored files
		if gb.shouldIgnorePath(path) {
			gb.stats.FilesSkipped++
			return nil
		}

		// Check file size limit
		if gb.config.MaxFileSize > 0 && info.Size() > gb.config.MaxFileSize {
			if gb.config.EnableDetailedLogging {
				fmt.Printf("Skipping large file: %s (%d bytes)\n", path, info.Size())
			}
			gb.stats.FilesSkipped++
			return nil
		}

		// Process supported file types
		if gb.isSupported(path) {
			// Convert to relative path for storage
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				// If we can't make it relative, use the full path
				relPath = path
			}
			
			err = gb.processFilePhase1WithPaths(path, relPath)
			if err != nil {
				// Silently track error without printing to console
				gb.stats.ErrorsEncountered++
				gb.stats.FilesWithErrors++
				phaseStats.ErrorCount++
			} else {
				gb.stats.FilesProcessed++
				phaseStats.ItemsProcessed++
			}
		} else {
			gb.stats.FilesSkipped++
		}

		return nil
	})

	if err != nil {
		return phaseStats, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Register all entities in the registry
	err = gb.registerAllEntities()
	if err != nil {
		return phaseStats, fmt.Errorf("failed to register entities: %w", err)
	}

	phaseStats.EndTime = time.Now()
	phaseStats.Duration = phaseStats.EndTime.Sub(phaseStats.StartTime)
	phaseStats.Details["entities_registered"] = len(gb.allEntities)
	phaseStats.Details["unresolved_relationships"] = len(gb.unresolvedRelationships)

	if gb.config.EnableDetailedLogging {
		fmt.Printf("Phase 1 completed: %d entities registered, %d unresolved relationships\n",
			len(gb.allEntities), len(gb.unresolvedRelationships))
	}

	return phaseStats, nil
}

// executePhase2 performs relationship resolution using the EntityRegistry
func (gb *GraphBuilder) executePhase2() (*PhaseStats, error) {
	phaseStats := &PhaseStats{
		PhaseName: "Relationship Resolution",
		StartTime: time.Now(),
		Details:   make(map[string]interface{}),
	}

	if gb.config.EnableDetailedLogging {
		fmt.Printf("Phase 2: Resolving %d relationships...\n", len(gb.unresolvedRelationships))
	}

	resolvedCount := 0
	failedCount := 0
	crossFileCount := 0

	for _, relationship := range gb.unresolvedRelationships {
		resolvedRel, err := gb.resolveRelationship(relationship)
		if err != nil {
			if gb.config.EnableDetailedLogging {
				fmt.Printf("Failed to resolve relationship %s: %v\n", relationship.String(), err)
			}
			failedCount++
			gb.stats.RelationshipsFailed++

			// Still store unresolved relationships if configured
			if gb.config.SaveUnresolvedRelationships {
				gb.resolvedRelationships = append(gb.resolvedRelationships, relationship)
			}
		} else {
			resolvedCount++
			gb.stats.RelationshipsResolved++
			gb.resolvedRelationships = append(gb.resolvedRelationships, resolvedRel)

			// Check if it's a cross-file relationship
			if resolvedRel.Source != nil && resolvedRel.Target != nil {
				if resolvedRel.Source.FilePath != resolvedRel.Target.FilePath {
					crossFileCount++
					gb.stats.CrossFileRelationships++
				}
			}
		}
		phaseStats.ItemsProcessed++
	}

	phaseStats.EndTime = time.Now()
	phaseStats.Duration = phaseStats.EndTime.Sub(phaseStats.StartTime)
	phaseStats.Details["resolved_count"] = resolvedCount
	phaseStats.Details["failed_count"] = failedCount
	phaseStats.Details["cross_file_count"] = crossFileCount

	if gb.config.EnableDetailedLogging {
		fmt.Printf("Phase 2 completed: %d resolved, %d failed, %d cross-file\n",
			resolvedCount, failedCount, crossFileCount)
	}

	return phaseStats, nil
}

// executePhase3 stores all entities and relationships in the database
func (gb *GraphBuilder) executePhase3() (*PhaseStats, error) {
	phaseStats := &PhaseStats{
		PhaseName: "Database Storage",
		StartTime: time.Now(),
		Details:   make(map[string]interface{}),
	}

	if gb.config.EnableDetailedLogging {
		fmt.Println("Phase 3: Storing entities and relationships in database...")
	}

	err := gb.storeInDatabase()
	if err != nil {
		return phaseStats, fmt.Errorf("failed to store in database: %w", err)
	}

	phaseStats.EndTime = time.Now()
	phaseStats.Duration = phaseStats.EndTime.Sub(phaseStats.StartTime)
	phaseStats.ItemsProcessed = len(gb.allEntities) + len(gb.resolvedRelationships)

	if gb.config.EnableDetailedLogging {
		fmt.Printf("Phase 3 completed: stored %d entities and %d relationships\n",
			len(gb.allEntities), len(gb.resolvedRelationships))
	}

	return phaseStats, nil
}

// Helper methods for the two-phase analysis

// isSupported checks if a file type is supported for analysis
func (gb *GraphBuilder) isSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedExtensions := []string{".py", ".go", ".ts", ".tsx", ".js", ".jsx"}

	for _, supportedExt := range supportedExtensions {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

// shouldIgnorePath reports whether the given path should be ignored according to config patterns.
func (gb *GraphBuilder) shouldIgnorePath(filePath string) bool {
	if gb.config == nil || len(gb.config.IgnorePatterns) == 0 {
		return false
	}

	base := filepath.Base(filePath)
	for _, pattern := range gb.config.IgnorePatterns {
		if pattern == "" {
			continue
		}
		// substring match on full path
		if strings.Contains(filePath, pattern) {
			return true
		}
		// basename glob match
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
	}
	return false
}

// processFilePhase1 analyzes a single file and extracts entities (Phase 1)
// Deprecated: Use processFilePhase1WithPaths instead
func (gb *GraphBuilder) processFilePhase1(filePath string) error {
	return gb.processFilePhase1WithPaths(filePath, filePath)
}

// processFilePhase1WithPaths analyzes a single file and extracts entities (Phase 1)
// fullPath is used for reading the file, relPath is stored in the entities
func (gb *GraphBuilder) processFilePhase1WithPaths(fullPath, relPath string) error {
	// Read file content using the full path
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Analyze the file based on its extension, but use relative path for storage
	var file *entities.File
	var relationships []*entities.Relationship

	ext := strings.ToLower(filepath.Ext(relPath))
	switch ext {
	case ".py":
		file, relationships, err = gb.pythonAnalyzer.AnalyzeFile(relPath, content)
		if err != nil {
			return fmt.Errorf("failed to analyze Python file: %w", err)
		}
	case ".go":
		// Use basic Go analyzer (enhanced analyzer has incomplete relationship detection)
		file, relationships, err = gb.goAnalyzer.AnalyzeFile(relPath, content)
		if err != nil {
			return fmt.Errorf("failed to analyze Go file: %w", err)
		}
	case ".ts", ".tsx":
		file, relationships, err = gb.typescriptAnalyzer.AnalyzeFile(relPath, content)
		if err != nil {
			return fmt.Errorf("failed to analyze TypeScript file: %w", err)
		}
	case ".js", ".jsx":
		// JavaScript could potentially use the TypeScript analyzer as well
		file, relationships, err = gb.typescriptAnalyzer.AnalyzeFile(relPath, content)
		if err != nil {
			return fmt.Errorf("failed to analyze JavaScript file: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	// Store the file and its entities using relative path as key
	gb.files[relPath] = file

	// Collect all entities and update statistics
	for _, entity := range file.GetAllEntities() {
		gb.allEntities[entity.ID] = entity
		gb.stats.EntitiesFound++

		// Update detailed statistics
		switch entity.Type {
		case entities.EntityTypeFunction:
			if entity.IsMethod() {
				gb.stats.MethodsFound++
			} else {
				gb.stats.FunctionsFound++
			}
		case entities.EntityTypeMethod:
			gb.stats.MethodsFound++
		case entities.EntityTypeClass:
			gb.stats.ClassesFound++
		case entities.EntityTypeStruct:
			gb.stats.StructsFound++
		case entities.EntityTypeInterface:
			gb.stats.InterfacesFound++
		case entities.EntityTypeVariable:
			gb.stats.VariablesFound++
		case entities.EntityTypeImport:
			gb.stats.ImportsFound++
		case entities.EntityTypeTestFunction, entities.EntityTypeTestCase,
			entities.EntityTypeTestSuite, entities.EntityTypeAssertion,
			entities.EntityTypeMock, entities.EntityTypeFixture:
			gb.stats.TestEntitiesFound++
		}
	}

	// Store unresolved relationships (Phase 1 only discovers them)
	gb.unresolvedRelationships = append(gb.unresolvedRelationships, relationships...)
	gb.stats.UnresolvedRelationshipsFound += len(relationships)

	return nil
}

// registerAllEntities registers all discovered entities in the EntityRegistry
func (gb *GraphBuilder) registerAllEntities() error {
	entities := make([]*entities.Entity, 0, len(gb.allEntities))
	for _, entity := range gb.allEntities {
		entities = append(entities, entity)
	}

	return gb.registry.RegisterEntities(entities)
}

// resolveRelationship resolves a single relationship using the EntityRegistry
func (gb *GraphBuilder) resolveRelationship(relationship *entities.Relationship) (*entities.Relationship, error) {
	// Create resolution context
	context := &entities.EntityResolutionContext{
		AllowCrossFile: gb.config.EnableCrossFileAnalysis,
		AllowBuiltins:  gb.config.EnableBuiltinResolution,
	}

	// Resolve source entity if needed
	var sourceEntity *entities.Entity
	if relationship.Source != nil {
		sourceEntity = relationship.Source
		context.CurrentFile = sourceEntity.FilePath
		context.CurrentEntity = sourceEntity
	} else if relationship.SourceID != "" {
		sourceEntity = gb.registry.GetEntityByID(relationship.SourceID)
		if sourceEntity != nil {
			context.CurrentFile = sourceEntity.FilePath
			context.CurrentEntity = sourceEntity
		}
	}

	// Resolve target entity if needed
	var targetEntity *entities.Entity
	if relationship.Target != nil {
		targetEntity = relationship.Target
	} else if relationship.TargetID != "" {
		// If TargetID looks like an entity ID, try direct lookup first
		targetEntity = gb.registry.GetEntityByID(relationship.TargetID)

		// If not found, treat it as a name and resolve it
		if targetEntity == nil {
			// Set expected types based on relationship type
			switch relationship.Type {
			case entities.RelationshipTypeCalls:
				context.ExpectedTypes = []entities.EntityType{
					entities.EntityTypeFunction,
					entities.EntityTypeMethod,
					entities.EntityTypeTestFunction,
				}
			case entities.RelationshipTypeUses, entities.RelationshipTypeEmbeds, entities.RelationshipTypeImplements:
				context.ExpectedTypes = []entities.EntityType{
					entities.EntityTypeStruct,
					entities.EntityTypeInterface,
					entities.EntityTypeClass,
				}
			}

			targetEntity = gb.registry.ResolveFunction(relationship.TargetID, context)
			if targetEntity == nil && len(context.ExpectedTypes) > 0 {
				// Try general resolution if function resolution failed
				targetEntity = gb.registry.ResolveEntity(relationship.TargetID, context)
			}
		}
	}

	// Check if resolution was successful
	if sourceEntity == nil {
		return nil, fmt.Errorf("failed to resolve source entity: %s", relationship.SourceID)
	}
	if targetEntity == nil {
		return nil, fmt.Errorf("failed to resolve target entity: %s", relationship.TargetID)
	}

	// Create resolution context for the relationship
	resolutionContext := &entities.RelationshipResolutionContext{
		SourceResolution: &entities.EntityResolutionResult{
			OriginalName:      relationship.SourceID,
			ResolvedEntity:    sourceEntity,
			ResolutionMethod:  "entity_registry_lookup",
			SearchedScopes:    []string{"registry"},
			CandidateEntities: []*entities.Entity{sourceEntity},
		},
		TargetResolution: &entities.EntityResolutionResult{
			OriginalName:     relationship.TargetID,
			ResolvedEntity:   targetEntity,
			ResolutionMethod: "entity_registry_resolution",
			SearchedScopes: func() []string {
				scopes := make([]string, len(context.ExpectedTypes))
				for i, et := range context.ExpectedTypes {
					scopes[i] = string(et)
				}
				return scopes
			}(),
			CandidateEntities: []*entities.Entity{targetEntity},
		},
		ResolutionStrategy: "entity_registry_resolution",
		AnalyzerContext:    "graph_builder_phase2",
		FilePath:           context.CurrentFile,
	}

	// Create resolved relationship with full type information
	resolvedRel := entities.NewRelationshipWithResolution(
		relationship.ID,
		relationship.Type,
		sourceEntity.ID,
		targetEntity.ID,
		sourceEntity.Type,
		targetEntity.Type,
		resolutionContext,
	)

	// Copy properties and location from original relationship
	for key, value := range relationship.Properties {
		resolvedRel.SetProperty(key, value)
	}
	resolvedRel.Location = relationship.Location

	// Validate against schema constraints
	if !resolvedRel.IsValidForSchema() {
		return nil, fmt.Errorf("relationship violates schema constraints: %s %s -> %s %s",
			resolvedRel.SourceType, resolvedRel.Type, resolvedRel.TargetType, resolvedRel.TargetID)
	}

	return resolvedRel, nil
}

// getRegistryStats retrieves current registry statistics
func (gb *GraphBuilder) getRegistryStats() *entities.RegistryStats {
	stats := gb.registry.GetStats()
	return &stats
}

// printAnalysisSummary prints a detailed analysis summary
func (gb *GraphBuilder) printAnalysisSummary() {
	fmt.Println("\n=== Graph Analysis Summary ===")
	fmt.Printf("Total Analysis Time: %v\n", gb.stats.TotalAnalysisTime)

	fmt.Println("\nFile Processing:")
	fmt.Printf("  Files Processed: %d\n", gb.stats.FilesProcessed)
	fmt.Printf("  Files Skipped: %d\n", gb.stats.FilesSkipped)
	fmt.Printf("  Files with Errors: %d\n", gb.stats.FilesWithErrors)

	fmt.Println("\nEntity Discovery:")
	fmt.Printf("  Total Entities: %d\n", gb.stats.EntitiesFound)
	fmt.Printf("  Functions: %d\n", gb.stats.FunctionsFound)
	fmt.Printf("  Methods: %d\n", gb.stats.MethodsFound)
	fmt.Printf("  Classes: %d\n", gb.stats.ClassesFound)
	fmt.Printf("  Structs: %d\n", gb.stats.StructsFound)
	fmt.Printf("  Interfaces: %d\n", gb.stats.InterfacesFound)
	fmt.Printf("  Variables: %d\n", gb.stats.VariablesFound)
	fmt.Printf("  Imports: %d\n", gb.stats.ImportsFound)
	fmt.Printf("  Test Entities: %d\n", gb.stats.TestEntitiesFound)

	fmt.Println("\nRelationship Processing:")
	fmt.Printf("  Unresolved Found: %d\n", gb.stats.UnresolvedRelationshipsFound)
	fmt.Printf("  Successfully Resolved: %d\n", gb.stats.RelationshipsResolved)
	fmt.Printf("  Failed to Resolve: %d\n", gb.stats.RelationshipsFailed)
	fmt.Printf("  Cross-File Relationships: %d\n", gb.stats.CrossFileRelationships)

	fmt.Println("\nPerformance Breakdown:")
	fmt.Printf("  Entity Registration: %v\n", gb.stats.EntityRegistrationTime)
	fmt.Printf("  Relationship Resolution: %v\n", gb.stats.RelationshipResolutionTime)
	fmt.Printf("  Database Storage: %v\n", gb.stats.DatabaseStorageTime)

	fmt.Println("\nError Summary:")
	fmt.Printf("  Errors Encountered: %d\n", gb.stats.ErrorsEncountered)
	fmt.Printf("  Warnings Generated: %d\n", gb.stats.WarningsGenerated)

	if gb.stats.RegistryStats != nil {
		fmt.Println("\nRegistry Statistics:")
		fmt.Printf("  Lookup Operations: %d\n", gb.stats.RegistryStats.LookupCount)
		fmt.Printf("  Successful Resolutions: %d\n", gb.stats.RegistryStats.ResolvedCount)
		fmt.Printf("  Failed Resolutions: %d\n", gb.stats.RegistryStats.UnresolvedCount)
	}

	fmt.Println("==============================\n")
}

// generateAnalysisReport generates a detailed analysis report (placeholder)
func (gb *GraphBuilder) generateAnalysisReport() {
	// This method could generate a detailed report file
	// For now, it's a placeholder
	if gb.config.EnableDetailedLogging {
		fmt.Println("Analysis report generation not yet implemented")
	}
}

// storeInDatabase stores all entities and relationships in the database
func (gb *GraphBuilder) storeInDatabase() error {
	if gb.config.EnableDetailedLogging {
		fmt.Printf("Storing %d entities and %d relationships in database...\n",
			len(gb.allEntities), len(gb.resolvedRelationships))
	}

	// First, store all files
	fileErrors := 0
	for filePath, file := range gb.files {
		err := gb.database.AddFileNode(filePath, file.Name, file.Language)
		if err != nil {
			// Silently track error without printing to console
			fileErrors++
			gb.stats.ErrorsEncountered++
		}
	}

	// Store all entities
	entityErrors := 0
	for _, entity := range gb.allEntities {
		err := gb.database.StoreEntity(entity)
		if err != nil {
			// Silently track error without printing to console
			entityErrors++
			gb.stats.ErrorsEncountered++
		}
	}

	// Store all resolved relationships using the enhanced StoreRelationship method
	relationshipErrors := 0
	for _, rel := range gb.resolvedRelationships {
		err := gb.database.StoreRelationship(rel)
		if err != nil {
			// Silently track error without printing to console
			relationshipErrors++
			gb.stats.ErrorsEncountered++
		}
	}

	if gb.config.EnableDetailedLogging && (fileErrors > 0 || entityErrors > 0 || relationshipErrors > 0) {
		fmt.Printf("Storage completed with errors: %d file errors, %d entity errors, %d relationship errors\n",
			fileErrors, entityErrors, relationshipErrors)
	}

	return nil
}

// GetFile retrieves a file by path
func (gb *GraphBuilder) GetFile(filePath string) *entities.File {
	return gb.files[filePath]
}

// GetEntity retrieves an entity by ID
func (gb *GraphBuilder) GetEntity(entityID string) *entities.Entity {
	return gb.allEntities[entityID]
}

// GetEntitiesByName retrieves all entities with a given name
func (gb *GraphBuilder) GetEntitiesByName(name string) []*entities.Entity {
	var result []*entities.Entity
	for _, entity := range gb.allEntities {
		if entity.Name == name {
			result = append(result, entity)
		}
	}
	return result
}

// GetFiles returns all processed files
func (gb *GraphBuilder) GetFiles() map[string]*entities.File {
	return gb.files
}

// GetAllEntities returns all entities
func (gb *GraphBuilder) GetAllEntities() map[string]*entities.Entity {
	return gb.allEntities
}

// GetAllRelationships returns all resolved relationships
func (gb *GraphBuilder) GetAllRelationships() []*entities.Relationship {
	return gb.resolvedRelationships
}

// GetUnresolvedRelationships returns all unresolved relationships
func (gb *GraphBuilder) GetUnresolvedRelationships() []*entities.Relationship {
	return gb.unresolvedRelationships
}

// GetEntityRegistry returns the entity registry for external access
func (gb *GraphBuilder) GetEntityRegistry() *entities.EntityRegistry {
	return gb.registry
}

// GetConfig returns the current configuration
func (gb *GraphBuilder) GetConfig() *GraphBuilderConfig {
	return gb.config
}

// GetPhaseStats returns statistics for a specific phase
func (gb *GraphBuilder) GetPhaseStats(phaseName string) *PhaseStats {
	return gb.phaseStats[phaseName]
}

// GetAllPhaseStats returns all phase statistics
func (gb *GraphBuilder) GetAllPhaseStats() map[string]*PhaseStats {
	// Return a copy to prevent external modification
	result := make(map[string]*PhaseStats)
	for k, v := range gb.phaseStats {
		result[k] = v
	}
	return result
}

// QueryGraph executes a query against the graph database
func (gb *GraphBuilder) QueryGraph(query string) (string, error) {
	return gb.database.ExecuteQuery(query)
}
