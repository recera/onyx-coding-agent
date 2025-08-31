// Package graph provides comprehensive functionality to analyze code repositories
// and build knowledge graph representations using KuzuDB as the embedded graph database.
//
// This package supports multi-language code analysis (Go, Python, TypeScript) with real-time
// file watching capabilities and AI agent integration. It extracts code entities (functions,
// classes, methods, structs, interfaces) and their relationships (calls, inheritance, imports)
// to build a queryable knowledge graph.
//
// Key Features:
//   - Multi-language support with Tree-sitter parsing
//   - Real-time file watching and incremental updates
//   - KuzuDB embedded graph database integration
//   - AI agent API with code quality metrics
//   - Natural language query support via LLM integration
//   - Architectural pattern detection
//   - Cross-language dependency analysis
//
// Basic Usage:
//
//	result, err := graph.BuildGraph(graph.BuildGraphOptions{
//		RepoPath:    "/path/to/repository",
//		CleanupDB:   true,
//		LoadEnvFile: true,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer result.Close()
//
//	// Query the graph
//	classes, err := result.QueryGraph("MATCH (c:Class) RETURN c.name")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(classes)
//
// Advanced Usage with Live Analysis:
//
//	database, err := db.NewKuzuDatabase("./analysis.db")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer database.Close()
//
//	liveAnalyzer, err := analyzer.NewLiveAnalyzer(database, analyzer.DefaultWatchOptions())
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Set up AI agent callbacks
//	liveAnalyzer.SetCallbacks(
//		func(filePath string, changeType analyzer.FileChangeType) {
//			fmt.Printf("File changed: %s\n", filePath)
//		},
//		func(stats *analyzer.UpdateStats) {
//			fmt.Printf("Graph updated: %+v\n", stats)
//		},
//		func(err error) {
//			log.Printf("Error: %v\n", err)
//		},
//	)
//
//	err = liveAnalyzer.StartWatching("./src")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Environment Variables:
//   - OPENAI_API_KEY: Required for LLM-powered natural language queries
//
// Dependencies:
//   - KuzuDB: Embedded graph database (github.com/kuzudb/go-kuzu)
//   - Tree-sitter: Language parsing (github.com/tree-sitter/go-tree-sitter)
//   - fsnotify: File watching (github.com/fsnotify/fsnotify)
//   - OpenAI API: Natural language queries (github.com/sashabaranov/go-openai)
package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/onyx/onyx-tui/graph_service/internal/analyzer"
	"github.com/onyx/onyx-tui/graph_service/internal/db"
	"github.com/onyx/onyx-tui/graph_service/internal/entities"
	"github.com/onyx/onyx-tui/graph_service/internal/git"
	"github.com/onyx/onyx-tui/graph_service/internal/llm"

	"github.com/joho/godotenv"
)

// BuildGraphOptions configures the code graph analysis process.
// These options control how the repository is analyzed, where the database
// is stored, and various behavioral settings.
//
// Example usage:
//
//	opts := BuildGraphOptions{
//		RepoPath:    "/path/to/repository",
//		DBPath:      "./analysis.db",
//		CleanupDB:   false,  // Keep database for reuse
//		LoadEnvFile: true,   // Load .env for API keys
//	}
type BuildGraphOptions struct {
	// RepoPath specifies the filesystem path to the repository to analyze.
	// This should be an absolute or relative path to a directory containing
	// source code files. If empty, RepoURL must be provided instead.
	//
	// Example: "/home/user/myproject" or "./src"
	RepoPath string

	// RepoURL specifies the Git repository URL to clone and analyze.
	// The repository will be cloned to a temporary directory and analyzed.
	// If empty, RepoPath must be provided instead.
	//
	// Supported formats:
	//   - HTTPS: "https://github.com/user/repo.git"
	//   - SSH: "git@github.com:user/repo.git"
	//   - Local: "file:///path/to/repo.git"
	RepoURL string

	// DBPath specifies where the KuzuDB database files will be created.
	// If empty, a temporary directory will be created automatically.
	// The database stores the knowledge graph and can be reused across
	// multiple analysis sessions.
	//
	// Example: "./analysis.db" or "/tmp/code-graph.db"
	DBPath string

	// CleanupDB determines whether to delete the database directory
	// after analysis completes. Set to true for one-time analysis,
	// false to preserve the database for reuse or inspection.
	//
	// Note: If DBPath is empty (temporary directory), cleanup will
	// occur regardless of this setting.
	CleanupDB bool

	// LoadEnvFile specifies whether to load environment variables
	// from a .env file in the current directory. This is useful for
	// loading API keys (like OPENAI_API_KEY) required for LLM integration.
	//
	// The .env file should contain key=value pairs, one per line:
	//   OPENAI_API_KEY=your_api_key_here
	LoadEnvFile bool

	// IgnorePatterns specifies paths/patterns to be excluded from static analysis.
	// Defaults include common build and VCS directories, plus ".goru". Patterns
	// match on substring within full path or basename glob.
	IgnorePatterns []string
}

