// Package db provides KuzuDB integration for storing and querying code graph data.
//
// This package encapsulates all database operations, schema management, and
// data persistence logic for the code analysis system. It uses KuzuDB as the
// embedded graph database to store entities (nodes) and relationships (edges)
// discovered during code analysis.
//
// Key Components:
//   - Database connection management with proper cleanup
//   - Schema creation and maintenance for code entities
//   - Entity and relationship storage with type safety
//   - Cypher query execution and result processing
//   - Schema introspection and metadata access
//
// Database Schema:
//   Node Types: File, Function, Class, Method, Struct, Interface, Import, Variable
//   Relationship Types: Contains, CALLS, IMPORTS, INHERITS, EMBEDS, IMPLEMENTS, DEFINES, USES
//
// Example usage:
//
//	database, err := db.NewKuzuDatabase("./analysis.db")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer database.Close()
//
//	err = database.CreateSchema()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Store entities and relationships
//	err = database.StoreEntity(entity)
//	if err != nil {
//		log.Printf("Failed to store entity: %v", err)
//	}
//
// Performance Notes:
//   - KuzuDB is optimized for read-heavy workloads typical in code analysis
//   - Batch operations are recommended for large datasets
//   - Prepared statements are used for type safety and performance
//   - Connection pooling is handled internally by KuzuDB
package db

import (
	"fmt"
	"strings"

	"github.com/onyx/onyx-tui/graph_service/internal/entities"

	"github.com/kuzudb/go-kuzu"
)

// KuzuDatabase encapsulates a KuzuDB embedded graph database instance and
// provides high-level operations for code graph storage and retrieval.
//
// This structure maintains both the database instance and an active connection,
// ensuring proper resource management and transaction handling. All operations
// are thread-safe when used through the provided methods.
//
// The database uses KuzuDB's columnar storage format optimized for analytical
// queries typical in code analysis workloads. It supports full Cypher query
// language for complex graph traversals and pattern matching.
//
// Lifecycle management:
//   1. Create with NewKuzuDatabase()
//   2. Initialize schema with CreateSchema()
//   3. Store entities and relationships
//   4. Query data with ExecuteQuery()
//   5. Clean up with Close()
//
// Thread Safety:
//   - Multiple goroutines can safely call query methods
//   - Write operations (StoreEntity, StoreRelationship) should be serialized
//   - Connection pooling is handled internally by KuzuDB
type KuzuDatabase struct {
	// DB holds the KuzuDB database instance. This provides access to
	// database-level operations like configuration and metadata.
	DB *kuzu.Database

	// Connection maintains an active connection to the database for
	// executing queries and transactions. KuzuDB supports multiple
	// concurrent connections to the same database.
	Connection *kuzu.Connection
}

// NewKuzuDatabase creates and initializes a new KuzuDB embedded database instance
// at the specified filesystem path.
//
// This function:
//   1. Creates the database directory if it doesn't exist
//   2. Initializes KuzuDB with default system configuration
//   3. Establishes an active connection for immediate use
//   4. Returns a ready-to-use KuzuDatabase instance
//
// The database files are stored persistently at the given path and can be
// reopened later to restore the complete graph state. The path should be
// a directory that the application has write access to.
//
// Parameters:
//   - dbPath: Filesystem path where database files will be stored.
//     Can be relative ("./analysis.db") or absolute ("/tmp/analysis.db").
//     The directory will be created if it doesn't exist.
//
// Returns:
//   - *KuzuDatabase: Initialized database instance ready for schema creation
//     and data operations. Must be closed with Close() when finished.
//   - error: Non-nil if database creation or connection fails due to:
//     * Insufficient filesystem permissions
//     * Invalid or inaccessible path
//     * KuzuDB initialization errors
//     * Resource constraints (disk space, memory)
//
// Example usage:
//
//	// Create in current directory
//	db, err := NewKuzuDatabase("./code-analysis.db")
//	if err != nil {
//		return fmt.Errorf("database creation failed: %w", err)
//	}
//	defer db.Close()
//
//	// Create in temporary directory
//	tmpDir, _ := os.MkdirTemp("", "analysis_*")
//	db, err := NewKuzuDatabase(tmpDir)
//	if err != nil {
//		return fmt.Errorf("database creation failed: %w", err)
//	}
//	defer func() {
//		db.Close()
//		os.RemoveAll(tmpDir)
//	}()
//
// Performance Considerations:
//   - Database creation is an expensive operation; reuse instances when possible
//   - Initial database size is minimal but grows with entity/relationship count
//   - KuzuDB uses memory-mapped files for efficient data access
//   - Consider SSD storage for better query performance on large graphs
func NewKuzuDatabase(dbPath string) (*KuzuDatabase, error) {
	// Open a database with default system configuration.
	systemConfig := kuzu.DefaultSystemConfig()
	db, err := kuzu.OpenDatabase(dbPath, systemConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Open a connection to the database.
	conn, err := kuzu.OpenConnection(db)
	if err != nil {
		db.Close() // Clean up the database if connection fails
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	fmt.Println("Successfully connected to KuzuDB.")

	return &KuzuDatabase{DB: db, Connection: conn}, nil
}

// Close cleans up and closes the database connection.
func (kdb *KuzuDatabase) Close() {
	if kdb.Connection != nil {
		kdb.Connection.Close()
	}
	if kdb.DB != nil {
		kdb.DB.Close()
	}
	fmt.Println("KuzuDB connection closed.")
}

// CreateSchema creates the necessary node and relationship tables for the code graph.
func (kdb *KuzuDatabase) CreateSchema() error {
	queries := []string{
		// Basic entity types
		`CREATE NODE TABLE IF NOT EXISTS File(path STRING, name STRING, language STRING, PRIMARY KEY (path))`,
		`CREATE NODE TABLE IF NOT EXISTS Function(id STRING, name STRING, signature STRING, body STRING, file_path STRING, PRIMARY KEY (id))`,
		`CREATE NODE TABLE IF NOT EXISTS Class(id STRING, name STRING, signature STRING, file_path STRING, PRIMARY KEY (id))`,

		// Enhanced Go-specific entity types
		`CREATE NODE TABLE IF NOT EXISTS Method(id STRING, name STRING, signature STRING, body STRING, receiver_type STRING, file_path STRING, PRIMARY KEY (id))`,
		`CREATE NODE TABLE IF NOT EXISTS Struct(id STRING, name STRING, type_definition STRING, file_path STRING, PRIMARY KEY (id))`,
		`CREATE NODE TABLE IF NOT EXISTS Interface(id STRING, name STRING, type_definition STRING, file_path STRING, PRIMARY KEY (id))`,
		`CREATE NODE TABLE IF NOT EXISTS Import(id STRING, name STRING, path STRING, alias STRING, file_path STRING, PRIMARY KEY (id))`,
		`CREATE NODE TABLE IF NOT EXISTS Variable(id STRING, name STRING, type STRING, value STRING, file_path STRING, PRIMARY KEY (id))`,

		// Test-specific entity types
		`CREATE NODE TABLE IF NOT EXISTS TestFunction(id STRING, name STRING, signature STRING, body STRING, file_path STRING, test_type STRING, test_target STRING, assertion_count INT64, test_framework STRING, PRIMARY KEY (id))`,
		`CREATE NODE TABLE IF NOT EXISTS TestCase(id STRING, name STRING, signature STRING, body STRING, file_path STRING, test_type STRING, test_target STRING, assertion_count INT64, test_framework STRING, PRIMARY KEY (id))`,
		`CREATE NODE TABLE IF NOT EXISTS TestSuite(id STRING, name STRING, signature STRING, file_path STRING, test_type STRING, test_framework STRING, test_count INT64, PRIMARY KEY (id))`,
		`CREATE NODE TABLE IF NOT EXISTS Assertion(id STRING, name STRING, assertion_type STRING, expected_value STRING, actual_value STRING, file_path STRING, PRIMARY KEY (id))`,
		`CREATE NODE TABLE IF NOT EXISTS Mock(id STRING, name STRING, mock_type STRING, target_entity STRING, file_path STRING, PRIMARY KEY (id))`,
		`CREATE NODE TABLE IF NOT EXISTS Fixture(id STRING, name STRING, fixture_type STRING, data_content STRING, file_path STRING, PRIMARY KEY (id))`,

		// Basic relationships
		`CREATE REL TABLE IF NOT EXISTS Contains(FROM File TO Function, FROM File TO Class, FROM File TO Method, FROM File TO Struct, FROM File TO Interface, FROM File TO Import, FROM File TO Variable, FROM File TO TestFunction, FROM File TO TestCase, FROM File TO TestSuite, FROM File TO Assertion, FROM File TO Mock, FROM File TO Fixture)`,
		`CREATE REL TABLE IF NOT EXISTS CALLS(FROM Function TO Function, FROM Method TO Function, FROM Function TO Method, FROM Method TO Method, FROM TestFunction TO Function, FROM TestFunction TO Method, FROM TestCase TO Function, FROM TestCase TO Method)`,
		`CREATE REL TABLE IF NOT EXISTS IMPORTS(FROM File TO File)`,
		`CREATE REL TABLE IF NOT EXISTS INHERITS(FROM Class TO Class)`,

		// Enhanced Go-specific relationships
		`CREATE REL TABLE IF NOT EXISTS EMBEDS(FROM Struct TO Struct, source_id STRING, target_id STRING)`,
		`CREATE REL TABLE IF NOT EXISTS IMPLEMENTS(FROM Struct TO Interface, source_id STRING, target_id STRING)`,
		`CREATE REL TABLE IF NOT EXISTS DEFINES(FROM Struct TO Method, FROM Interface TO Method)`,
		`CREATE REL TABLE IF NOT EXISTS USES(FROM Function TO Struct, FROM Method TO Struct, FROM Function TO Interface, FROM Method TO Interface)`,

		// Test Coverage relationships
		`CREATE REL TABLE IF NOT EXISTS TESTS(FROM TestFunction TO Function, FROM TestFunction TO Method, FROM TestFunction TO Class, FROM TestCase TO Function, FROM TestCase TO Method, FROM TestCase TO Class, confidence_score DOUBLE)`,
		`CREATE REL TABLE IF NOT EXISTS COVERS(FROM TestFunction TO Function, FROM TestFunction TO Method, FROM TestCase TO Function, FROM TestCase TO Method, coverage_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS MOCKS(FROM TestFunction TO Function, FROM TestFunction TO Method, FROM TestCase TO Function, FROM TestCase TO Method, FROM Mock TO Function, FROM Mock TO Method, mock_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS SETUP_FOR(FROM TestFunction TO TestCase, FROM Fixture TO TestFunction, FROM Fixture TO TestCase)`,
		`CREATE REL TABLE IF NOT EXISTS TEARDOWN_FOR(FROM TestFunction TO TestCase)`,
		`CREATE REL TABLE IF NOT EXISTS ASSERTS(FROM TestFunction TO Assertion, FROM TestCase TO Assertion, assertion_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS VERIFIES(FROM TestFunction TO Function, FROM TestFunction TO Method, FROM TestCase TO Function, FROM TestCase TO Method, verification_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS SPIES(FROM TestFunction TO Function, FROM TestFunction TO Method, FROM TestCase TO Function, FROM TestCase TO Method, spy_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS STUBS(FROM TestFunction TO Function, FROM TestFunction TO Method, FROM TestCase TO Function, FROM TestCase TO Method, stub_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS FIXTURES(FROM TestFunction TO Fixture, FROM TestCase TO Fixture, fixture_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS RUNS_TEST(FROM TestSuite TO TestFunction, FROM TestSuite TO TestCase)`,
		`CREATE REL TABLE IF NOT EXISTS GROUPS_TESTS(FROM TestSuite TO TestFunction, FROM TestSuite TO TestCase, group_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS SKIPS(FROM TestFunction TO TestFunction, FROM TestCase TO TestCase, skip_condition STRING)`,
		`CREATE REL TABLE IF NOT EXISTS DEPENDS(FROM TestFunction TO TestFunction, FROM TestCase TO TestCase, FROM TestFunction TO Fixture, FROM TestCase TO Fixture, dependency_type STRING)`,

		// TypeScript-specific relationships
		`CREATE REL TABLE IF NOT EXISTS RE_EXPORTS(FROM File TO File, FROM Function TO Function, FROM Class TO Class, export_name STRING, export_alias STRING)`,
		`CREATE REL TABLE IF NOT EXISTS DECORATES(FROM Function TO Class, FROM Function TO Method, FROM Function TO Function, decorator_name STRING, decorator_params STRING)`,
		`CREATE REL TABLE IF NOT EXISTS CONSTRAINS(FROM Interface TO Class, FROM Interface TO Function, constraint_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS DYNAMIC_IMPORT(FROM Function TO File, FROM Method TO File, import_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS HAS_PROPS(FROM Class TO Variable, FROM Interface TO Variable, prop_type STRING, is_required BOOLEAN)`,
		`CREATE REL TABLE IF NOT EXISTS RENDERS_JSX(FROM Function TO Class, FROM Method TO Class, jsx_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS INJECTS(FROM Class TO Class, FROM Function TO Variable, injection_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS CONSUMES_SERVICE(FROM Class TO Class, FROM Function TO Class, service_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS HANDLES_ROUTE(FROM Function TO Variable, FROM Method TO Variable, route_path STRING, http_method STRING)`,
		`CREATE REL TABLE IF NOT EXISTS CALLS_API(FROM Function TO Function, FROM Method TO Function, api_endpoint STRING, http_method STRING)`,
		`CREATE REL TABLE IF NOT EXISTS EXPOSES_ENDPOINT(FROM Function TO Variable, FROM Method TO Variable, endpoint_path STRING, endpoint_type STRING)`,
		`CREATE REL TABLE IF NOT EXISTS USES_MIDDLEWARE(FROM Function TO Function, FROM Class TO Function, middleware_type STRING)`,
	}

	fmt.Println("Initializing database schema...")
	for _, query := range queries {
		_, err := kdb.Connection.Query(query)
		if err != nil {
			return fmt.Errorf("failed to execute schema query '%s': %w", query, err)
		}
	}

	fmt.Println("Database schema initialized successfully.")
	return nil
}

// executePreparedStatement is a helper to prepare and execute a query with parameters.
func (kdb *KuzuDatabase) executePreparedStatement(query string, params map[string]interface{}) error {
	stmt, err := kdb.Connection.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	result, err := kdb.Connection.Execute(stmt, params)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}
	result.Close()
	return nil
}

// AddFileNode adds a new File node to the graph.
func (kdb *KuzuDatabase) AddFileNode(path, name, language string) error {
	query := "CREATE (f:File {path: $path, name: $name, language: $language})"
	params := map[string]interface{}{
		"path":     path,
		"name":     name,
		"language": language,
	}
	return kdb.executePreparedStatement(query, params)
}

// AddClassNode adds a class node to the graph.
func (kdb *KuzuDatabase) AddClassNode(id, name, signature, filePath string) error {
	query := `CREATE (c:Class {id: $id, name: $name, signature: $signature, file_path: $file_path})`
	params := map[string]interface{}{
		"id":        id,
		"name":      name,
		"signature": signature,
		"file_path": filePath,
	}
	return kdb.executePreparedStatement(query, params)
}

// AddFunctionNode adds a function node to the graph.
func (kdb *KuzuDatabase) AddFunctionNode(id, name, signature, body, filePath string) error {
	query := `CREATE (f:Function {id: $id, name: $name, signature: $signature, body: $body, file_path: $file_path})`
	params := map[string]interface{}{
		"id":        id,
		"name":      name,
		"signature": signature,
		"body":      body,
		"file_path": filePath,
	}
	return kdb.executePreparedStatement(query, params)
}

// AddContainsRel creates a CONTAINS relationship between a file and a function/class.
func (kdb *KuzuDatabase) AddContainsRel(filePath, childId, childType string) error {
	query := fmt.Sprintf(`
		MATCH (f:File {path: $file_path})
		MATCH (c:%s {id: $child_id})
		CREATE (f)-[:Contains]->(c)
	`, childType)

	params := map[string]interface{}{
		"file_path": filePath,
		"child_id":  childId,
	}
	return kdb.executePreparedStatement(query, params)
}

// AddCallsRel creates a CALLS relationship between two Functions.
func (kdb *KuzuDatabase) AddCallsRel(callerId, calleeId string) error {
	query := "MATCH (caller:Function {id: $callerId}), (callee:Function {id: $calleeId}) CREATE (caller)-[:CALLS]->(callee)"
	params := map[string]interface{}{
		"callerId": callerId,
		"calleeId": calleeId,
	}
	return kdb.executePreparedStatement(query, params)
}

// AddImportsRel creates an IMPORTS relationship between two Files.
func (kdb *KuzuDatabase) AddImportsRel(importerPath, importeePath string) error {
	query := "MATCH (importer:File {path: $importerPath}), (importee:File {path: $importeePath}) CREATE (importer)-[:IMPORTS]->(importee)"
	params := map[string]interface{}{
		"importerPath": importerPath,
		"importeePath": importeePath,
	}
	return kdb.executePreparedStatement(query, params)
}

// AddInheritsRel creates an INHERITS relationship between two Classes.
func (kdb *KuzuDatabase) AddInheritsRel(childId, parentName string) error {
	// First try to find parent class by name (within same file or project)
	query := `
		MATCH (child:Class {id: $childId})
		MATCH (parent:Class {name: $parentName})
		CREATE (child)-[:INHERITS]->(parent)
	`
	params := map[string]interface{}{
		"childId":    childId,
		"parentName": parentName,
	}
	return kdb.executePreparedStatement(query, params)
}

// ExecuteQuery executes a query and returns the result as a string.
func (kdb *KuzuDatabase) ExecuteQuery(query string) (string, error) {
	result, err := kdb.Connection.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %w", err)
	}
	defer result.Close()

	var resultBuilder strings.Builder
	for result.HasNext() {
		tuple, err := result.Next()
		if err != nil {
			return "", fmt.Errorf("failed to get next tuple: %w", err)
		}
		var rowItems []string
		for i := uint64(0); ; i++ {
			val, err := tuple.GetValue(i)
			if err != nil {
				// Assumes error means we are out of bounds.
				break
			}
			rowItems = append(rowItems, fmt.Sprintf("%v", val))
		}
		resultBuilder.WriteString(strings.Join(rowItems, "\t|\t"))
		resultBuilder.WriteString("\n")
	}

	return resultBuilder.String(), nil
}