// BuildGraphResult contains the complete results of code graph analysis.
// This structure provides access to the database, statistics, and all
// analyzed entities and relationships. It serves as the main interface
// for querying and inspecting the generated knowledge graph.
//
// The caller is responsible for calling Close() to clean up database
// connections when done with the result.
//
// Example usage:
//
//	result, err := BuildGraph(opts)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer result.Close()
//
//	// Access statistics
//	fmt.Printf("Analyzed %d files, found %d functions\n",
//		result.Stats.FilesCount, result.Stats.FunctionsCount)
//
//	// Query specific entities
//	entities := result.GetEntityByName("MyClass")
//	for _, entity := range entities {
//		fmt.Printf("Found %s at %s\n", entity.Name, entity.FilePath)
//	}
type BuildGraphResult struct {
	// DBPath contains the filesystem path to the KuzuDB database directory.
	// This can be used to reconnect to the database later or for backup purposes.
	DBPath string

	// Database provides direct access to the KuzuDB instance for advanced
	// operations like custom Cypher queries or schema inspection.
	// Use with caution as direct database operations may affect consistency.
	Database *db.KuzuDatabase

	// Stats contains high-level statistics about the analysis results,
	// including counts of files, entities, and any errors encountered.
	Stats BuildGraphStats

	// Builder provides access to the graph builder instance with full
	// entity and relationship data. This is the preferred way to access
	// detailed analysis results programmatically.
	Builder *analyzer.GraphBuilder
}

// BuildGraphStats provides quantitative metrics about the code analysis.
// These statistics give an overview of the repository's structure and
// the success rate of the analysis process.
//
// Example interpretation:
//   - FilesCount: Total source files analyzed
//   - FunctionsCount: Standalone functions found
//   - MethodsCount: Class/struct methods found
//   - ClassesCount: Classes, structs, interfaces found
//   - CallsCount: Function/method call relationships
//   - ErrorsCount: Files that couldn't be analyzed
type BuildGraphStats struct {
	// FunctionsCount is the total number of standalone functions
	// discovered across all analyzed files. This excludes methods
	// that belong to classes or structs.
	FunctionsCount int

	// ClassesCount is the total number of classes, structs, and
	// interfaces found. This represents the major structural
	// components of the codebase.
	ClassesCount int

	// CallsCount approximates the number of function and method
	// call relationships identified. This gives an indication of
	// code interconnectedness and complexity.
	CallsCount int

	// MethodsCount is the total number of methods belonging to
	// classes, structs, or interfaces. This excludes standalone
	// functions counted in FunctionsCount.
	MethodsCount int

	// FilesCount is the total number of source files that were
	// successfully processed during analysis. Files that couldn't
	// be parsed are excluded from this count.
	FilesCount int

	// ErrorsCount indicates the number of files or entities that
	// couldn't be analyzed due to parsing errors, unsupported
	// language features, or other issues.
	ErrorsCount int
}

// AnalysisResult provides comprehensive access to all entities, files,
// and relationships discovered during code analysis. This structure
// offers a complete view of the knowledge graph for programmatic
// inspection and processing.
//
// Use this when you need to:
//   - Iterate over all discovered entities
//   - Analyze cross-file relationships
//   - Generate reports or documentation
//   - Perform custom analysis algorithms
//
// Example usage:
//
//	analysis := result.GetAnalysisResult()
//	for filePath, file := range analysis.Files {
//		fmt.Printf("File: %s has %d entities\n",
//			filePath, len(file.GetAllEntities()))
//	}
type AnalysisResult struct {
	// Files maps file paths to File entities containing all entities
	// discovered in each source file. Keys are relative paths from
	// the repository root.
	Files map[string]*entities.File

	// Entities maps entity IDs to Entity objects for all discovered
	// code entities (functions, classes, methods, etc.). Entity IDs
	// are globally unique within the analysis.
	Entities map[string]*entities.Entity

	// Relationships contains all discovered relationships between
	// entities, such as function calls, class inheritance, and
	// import dependencies.
	Relationships []*entities.Relationship

	// Stats provides detailed internal statistics from the analysis
	// engine, including performance metrics and entity counts.
	Stats *analyzer.BuildStats
}

// BuildGraph performs comprehensive analysis of a code repository and constructs
// a knowledge graph representation using KuzuDB as the embedded graph database.
//
// This function is the main entry point for code analysis. It handles repository
// setup (cloning if needed), database initialization, schema creation, and
// orchestrates the multi-language analysis process.
//
// The analysis process includes:
//  1. Repository preparation (cloning from URL if needed)
//  2. Database initialization and schema setup
//  3. Multi-language parsing with Tree-sitter
//  4. Entity extraction (functions, classes, methods, etc.)
//  5. Relationship discovery (calls, inheritance, imports)
//  6. Graph storage in KuzuDB
//
// Supported Languages:
//   - Go: Full support for packages, functions, methods, structs, interfaces
//   - Python: Classes, functions, methods, imports, inheritance
//   - TypeScript: Planned support for classes, functions, interfaces, modules
//
// Parameters:
//   - opts: Configuration options controlling analysis behavior, database location,
//     and repository source. See BuildGraphOptions for detailed field descriptions.
//
// Returns:
//   - *BuildGraphResult: Complete analysis results with database access, statistics,
//     and entity/relationship data. Caller must call Close() when finished.
//   - error: Non-nil if analysis fails due to repository access, database issues,
//     or parsing errors that prevent analysis completion.
//
// Example usage:
//
//	// Analyze local repository
//	result, err := BuildGraph(BuildGraphOptions{
//		RepoPath:    "./my-project",
//		DBPath:      "./analysis.db",
//		CleanupDB:   false,
//		LoadEnvFile: true,
//	})
//	if err != nil {
//		return fmt.Errorf("analysis failed: %w", err)
//	}
//	defer result.Close()
//
//	// Analyze remote repository
//	result, err := BuildGraph(BuildGraphOptions{
//		RepoURL:     "https://github.com/user/repo.git",
//		CleanupDB:   true,  // Clean up temporary files
//		LoadEnvFile: true,
//	})
//	if err != nil {
//		return fmt.Errorf("analysis failed: %w", err)
//	}
//	defer result.Close()
//
// Error Conditions:
//   - Repository not found or inaccessible
//   - Database initialization or schema creation failure
//   - Insufficient permissions for file/directory access
//   - Out of disk space for database storage
//   - Parsing failures in critical files (warnings for individual files)
//
// Performance Considerations:
//   - Analysis time scales with repository size and complexity
//   - Database storage requirements are proportional to entity count
//   - Memory usage peaks during relationship resolution phase
//   - Consider using temporary databases for large one-time analyses
func BuildGraph(opts BuildGraphOptions) (*BuildGraphResult, error) {
	// Load environment variables if requested
	if opts.LoadEnvFile {
		_ = godotenv.Load() // Silently continue if .env doesn't exist
	}

	// Determine repository path
	repoPath := opts.RepoPath
	if repoPath == "" && opts.RepoURL != "" {
		// Clone repository if URL is provided
		var err error
		repoPath, err = git.CloneRepository(opts.RepoURL)
		if err != nil {
			return nil, fmt.Errorf("failed to clone repository: %w", err)
		}
		// Clean up cloned repo when done
		defer os.RemoveAll(repoPath)
	} else if repoPath == "" {
		return nil, fmt.Errorf("either RepoPath or RepoURL must be provided")
	}

	// Determine database path
	dbPath := opts.DBPath
	if dbPath == "" {
		// Create temporary directory for database
		var err error
		dbPath, err = os.MkdirTemp("", "kuzudb_*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp dir for db: %w", err)
		}
	}

	// Clean up database directory if requested
	if opts.CleanupDB {
		defer os.RemoveAll(dbPath)
	}

	// Initialize KuzuDB
	kdb, err := db.NewKuzuDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create schema
	err = kdb.CreateSchema()
	if err != nil {
		kdb.Close() // Clean up on error
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Create graph builder with the new entity system and configure ignore patterns
	config := analyzer.DefaultGraphBuilderConfig()
	if len(opts.IgnorePatterns) > 0 {
		// Merge while ensuring ".goru" remains ignored by default
		config.IgnorePatterns = append(config.IgnorePatterns, opts.IgnorePatterns...)
	}
	// Always ensure the database path itself is ignored
	if dbPath != "" {
		// Get the relative path of the database from the repo path
		relDbPath, err := filepath.Rel(repoPath, dbPath)
		if err == nil && !strings.HasPrefix(relDbPath, "..") {
			// Only add to ignore if the database is inside the repo
			config.IgnorePatterns = append(config.IgnorePatterns, relDbPath)
		} else {
			// If database is outside repo, just add the base name to be safe
			config.IgnorePatterns = append(config.IgnorePatterns, filepath.Base(dbPath))
		}
	}
	builder := analyzer.NewGraphBuilderWithConfig(kdb, config)

	// Build the graph using the sophisticated analyzer
	stats, err := builder.BuildGraph(repoPath)
	if err != nil {
		kdb.Close() // Clean up on error
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	// Convert internal stats to external stats format for backward compatibility
	extStats := BuildGraphStats{
		FunctionsCount: stats.FunctionsFound,
		ClassesCount:   stats.ClassesFound,
		MethodsCount:   stats.MethodsFound,
		CallsCount:     stats.UnresolvedRelationshipsFound, // Total relationships discovered
		FilesCount:     stats.FilesProcessed,
		ErrorsCount:    stats.ErrorsEncountered,
	}

	// Return result - note: caller is responsible for closing the database
	return &BuildGraphResult{
		DBPath:   dbPath,
		Database: kdb,
		Stats:    extStats,
		Builder:  builder,
	}, nil
}

// GetAnalysisResult returns detailed analysis results with full entity access
func (r *BuildGraphResult) GetAnalysisResult() *AnalysisResult {
	if r.Builder == nil {
		return &AnalysisResult{
			Files:         make(map[string]*entities.File),
			Entities:      make(map[string]*entities.Entity),
			Relationships: make([]*entities.Relationship, 0),
		}
	}

	return &AnalysisResult{
		Files:         r.Builder.GetFiles(),
		Entities:      r.Builder.GetAllEntities(),
		Relationships: r.Builder.GetAllRelationships(),
	}
}

// QueryGraph uses LLM to generate and execute a Cypher query against the code graph
// based on a natural language question
func QueryGraph(db *db.KuzuDatabase, question string) (string, error) {
	// Initialize LLM client
	client, err := llm.NewLLMClient()
	if err != nil {
		return "", fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Get database schema
	schema, err := db.GetSchema()
	if err != nil {
		return "", fmt.Errorf("failed to get database schema: %w", err)
	}

	// Generate Cypher query
	query, err := client.GenerateQuery(question, schema)
	if err != nil {
		return "", fmt.Errorf("failed to generate query: %w", err)
	}

	// Execute query
	result, err := db.ExecuteQuery(query)
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %w", err)
	}

	return result, nil
}

// QueryGraphWithBuilder is a convenience method to query using the graph builder
func (r *BuildGraphResult) QueryGraph(question string) (string, error) {
	if r.Builder == nil {
		return QueryGraph(r.Database, question)
	}
	return r.Builder.QueryGraph(question)
}

// GetEntityByName finds entities by name across all files
func (r *BuildGraphResult) GetEntityByName(name string) []*entities.Entity {
	if r.Builder == nil {
		return nil
	}
	return r.Builder.GetEntitiesByName(name)
}

// GetFile retrieves a file by path
func (r *BuildGraphResult) GetFile(filePath string) *entities.File {
	if r.Builder == nil {
		return nil
	}
	return r.Builder.GetFile(filePath)
}

// GetAllFiles returns all processed files
func (r *BuildGraphResult) GetAllFiles() map[string]*entities.File {
	if r.Builder == nil {
		return make(map[string]*entities.File)
	}
	return r.Builder.GetFiles()
}

// GetAllEntities returns all entities
func (r *BuildGraphResult) GetAllEntities() map[string]*entities.Entity {
	if r.Builder == nil {
		return make(map[string]*entities.Entity)
	}
	return r.Builder.GetAllEntities()
}

// GetAllRelationships returns all relationships
func (r *BuildGraphResult) GetAllRelationships() []*entities.Relationship {
	if r.Builder == nil {
		return make([]*entities.Relationship, 0)
	}
	return r.Builder.GetAllRelationships()
}

// Close closes the database connection
func (r *BuildGraphResult) Close() {
	if r.Database != nil {
		r.Database.Close()
	}
}

// GetEntityFilePath returns the file path for an entity by name.
// This is a convenience method for AI coders to quickly locate source files.
//
// Parameters:
//   - entityName: Name of the entity (function, class, method, etc.)
//
// Returns:
//   - string: File path where the entity is defined
//   - error: Non-nil if entity not found or query fails
//
// Example:
//
//	filePath, err := result.GetEntityFilePath("calculateSum")
//	if err != nil {
//		log.Printf("Entity not found: %v", err)
//		return
//	}
//	fmt.Printf("Function calculateSum is in: %s\n", filePath)
func (r *BuildGraphResult) GetEntityFilePath(entityName string) (string, error) {
	query := fmt.Sprintf(`
		MATCH (n {name: "%s"})
		RETURN n.file_path as file_path
		LIMIT 1
	`, entityName)

	result, err := r.QueryGraph(query)
	if err != nil {
		return "", fmt.Errorf("failed to query entity file path: %w", err)
	}

	if result == "" {
		return "", fmt.Errorf("entity '%s' not found", entityName)
	}

	// Extract file path from result (remove newline)
	filePath := strings.TrimSpace(result)
	return filePath, nil
}

// GetFileEntities returns all entities defined in a specific file.
// This helps AI coders understand the complete context of a source file.
//
// Parameters:
//   - filePath: Path to the source file (can be relative or absolute)
//
// Returns:
//   - []*entities.Entity: List of entities defined in the file
//   - error: Non-nil if file not found or query fails
//
// Example:
//
//	entities, err := result.GetFileEntities("src/main.go")
//	if err != nil {
//		log.Printf("Failed to get entities: %v", err)
//		return
//	}
//	for _, entity := range entities {
//		fmt.Printf("%s: %s\n", entity.Type, entity.Name)
//	}
func (r *BuildGraphResult) GetFileEntities(filePath string) ([]*entities.Entity, error) {
	if r.Builder == nil {
		return nil, fmt.Errorf("builder not available")
	}

	// Normalize the file path to match storage format
	var matchingEntities []*entities.Entity

	allEntities := r.Builder.GetAllEntities()
	for _, entity := range allEntities {
		// Check if entity's file path matches (handle both exact and suffix matching)
		if entity.FilePath == filePath || strings.HasSuffix(entity.FilePath, filePath) {
			matchingEntities = append(matchingEntities, entity)
		}
	}

	return matchingEntities, nil
}

// GetRelatedFiles finds all files that have relationships with entities in the given file.
// Useful for AI coders to understand file dependencies and impact analysis.
//
// Parameters:
//   - filePath: Path to the source file to analyze
//
// Returns:
//   - []string: List of related file paths
//   - error: Non-nil if query fails
//
// Example:
//
//	relatedFiles, err := result.GetRelatedFiles("src/models/user.go")
//	if err != nil {
//		log.Printf("Failed to find related files: %v", err)
//		return
//	}
//	fmt.Printf("Files that depend on user.go: %v\n", relatedFiles)
func (r *BuildGraphResult) GetRelatedFiles(filePath string) ([]string, error) {
	query := fmt.Sprintf(`
		MATCH (f:File {path: "%s"})-[:Contains]->(entity)
		MATCH (entity)-[rel]-(related)
		WHERE related.file_path IS NOT NULL AND related.file_path <> "%s"
		RETURN DISTINCT related.file_path
		ORDER BY related.file_path
	`, filePath, filePath)

	result, err := r.QueryGraph(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query related files: %w", err)
	}

	// Parse the result into a slice of file paths
	var files []string
	if result != "" {
		lines := strings.Split(strings.TrimSpace(result), "\n")
		files = append(files, lines...)
	}

	return files, nil
}

// GetCrossFileReferences finds all cross-file references (calls, imports, etc.) for a given file.
// Essential for AI coders to understand external dependencies.
//
// Parameters:
//   - filePath: Path to the source file
//
// Returns:
//   - map[string][]string: Map of relationship type to list of external file paths
//   - error: Non-nil if query fails
//
// Example:
//
//	refs, err := result.GetCrossFileReferences("src/handlers/api.go")
//	if err != nil {
//		log.Printf("Failed to get references: %v", err)
//		return
//	}
//	for relType, files := range refs {
//		fmt.Printf("%s relationships with: %v\n", relType, files)
//	}
func (r *BuildGraphResult) GetCrossFileReferences(filePath string) (map[string][]string, error) {
	query := fmt.Sprintf(`
		MATCH (source)-[rel]-(target)
		WHERE source.file_path = "%s" AND target.file_path <> "%s"
		RETURN type(rel) as relationship_type, collect(DISTINCT target.file_path) as target_files
	`, filePath, filePath)

	result, err := r.QueryGraph(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cross-file references: %w", err)
	}

	// Parse result into map
	references := make(map[string][]string)
	if result != "" {
		lines := strings.Split(strings.TrimSpace(result), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "\t|\t")
			if len(parts) >= 2 {
				relType := parts[0]
				// Parse the array of files
				filesStr := strings.Trim(parts[1], "[]")
				if filesStr != "" {
					files := strings.Split(filesStr, ", ")
					for i := range files {
						files[i] = strings.Trim(files[i], "\"")
					}
					references[relType] = files
				}
			}
		}
	}

	return references, nil
}