// GetSchema returns the database schema as a formatted string.
func (kdb *KuzuDatabase) GetSchema() (string, error) {
	query := `CALL SHOW_TABLES() RETURN *`
	result, err := kdb.Connection.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to get schema: %w", err)
	}
	defer result.Close()

	var schemaBuilder strings.Builder
	for result.HasNext() {
		tuple, err := result.Next()
		if err != nil {
			return "", fmt.Errorf("failed to get next schema tuple: %w", err)
		}
		var rowItems []string
		for i := uint64(0); ; i++ {
			val, err := tuple.GetValue(i)
			if err != nil {
				// Assumes error means we are out of bounds.
				break
			}
			rowItems = append(rowItems, fmt.Sprintf("%v", val))
		}
		schemaBuilder.WriteString(strings.Join(rowItems, "\t|\t"))
		schemaBuilder.WriteString("\n")
	}

	return schemaBuilder.String(), nil
}

// StoreEntity stores an entity in the database
func (kdb *KuzuDatabase) StoreEntity(entity *entities.Entity) error {
	var query string

	// Escape quotes in strings for safe SQL
	safeName := strings.ReplaceAll(entity.Name, "\"", "\\\"")
	safeSignature := strings.ReplaceAll(entity.Signature, "\"", "\\\"")
	safeBody := strings.ReplaceAll(entity.Body, "\"", "\\\"")
	safeFilePath := strings.ReplaceAll(entity.FilePath, "\"", "\\\"")

	switch entity.Type {
	case entities.EntityTypeFunction:
		query = fmt.Sprintf(`CREATE (f:Function {id: "%s", name: "%s", signature: "%s", body: "%s", file_path: "%s"})`,
			entity.ID, safeName, safeSignature, safeBody, safeFilePath)
	case entities.EntityTypeMethod:
		receiverType := entity.GetProperty("receiver_type")
		if receiverType == nil {
			receiverType = ""
		}
		query = fmt.Sprintf(`CREATE (m:Method {id: "%s", name: "%s", signature: "%s", body: "%s", receiver_type: "%s", file_path: "%s"})`,
			entity.ID, safeName, safeSignature, safeBody, receiverType, safeFilePath)
	case entities.EntityTypeClass:
		query = fmt.Sprintf(`CREATE (c:Class {id: "%s", name: "%s", signature: "%s", file_path: "%s"})`,
			entity.ID, safeName, safeSignature, safeFilePath)
	case entities.EntityTypeStruct:
		typeDef := entity.GetProperty("type_definition")
		if typeDef == nil {
			typeDef = ""
		}
		safeTypeDef := strings.ReplaceAll(fmt.Sprintf("%v", typeDef), "\"", "\\\"")
		query = fmt.Sprintf(`CREATE (s:Struct {id: "%s", name: "%s", type_definition: "%s", file_path: "%s"})`,
			entity.ID, safeName, safeTypeDef, safeFilePath)
	case entities.EntityTypeInterface:
		typeDef := entity.GetProperty("type_definition")
		if typeDef == nil {
			typeDef = ""
		}
		safeTypeDef := strings.ReplaceAll(fmt.Sprintf("%v", typeDef), "\"", "\\\"")
		query = fmt.Sprintf(`CREATE (i:Interface {id: "%s", name: "%s", type_definition: "%s", file_path: "%s"})`,
			entity.ID, safeName, safeTypeDef, safeFilePath)
	case entities.EntityTypeImport:
		path := entity.GetProperty("path")
		alias := entity.GetProperty("alias")
		if path == nil {
			path = ""
		}
		if alias == nil {
			alias = ""
		}
		safePath := strings.ReplaceAll(fmt.Sprintf("%v", path), "\"", "\\\"")
		safeAlias := strings.ReplaceAll(fmt.Sprintf("%v", alias), "\"", "\\\"")
		query = fmt.Sprintf(`CREATE (imp:Import {id: "%s", name: "%s", path: "%s", alias: "%s", file_path: "%s"})`,
			entity.ID, safeName, safePath, safeAlias, safeFilePath)
	case entities.EntityTypeVariable:
		varType := entity.GetProperty("type")
		value := entity.GetProperty("value")
		if varType == nil {
			varType = ""
		}
		if value == nil {
			value = ""
		}
		safeType := strings.ReplaceAll(fmt.Sprintf("%v", varType), "\"", "\\\"")
		safeValue := strings.ReplaceAll(fmt.Sprintf("%v", value), "\"", "\\\"")
		query = fmt.Sprintf(`CREATE (v:Variable {id: "%s", name: "%s", type: "%s", value: "%s", file_path: "%s"})`,
			entity.ID, safeName, safeType, safeValue, safeFilePath)
	
	// Test entity types
	case entities.EntityTypeTestFunction:
		testType := entity.GetTestType()
		testTarget := entity.GetTestTarget()
		assertionCount := entity.GetAssertionCount()
		testFramework := entity.GetTestFramework()
		query = fmt.Sprintf(`CREATE (tf:TestFunction {id: "%s", name: "%s", signature: "%s", body: "%s", file_path: "%s", test_type: "%s", test_target: "%s", assertion_count: %d, test_framework: "%s"})`,
			entity.ID, safeName, safeSignature, safeBody, safeFilePath, testType, testTarget, assertionCount, testFramework)
	
	case entities.EntityTypeTestCase:
		testType := entity.GetTestType()
		testTarget := entity.GetTestTarget()
		assertionCount := entity.GetAssertionCount()
		testFramework := entity.GetTestFramework()
		query = fmt.Sprintf(`CREATE (tc:TestCase {id: "%s", name: "%s", signature: "%s", body: "%s", file_path: "%s", test_type: "%s", test_target: "%s", assertion_count: %d, test_framework: "%s"})`,
			entity.ID, safeName, safeSignature, safeBody, safeFilePath, testType, testTarget, assertionCount, testFramework)
	
	case entities.EntityTypeTestSuite:
		testType := entity.GetTestType()
		testFramework := entity.GetTestFramework()
		testCount := int64(0)
		if tc := entity.GetProperty("test_count"); tc != nil {
			if count, ok := tc.(int); ok {
				testCount = int64(count)
			}
		}
		query = fmt.Sprintf(`CREATE (ts:TestSuite {id: "%s", name: "%s", signature: "%s", file_path: "%s", test_type: "%s", test_framework: "%s", test_count: %d})`,
			entity.ID, safeName, safeSignature, safeFilePath, testType, testFramework, testCount)
	
	case entities.EntityTypeAssertion:
		assertionType := ""
		expectedValue := ""
		actualValue := ""
		if at := entity.GetProperty("assertion_type"); at != nil {
			assertionType = fmt.Sprintf("%v", at)
		}
		if ev := entity.GetProperty("expected_value"); ev != nil {
			expectedValue = fmt.Sprintf("%v", ev)
		}
		if av := entity.GetProperty("actual_value"); av != nil {
			actualValue = fmt.Sprintf("%v", av)
		}
		safeAssertionType := strings.ReplaceAll(assertionType, "\"", "\\\"")
		safeExpectedValue := strings.ReplaceAll(expectedValue, "\"", "\\\"")
		safeActualValue := strings.ReplaceAll(actualValue, "\"", "\\\"")
		query = fmt.Sprintf(`CREATE (a:Assertion {id: "%s", name: "%s", assertion_type: "%s", expected_value: "%s", actual_value: "%s", file_path: "%s"})`,
			entity.ID, safeName, safeAssertionType, safeExpectedValue, safeActualValue, safeFilePath)
	
	case entities.EntityTypeMock:
		mockType := ""
		targetEntity := ""
		if mt := entity.GetProperty("mock_type"); mt != nil {
			mockType = fmt.Sprintf("%v", mt)
		}
		if te := entity.GetProperty("target_entity"); te != nil {
			targetEntity = fmt.Sprintf("%v", te)
		}
		safeMockType := strings.ReplaceAll(mockType, "\"", "\\\"")
		safeTargetEntity := strings.ReplaceAll(targetEntity, "\"", "\\\"")
		query = fmt.Sprintf(`CREATE (m:Mock {id: "%s", name: "%s", mock_type: "%s", target_entity: "%s", file_path: "%s"})`,
			entity.ID, safeName, safeMockType, safeTargetEntity, safeFilePath)
	
	case entities.EntityTypeFixture:
		fixtureType := ""
		dataContent := ""
		if ft := entity.GetProperty("fixture_type"); ft != nil {
			fixtureType = fmt.Sprintf("%v", ft)
		}
		if dc := entity.GetProperty("data_content"); dc != nil {
			dataContent = fmt.Sprintf("%v", dc)
		}
		safeFixtureType := strings.ReplaceAll(fixtureType, "\"", "\\\"")
		safeDataContent := strings.ReplaceAll(dataContent, "\"", "\\\"")
		query = fmt.Sprintf(`CREATE (f:Fixture {id: "%s", name: "%s", fixture_type: "%s", data_content: "%s", file_path: "%s"})`,
			entity.ID, safeName, safeFixtureType, safeDataContent, safeFilePath)
	
	default:
		return fmt.Errorf("unsupported entity type: %s", entity.Type)
	}

	_, err := kdb.Connection.Query(query)
	return err
}