// Test Coverage API Methods

// TestCoverageResult represents the coverage analysis result for an entity
type TestCoverageResult struct {
	EntityID        string                 `json:"entity_id"`
	EntityName      string                 `json:"entity_name"`
	EntityType      string                 `json:"entity_type"`
	FilePath        string                 `json:"file_path"`
	IsCovered       bool                   `json:"is_covered"`
	DirectTests     []*TestInfo            `json:"direct_tests"`
	IndirectTests   []*TestInfo            `json:"indirect_tests"`
	CoverageScore   float64                `json:"coverage_score"`
	TestCount       int                    `json:"test_count"`
	AssertionCount  int                    `json:"assertion_count"`
	CoverageMetrics map[string]interface{} `json:"coverage_metrics"`
}

// TestInfo represents information about a test that covers an entity
type TestInfo struct {
	TestID          string  `json:"test_id"`
	TestName        string  `json:"test_name"`
	TestType        string  `json:"test_type"`
	TestFramework   string  `json:"test_framework"`
	FilePath        string  `json:"file_path"`
	ConfidenceScore float64 `json:"confidence_score"`
	CoverageType    string  `json:"coverage_type"`
	AssertionCount  int     `json:"assertion_count"`
}

// CoverageMetrics represents overall coverage statistics for the codebase
type CoverageMetrics struct {
	TotalEntities      int                    `json:"total_entities"`
	TestedEntities     int                    `json:"tested_entities"`
	UntestedEntities   int                    `json:"untested_entities"`
	CoveragePercentage float64                `json:"coverage_percentage"`
	TestCount          int                    `json:"test_count"`
	TotalAssertions    int                    `json:"total_assertions"`
	CoverageByType     map[string]float64     `json:"coverage_by_type"`
	CoverageByFile     map[string]float64     `json:"coverage_by_file"`
	TestFrameworks     map[string]int         `json:"test_frameworks"`
	TestTypes          map[string]int         `json:"test_types"`
	Details            map[string]interface{} `json:"details"`
}

// GetTestCoverage calculates coverage for a specific entity
func (r *BuildGraphResult) GetTestCoverage(entityID string) (*TestCoverageResult, error) {
	if r.Builder == nil {
		return nil, fmt.Errorf("builder not available")
	}

	// Get the entity
	entity := r.Builder.GetEntity(entityID)
	if entity == nil {
		return nil, fmt.Errorf("entity not found: %s", entityID)
	}

	result := &TestCoverageResult{
		EntityID:        entity.ID,
		EntityName:      entity.Name,
		EntityType:      string(entity.Type),
		FilePath:        entity.FilePath,
		DirectTests:     make([]*TestInfo, 0),
		IndirectTests:   make([]*TestInfo, 0),
		CoverageMetrics: make(map[string]interface{}),
	}

	// Find direct tests (tests that directly test this entity)
	directTests := r.findDirectTests(entity)
	result.DirectTests = directTests

	// Find indirect tests (tests that call functions that test this entity)
	indirectTests := r.findIndirectTests(entity)
	result.IndirectTests = indirectTests

	// Calculate coverage metrics
	result.TestCount = len(directTests) + len(indirectTests)
	result.IsCovered = result.TestCount > 0

	// Calculate assertion count
	assertionCount := 0
	for _, test := range directTests {
		assertionCount += test.AssertionCount
	}
	for _, test := range indirectTests {
		assertionCount += test.AssertionCount
	}
	result.AssertionCount = assertionCount

	// Calculate coverage score (0.0 to 1.0)
	result.CoverageScore = r.calculateCoverageScore(directTests, indirectTests)

	// Add additional coverage metrics
	result.CoverageMetrics["has_unit_tests"] = r.hasTestType(directTests, "unit")
	result.CoverageMetrics["has_integration_tests"] = r.hasTestType(directTests, "integration")
	result.CoverageMetrics["has_direct_coverage"] = len(directTests) > 0
	result.CoverageMetrics["has_indirect_coverage"] = len(indirectTests) > 0
	result.CoverageMetrics["confidence_score"] = r.calculateConfidenceScore(directTests, indirectTests)

	return result, nil
}