// StoreRelationship stores a relationship in the database using type-aware queries
// This method fixes the "bound by multiple node labels" error by using specific
// node labels in MATCH clauses based on the relationship's entity type metadata.
func (kdb *KuzuDatabase) StoreRelationship(rel *entities.Relationship) error {
	// Validate that we have the necessary type information
	if rel.SourceType == "" || rel.TargetType == "" {
		return fmt.Errorf("relationship missing type information: source=%s, target=%s", 
			rel.SourceType, rel.TargetType)
	}

	// Validate relationship against schema constraints
	if !rel.IsValidForSchema() {
		return fmt.Errorf("relationship violates schema constraints: %s %s -> %s %s", 
			rel.SourceType, rel.Type, rel.TargetType, rel.TargetID)
	}

	switch rel.Type {
	case entities.RelationshipTypeCalls:
		return kdb.storeCAllsRelationship(rel)
	case entities.RelationshipTypeContains:
		return kdb.storeContainsRelationship(rel)
	case entities.RelationshipTypeInherits:
		return kdb.storeInheritsRelationship(rel)
	case entities.RelationshipTypeEmbeds:
		return kdb.storeEmbedsRelationship(rel)
	case entities.RelationshipTypeImplements:
		return kdb.storeImplementsRelationship(rel)
	case entities.RelationshipTypeDefines:
		return kdb.storeDefinesRelationship(rel)
	case entities.RelationshipTypeUses:
		return kdb.storeUsesRelationship(rel)
	
	// Test Coverage relationships
	case entities.RelationshipTypeTests:
		return kdb.storeTestsRelationship(rel)
	case entities.RelationshipTypeCovers:
		return kdb.storeCoversRelationship(rel)
	case entities.RelationshipTypeMocks:
		return kdb.storeMocksRelationship(rel)
	case entities.RelationshipTypeSetupFor:
		return kdb.storeSetupForRelationship(rel)
	case entities.RelationshipTypeTeardownFor:
		return kdb.storeTeardownForRelationship(rel)
	case entities.RelationshipTypeAsserts:
		return kdb.storeAssertsRelationship(rel)
	case entities.RelationshipTypeVerifies:
		return kdb.storeVerifiesRelationship(rel)
	case entities.RelationshipTypeSpies:
		return kdb.storeSpiesRelationship(rel)
	case entities.RelationshipTypeStubs:
		return kdb.storeStubsRelationship(rel)
	case entities.RelationshipTypeFixtures:
		return kdb.storeFixturesRelationship(rel)
	case entities.RelationshipTypeRunsTest:
		return kdb.storeRunsTestRelationship(rel)
	case entities.RelationshipTypeGroupsTests:
		return kdb.storeGroupsTestsRelationship(rel)
	case entities.RelationshipTypeSkips:
		return kdb.storeSkipsRelationship(rel)
	case entities.RelationshipTypeDepends:
		return kdb.storeDependsRelationship(rel)
	
	default:
		return fmt.Errorf("unsupported relationship type: %s", rel.Type)
	}
}