// GetUncoveredEntities finds all entities without tests
func (r *BuildGraphResult) GetUncoveredEntities() ([]*entities.Entity, error) {
	if r.Builder == nil {
		return nil, fmt.Errorf("builder not available")
	}

	uncovered := make([]*entities.Entity, 0)

	// Get all production entities (non-test entities)
	allEntities := r.Builder.GetAllEntities()
	for _, entity := range allEntities {
		if !entity.IsTest() && r.isProductionEntity(entity) {
			// Check if this entity has any test coverage
			coverage, err := r.GetTestCoverage(entity.ID)
			if err != nil {
				continue // Skip entities we can't analyze
			}

			if !coverage.IsCovered {
				uncovered = append(uncovered, entity)
			}
		}
	}

	return uncovered, nil
}

// GetTestsByTarget finds all tests for a specific entity
func (r *BuildGraphResult) GetTestsByTarget(entityID string) ([]*entities.Entity, error) {
	if r.Builder == nil {
		return nil, fmt.Errorf("builder not available")
	}

	tests := make([]*entities.Entity, 0)

	// Find direct tests using TESTS relationships
	query := fmt.Sprintf(`
		MATCH (test)-[:TESTS]->(target {id: "%s"})
		RETURN test.id, test.name, test.file_path
	`, entityID)

	result, err := r.QueryGraph(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tests: %w", err)
	}

	if result != "" {
		lines := strings.Split(strings.TrimSpace(result), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "\t|\t")
			if len(parts) >= 1 {
				testID := strings.Trim(parts[0], "\"")
				if testEntity := r.Builder.GetEntity(testID); testEntity != nil {
					tests = append(tests, testEntity)
				}
			}
		}
	}

	// Also find tests using COVERS relationships
	coverQuery := fmt.Sprintf(`
		MATCH (test)-[:COVERS]->(target {id: "%s"})
		RETURN test.id, test.name, test.file_path
	`, entityID)

	coverResult, err := r.QueryGraph(coverQuery)
	if err == nil && coverResult != "" {
		lines := strings.Split(strings.TrimSpace(coverResult), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "\t|\t")
			if len(parts) >= 1 {
				testID := strings.Trim(parts[0], "\"")
				if testEntity := r.Builder.GetEntity(testID); testEntity != nil {
					// Check if we already have this test
					found := false
					for _, existingTest := range tests {
						if existingTest.ID == testEntity.ID {
							found = true
							break
						}
					}
					if !found {
						tests = append(tests, testEntity)
					}
				}
			}
		}
	}

	return tests, nil
}

// GetCoverageMetrics returns overall coverage statistics for the codebase
func (r *BuildGraphResult) GetCoverageMetrics() (*CoverageMetrics, error) {
	if r.Builder == nil {
		return nil, fmt.Errorf("builder not available")
	}

	metrics := &CoverageMetrics{
		CoverageByType: make(map[string]float64),
		CoverageByFile: make(map[string]float64),
		TestFrameworks: make(map[string]int),
		TestTypes:      make(map[string]int),
		Details:        make(map[string]interface{}),
	}

	allEntities := r.Builder.GetAllEntities()

	// Count production entities and tests
	productionEntities := make([]*entities.Entity, 0)
	testEntities := make([]*entities.Entity, 0)

	for _, entity := range allEntities {
		if entity.IsTest() {
			testEntities = append(testEntities, entity)

			// Count test frameworks
			framework := entity.GetTestFramework()
			if framework != "" {
				metrics.TestFrameworks[framework]++
			}

			// Count test types
			testType := entity.GetTestType()
			if testType != "" {
				metrics.TestTypes[testType]++
			}

			// Count assertions
			metrics.TotalAssertions += entity.GetAssertionCount()
		} else if r.isProductionEntity(entity) {
			productionEntities = append(productionEntities, entity)
		}
	}

	metrics.TotalEntities = len(productionEntities)
	metrics.TestCount = len(testEntities)

	// Calculate coverage for each production entity
	testedEntities := 0
	coverageByType := make(map[string]map[string]int)
	coverageByFile := make(map[string]map[string]int)

	for _, entity := range productionEntities {
		coverage, err := r.GetTestCoverage(entity.ID)
		if err != nil {
			continue
		}

		if coverage.IsCovered {
			testedEntities++
		}

		// Track coverage by entity type
		entityType := string(entity.Type)
		if coverageByType[entityType] == nil {
			coverageByType[entityType] = make(map[string]int)
		}
		if coverage.IsCovered {
			coverageByType[entityType]["covered"]++
		} else {
			coverageByType[entityType]["uncovered"]++
		}

		// Track coverage by file
		filePath := entity.FilePath
		if coverageByFile[filePath] == nil {
			coverageByFile[filePath] = make(map[string]int)
		}
		if coverage.IsCovered {
			coverageByFile[filePath]["covered"]++
		} else {
			coverageByFile[filePath]["uncovered"]++
		}
	}

	metrics.TestedEntities = testedEntities
	metrics.UntestedEntities = metrics.TotalEntities - testedEntities

	// Calculate coverage percentage
	if metrics.TotalEntities > 0 {
		metrics.CoveragePercentage = float64(testedEntities) / float64(metrics.TotalEntities) * 100.0
	}

	// Calculate coverage by type percentages
	for entityType, counts := range coverageByType {
		total := counts["covered"] + counts["uncovered"]
		if total > 0 {
			metrics.CoverageByType[entityType] = float64(counts["covered"]) / float64(total) * 100.0
		}
	}

	// Calculate coverage by file percentages
	for filePath, counts := range coverageByFile {
		total := counts["covered"] + counts["uncovered"]
		if total > 0 {
			metrics.CoverageByFile[filePath] = float64(counts["covered"]) / float64(total) * 100.0
		}
	}

	// Add additional details
	metrics.Details["average_assertions_per_test"] = r.calculateAverageAssertions(testEntities)
	metrics.Details["test_to_production_ratio"] = r.calculateTestRatio(len(testEntities), len(productionEntities))
	metrics.Details["coverage_quality_score"] = r.calculateCoverageQuality(metrics)

	return metrics, nil
}

// Helper methods for coverage calculations

// findDirectTests finds tests that directly test the given entity
func (r *BuildGraphResult) findDirectTests(entity *entities.Entity) []*TestInfo {
	tests := make([]*TestInfo, 0)

	// Query for direct TESTS relationships
	query := fmt.Sprintf(`
		MATCH (test)-[rel:TESTS]->(target {id: "%s"})
		RETURN test.id, test.name, test.test_type, test.test_framework, test.file_path, test.assertion_count, rel.confidence_score
	`, entity.ID)

	result, err := r.QueryGraph(query)
	if err != nil {
		return tests
	}

	if result != "" {
		lines := strings.Split(strings.TrimSpace(result), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "\t|\t")
			if len(parts) >= 6 {
				testInfo := &TestInfo{
					TestID:        strings.Trim(parts[0], "\""),
					TestName:      strings.Trim(parts[1], "\""),
					TestType:      strings.Trim(parts[2], "\""),
					TestFramework: strings.Trim(parts[3], "\""),
					FilePath:      strings.Trim(parts[4], "\""),
					CoverageType:  "direct",
				}

				// Parse assertion count
				if len(parts) >= 6 {
					if count, err := fmt.Sscanf(parts[5], "%d", &testInfo.AssertionCount); err != nil || count != 1 {
						testInfo.AssertionCount = 0
					}
				}

				// Parse confidence score
				if len(parts) >= 7 {
					if score, err := fmt.Sscanf(parts[6], "%f", &testInfo.ConfidenceScore); err != nil || score != 1 {
						testInfo.ConfidenceScore = 0.5
					}
				}

				tests = append(tests, testInfo)
			}
		}
	}

	return tests
}

// findIndirectTests finds tests that indirectly test the given entity
func (r *BuildGraphResult) findIndirectTests(entity *entities.Entity) []*TestInfo {
	tests := make([]*TestInfo, 0)

	// Query for indirect COVERS relationships
	query := fmt.Sprintf(`
		MATCH (test)-[rel:COVERS]->(target {id: "%s"})
		WHERE rel.coverage_type = "indirect"
		RETURN test.id, test.name, test.test_type, test.test_framework, test.file_path, test.assertion_count, rel.coverage_type
	`, entity.ID)

	result, err := r.QueryGraph(query)
	if err != nil {
		return tests
	}

	if result != "" {
		lines := strings.Split(strings.TrimSpace(result), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "\t|\t")
			if len(parts) >= 6 {
				testInfo := &TestInfo{
					TestID:          strings.Trim(parts[0], "\""),
					TestName:        strings.Trim(parts[1], "\""),
					TestType:        strings.Trim(parts[2], "\""),
					TestFramework:   strings.Trim(parts[3], "\""),
					FilePath:        strings.Trim(parts[4], "\""),
					ConfidenceScore: 0.3, // Lower confidence for indirect tests
				}

				// Parse assertion count
				if len(parts) >= 6 {
					if count, err := fmt.Sscanf(parts[5], "%d", &testInfo.AssertionCount); err != nil || count != 1 {
						testInfo.AssertionCount = 0
					}
				}

				// Set coverage type
				if len(parts) >= 7 {
					testInfo.CoverageType = strings.Trim(parts[6], "\"")
				}

				tests = append(tests, testInfo)
			}
		}
	}

	return tests
}