// storeCAllsRelationship stores CALLS relationships with proper type-aware queries
func (kdb *KuzuDatabase) storeCAllsRelationship(rel *entities.Relationship) error {
	// Build type-aware query based on source and target types
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:CALLS]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store CALLS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeContainsRelationship stores Contains relationships with proper type-aware queries
func (kdb *KuzuDatabase) storeContainsRelationship(rel *entities.Relationship) error {
	// Contains relationships are typically File -> Entity
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:Contains]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store Contains relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeInheritsRelationship stores INHERITS relationships with proper type-aware queries
func (kdb *KuzuDatabase) storeInheritsRelationship(rel *entities.Relationship) error {
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:INHERITS]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store INHERITS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeEmbedsRelationship stores EMBEDS relationships with proper type-aware queries
func (kdb *KuzuDatabase) storeEmbedsRelationship(rel *entities.Relationship) error {
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:EMBEDS {source_id: "%s", target_id: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, rel.SourceID, rel.TargetID)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store EMBEDS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeImplementsRelationship stores IMPLEMENTS relationships with proper type-aware queries
func (kdb *KuzuDatabase) storeImplementsRelationship(rel *entities.Relationship) error {
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:IMPLEMENTS {source_id: "%s", target_id: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, rel.SourceID, rel.TargetID)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store IMPLEMENTS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeDefinesRelationship stores DEFINES relationships with proper type-aware queries
func (kdb *KuzuDatabase) storeDefinesRelationship(rel *entities.Relationship) error {
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:DEFINES]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store DEFINES relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeUsesRelationship stores USES relationships with proper type-aware queries
func (kdb *KuzuDatabase) storeUsesRelationship(rel *entities.Relationship) error {
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:USES]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store USES relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// Test Coverage relationship storage methods

// storeTestsRelationship stores TESTS relationships with confidence score
func (kdb *KuzuDatabase) storeTestsRelationship(rel *entities.Relationship) error {
	confidenceScore := rel.GetConfidenceScore()
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:TESTS {confidence_score: %f}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, confidenceScore)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store TESTS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeCoversRelationship stores COVERS relationships with coverage type
func (kdb *KuzuDatabase) storeCoversRelationship(rel *entities.Relationship) error {
	coverageType := rel.GetCoverageType()
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:COVERS {coverage_type: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, coverageType)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store COVERS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeMocksRelationship stores MOCKS relationships with mock type
func (kdb *KuzuDatabase) storeMocksRelationship(rel *entities.Relationship) error {
	mockType := rel.GetMockType()
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:MOCKS {mock_type: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, mockType)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store MOCKS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// Remaining test relationship storage methods

// storeSetupForRelationship stores SETUP_FOR relationships
func (kdb *KuzuDatabase) storeSetupForRelationship(rel *entities.Relationship) error {
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:SETUP_FOR]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store SETUP_FOR relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeTeardownForRelationship stores TEARDOWN_FOR relationships
func (kdb *KuzuDatabase) storeTeardownForRelationship(rel *entities.Relationship) error {
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:TEARDOWN_FOR]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store TEARDOWN_FOR relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeAssertsRelationship stores ASSERTS relationships with assertion type
func (kdb *KuzuDatabase) storeAssertsRelationship(rel *entities.Relationship) error {
	assertionType := rel.GetAssertionType()
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:ASSERTS {assertion_type: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, assertionType)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store ASSERTS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeVerifiesRelationship stores VERIFIES relationships with verification type
func (kdb *KuzuDatabase) storeVerifiesRelationship(rel *entities.Relationship) error {
	verificationType := ""
	if vt := rel.GetProperty("verification_type"); vt != nil {
		verificationType = fmt.Sprintf("%v", vt)
	}
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:VERIFIES {verification_type: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, verificationType)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store VERIFIES relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeSpiesRelationship stores SPIES relationships with spy type  
func (kdb *KuzuDatabase) storeSpiesRelationship(rel *entities.Relationship) error {
	spyType := ""
	if st := rel.GetProperty("spy_type"); st != nil {
		spyType = fmt.Sprintf("%v", st)
	}
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:SPIES {spy_type: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, spyType)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store SPIES relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeStubsRelationship stores STUBS relationships with stub type
func (kdb *KuzuDatabase) storeStubsRelationship(rel *entities.Relationship) error {
	stubType := ""
	if st := rel.GetProperty("stub_type"); st != nil {
		stubType = fmt.Sprintf("%v", st)
	}
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:STUBS {stub_type: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, stubType)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store STUBS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeFixturesRelationship stores FIXTURES relationships with fixture type
func (kdb *KuzuDatabase) storeFixturesRelationship(rel *entities.Relationship) error {
	fixtureType := ""
	if ft := rel.GetProperty("fixture_type"); ft != nil {
		fixtureType = fmt.Sprintf("%v", ft)
	}
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:FIXTURES {fixture_type: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, fixtureType)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store FIXTURES relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeRunsTestRelationship stores RUNS_TEST relationships
func (kdb *KuzuDatabase) storeRunsTestRelationship(rel *entities.Relationship) error {
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:RUNS_TEST]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store RUNS_TEST relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeGroupsTestsRelationship stores GROUPS_TESTS relationships with group type
func (kdb *KuzuDatabase) storeGroupsTestsRelationship(rel *entities.Relationship) error {
	groupType := ""
	if gt := rel.GetProperty("group_type"); gt != nil {
		groupType = fmt.Sprintf("%v", gt)
	}
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:GROUPS_TESTS {group_type: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, groupType)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store GROUPS_TESTS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeSkipsRelationship stores SKIPS relationships with skip condition
func (kdb *KuzuDatabase) storeSkipsRelationship(rel *entities.Relationship) error {
	skipCondition := ""
	if sc := rel.GetProperty("skip_condition"); sc != nil {
		skipCondition = fmt.Sprintf("%v", sc)
	}
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:SKIPS {skip_condition: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, skipCondition)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store SKIPS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}

// storeDependsRelationship stores DEPENDS relationships with dependency type
func (kdb *KuzuDatabase) storeDependsRelationship(rel *entities.Relationship) error {
	dependencyType := ""
	if dt := rel.GetProperty("dependency_type"); dt != nil {
		dependencyType = fmt.Sprintf("%v", dt)
	}
	query := fmt.Sprintf(`
		MATCH (source:%s {id: "%s"})
		MATCH (target:%s {id: "%s"})
		CREATE (source)-[:DEPENDS {dependency_type: "%s"}]->(target)
	`, rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, dependencyType)
	
	_, err := kdb.Connection.Query(query)
	if err != nil {
		return fmt.Errorf("failed to store DEPENDS relationship from %s:%s to %s:%s: %w", 
			rel.SourceType, rel.SourceID, rel.TargetType, rel.TargetID, err)
	}
	return nil
}