// isProductionEntity determines if an entity is production code (not test or utility)
func (r *BuildGraphResult) isProductionEntity(entity *entities.Entity) bool {
	if entity == nil {
		return false
	}

	// Skip test entities
	if entity.IsTest() {
		return false
	}

	// Skip entities in test files
	if entity.IsTestFile() {
		return false
	}

	// Include functions, methods, classes, structs, interfaces
	productionTypes := map[entities.EntityType]bool{
		entities.EntityTypeFunction:  true,
		entities.EntityTypeMethod:    true,
		entities.EntityTypeClass:     true,
		entities.EntityTypeStruct:    true,
		entities.EntityTypeInterface: true,
	}

	return productionTypes[entity.Type]
}

// calculateCoverageScore calculates a coverage score from 0.0 to 1.0
func (r *BuildGraphResult) calculateCoverageScore(directTests, indirectTests []*TestInfo) float64 {
	if len(directTests) == 0 && len(indirectTests) == 0 {
		return 0.0
	}

	score := 0.0

	// Direct tests contribute more to the score
	for _, test := range directTests {
		testScore := 0.6 * test.ConfidenceScore

		// Bonus for assertion count
		if test.AssertionCount > 0 {
			testScore += 0.2 * float64(minInt(test.AssertionCount, 5)) / 5.0
		}

		// Bonus for test type
		if test.TestType == "unit" {
			testScore += 0.1
		} else if test.TestType == "integration" {
			testScore += 0.15
		}

		score += testScore
	}

	// Indirect tests contribute less
	for _, test := range indirectTests {
		score += 0.3 * test.ConfidenceScore
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// hasTestType checks if any test in the list has the specified type
func (r *BuildGraphResult) hasTestType(tests []*TestInfo, testType string) bool {
	for _, test := range tests {
		if test.TestType == testType {
			return true
		}
	}
	return false
}

// calculateConfidenceScore calculates the overall confidence in test coverage
func (r *BuildGraphResult) calculateConfidenceScore(directTests, indirectTests []*TestInfo) float64 {
	if len(directTests) == 0 && len(indirectTests) == 0 {
		return 0.0
	}

	totalScore := 0.0
	totalTests := 0

	for _, test := range directTests {
		totalScore += test.ConfidenceScore * 1.0 // Full weight for direct tests
		totalTests++
	}

	for _, test := range indirectTests {
		totalScore += test.ConfidenceScore * 0.5 // Half weight for indirect tests
		totalTests++
	}

	if totalTests == 0 {
		return 0.0
	}

	return totalScore / float64(totalTests)
}

// calculateAverageAssertions calculates the average number of assertions per test
func (r *BuildGraphResult) calculateAverageAssertions(testEntities []*entities.Entity) float64 {
	if len(testEntities) == 0 {
		return 0.0
	}

	totalAssertions := 0
	for _, entity := range testEntities {
		totalAssertions += entity.GetAssertionCount()
	}

	return float64(totalAssertions) / float64(len(testEntities))
}

// calculateTestRatio calculates the ratio of test entities to production entities
func (r *BuildGraphResult) calculateTestRatio(testCount, productionCount int) float64 {
	if productionCount == 0 {
		return 0.0
	}

	return float64(testCount) / float64(productionCount)
}

// calculateCoverageQuality calculates an overall quality score for test coverage
func (r *BuildGraphResult) calculateCoverageQuality(metrics *CoverageMetrics) float64 {
	score := 0.0

	// Coverage percentage (40% of score)
	score += (metrics.CoveragePercentage / 100.0) * 0.4

	// Test ratio (20% of score)
	if metrics.TotalEntities > 0 {
		ratio := float64(metrics.TestCount) / float64(metrics.TotalEntities)
		score += minFloat(ratio, 1.0) * 0.2
	}

	// Average assertions per test (20% of score)
	if avgAssertions, ok := metrics.Details["average_assertions_per_test"].(float64); ok {
		score += minFloat(avgAssertions/3.0, 1.0) * 0.2 // Target: 3 assertions per test
	}

	// Framework diversity (10% of score)
	frameworkCount := len(metrics.TestFrameworks)
	score += minFloat(float64(frameworkCount)/2.0, 1.0) * 0.1

	// Test type diversity (10% of score)
	typeCount := len(metrics.TestTypes)
	score += minFloat(float64(typeCount)/3.0, 1.0) * 0.1

	return score
}

// min returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// minFloat returns the minimum of two float64 values
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
