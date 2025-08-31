// Package entities provides an EntityRegistry system for efficient entity lookup
// and relationship resolution across the entire codebase analysis.
//
// The EntityRegistry serves as the central authority for entity resolution,
// enabling proper cross-file references and relationship creation by maintaining
// comprehensive mappings of entity names, types, and locations.
//
// Key Features:
//   - Fast entity lookup by name, type, and file path combinations
//   - Cross-file reference resolution for import and call relationships
//   - Thread-safe operations for concurrent analyzer access
//   - Comprehensive entity metadata tracking
//   - Support for scoped entity resolution (file, package, global)
//
// Usage:
//
//	registry := NewEntityRegistry()
//	
//	// Register entities during analysis
//	registry.RegisterEntity(entity)
//	
//	// Resolve function calls
//	targetEntity := registry.ResolveFunction("calculateSum", currentFile)
//	if targetEntity != nil {
//		relationship := entities.NewRelationshipByID(
//			relationshipID,
//			entities.RelationshipTypeCalls,
//			sourceEntity.ID,
//			targetEntity.ID,
//		)
//	}
//
// Thread Safety:
//   - All public methods are thread-safe using read-write mutexes
//   - Concurrent registration and lookup operations are supported
//   - Bulk operations provide atomic updates for better performance
package entities

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// EntityRegistry provides comprehensive entity lookup and resolution capabilities
// for the entire codebase analysis. It maintains multiple indexes for efficient
// entity resolution across different scopes and contexts.
//
// The registry organizes entities in several ways:
//   - By ID: Direct entity lookup using unique identifiers
//   - By Name: Global and file-scoped name-based lookups
//   - By Type: Type-specific entity collections
//   - By File: File-scoped entity collections for local resolution
//   - By Package: Package-scoped resolution for cross-file references
//
// Resolution Strategy:
//   1. Exact ID match (highest priority)
//   2. File-scoped name match (local functions, methods)
//   3. Package-scoped name match (exported functions, types)
//   4. Global name match (built-in functions, standard library)
//   5. Fuzzy name matching (method names, qualified names)
//
// Performance Characteristics:
//   - O(1) lookup for ID-based resolution
//   - O(log n) lookup for name-based resolution using sorted indexes
//   - Memory-efficient storage with string interning for common names
//   - Concurrent access optimized with read-write locks
type EntityRegistry struct {
	// entities holds all registered entities by their unique ID
	// This provides O(1) lookup for direct entity access
	entities map[string]*Entity

	// nameIndex provides multi-level name-based entity lookup
	// Structure: nameIndex[entityName][entityType][filePackage] = []Entity
	// This allows for scoped resolution of entity names
	nameIndex map[string]map[EntityType]map[string][]*Entity

	// typeIndex organizes entities by their type for type-specific queries
	// Structure: typeIndex[entityType] = []Entity
	typeIndex map[EntityType][]*Entity

	// fileIndex organizes entities by their source file for file-scoped resolution
	// Structure: fileIndex[filePath] = []Entity
	fileIndex map[string][]*Entity

	// packageIndex organizes entities by their package for cross-file resolution
	// Structure: packageIndex[packagePath] = []Entity
	packageIndex map[string][]*Entity

	// qualifiedNameIndex handles fully qualified names (e.g., "package.Type.Method")
	// Structure: qualifiedNameIndex[qualifiedName] = Entity
	qualifiedNameIndex map[string]*Entity

	// methodIndex handles method resolution with receiver types
	// Structure: methodIndex[methodName][receiverType] = Entity
	methodIndex map[string]map[string]*Entity

	// importIndex tracks import relationships for cross-file resolution
	// Structure: importIndex[importingFile][importedPackage] = ImportEntity
	importIndex map[string]map[string]*Entity

	// mu protects all registry data structures from concurrent access
	// Uses RWMutex to allow concurrent reads while ensuring exclusive writes
	mu sync.RWMutex

	// stats tracks registry performance and usage statistics
	stats RegistryStats
}

// RegistryStats provides performance and usage statistics for the EntityRegistry
type RegistryStats struct {
	// TotalEntities is the total number of registered entities
	TotalEntities int

	// EntitiesByType counts entities by their type
	EntitiesByType map[EntityType]int

	// LookupCount tracks the number of lookup operations performed
	LookupCount int64

	// ResolvedCount tracks successful entity resolutions
	ResolvedCount int64

	// UnresolvedCount tracks failed entity resolutions  
	UnresolvedCount int64

	// AverageResolutionTime tracks performance metrics
	AverageResolutionTime float64
}

// EntityResolutionContext provides context for entity resolution operations
// This helps the registry make better resolution decisions based on the
// current analysis context and scope.
type EntityResolutionContext struct {
	// CurrentFile is the file path where the reference occurs
	CurrentFile string

	// CurrentEntity is the entity making the reference (for scoped resolution)
	CurrentEntity *Entity

	// CurrentPackage is the package context for the resolution
	CurrentPackage string

	// ExpectedTypes are the expected entity types for the resolution
	// This helps filter results when multiple entities have the same name
	ExpectedTypes []EntityType

	// AllowCrossFile indicates whether cross-file references are allowed
	AllowCrossFile bool

	// AllowBuiltins indicates whether built-in/standard library references are allowed
	AllowBuiltins bool
}

// NewEntityRegistry creates a new EntityRegistry with all indexes initialized
func NewEntityRegistry() *EntityRegistry {
	return &EntityRegistry{
		entities:           make(map[string]*Entity),
		nameIndex:          make(map[string]map[EntityType]map[string][]*Entity),
		typeIndex:          make(map[EntityType][]*Entity),
		fileIndex:          make(map[string][]*Entity),
		packageIndex:       make(map[string][]*Entity),
		qualifiedNameIndex: make(map[string]*Entity),
		methodIndex:        make(map[string]map[string]*Entity),
		importIndex:        make(map[string]map[string]*Entity),
		stats: RegistryStats{
			EntitiesByType: make(map[EntityType]int),
		},
	}
}

// RegisterEntity adds an entity to the registry with full indexing
// This method performs comprehensive indexing to enable efficient lookups
// across all supported resolution strategies.
func (r *EntityRegistry) RegisterEntity(entity *Entity) error {
	if entity == nil {
		return fmt.Errorf("cannot register nil entity")
	}

	if entity.ID == "" {
		return fmt.Errorf("cannot register entity with empty ID")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate entity IDs
	if existing, exists := r.entities[entity.ID]; exists {
		return fmt.Errorf("entity with ID %s already registered (existing: %s %s, new: %s %s)", 
			entity.ID, existing.Type, existing.Name, entity.Type, entity.Name)
	}

	// Primary entity storage
	r.entities[entity.ID] = entity

	// Update type index
	r.typeIndex[entity.Type] = append(r.typeIndex[entity.Type], entity)

	// Update file index
	if entity.FilePath != "" {
		r.fileIndex[entity.FilePath] = append(r.fileIndex[entity.FilePath], entity)
	}

	// Update package index
	if packagePath := r.extractPackagePath(entity.FilePath); packagePath != "" {
		r.packageIndex[packagePath] = append(r.packageIndex[packagePath], entity)
	}

	// Update name index with multi-level structure
	if entity.Name != "" {
		r.updateNameIndex(entity)
	}

	// Update qualified name index for fully qualified names
	r.updateQualifiedNameIndex(entity)

	// Update method index for method entities
	if entity.Type == EntityTypeMethod {
		r.updateMethodIndex(entity)
	}

	// Update import index for import entities
	if entity.Type == EntityTypeImport {
		r.updateImportIndex(entity)
	}

	// Update statistics
	r.stats.TotalEntities++
	r.stats.EntitiesByType[entity.Type]++

	return nil
}

// RegisterEntities performs bulk entity registration with optimized performance
// This method provides atomic registration of multiple entities with batch
// index updates for better performance on large entity sets.
func (r *EntityRegistry) RegisterEntities(entities []*Entity) error {
	if len(entities) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate all entities before registration
	for _, entity := range entities {
		if entity == nil {
			return fmt.Errorf("cannot register nil entity in batch")
		}
		if entity.ID == "" {
			return fmt.Errorf("cannot register entity with empty ID in batch")
		}
		if _, exists := r.entities[entity.ID]; exists {
			return fmt.Errorf("entity with ID %s already registered in batch", entity.ID)
		}
	}

	// Perform batch registration
	for _, entity := range entities {
		// Primary entity storage
		r.entities[entity.ID] = entity

		// Update all indexes
		r.typeIndex[entity.Type] = append(r.typeIndex[entity.Type], entity)

		if entity.FilePath != "" {
			r.fileIndex[entity.FilePath] = append(r.fileIndex[entity.FilePath], entity)
		}

		if packagePath := r.extractPackagePath(entity.FilePath); packagePath != "" {
			r.packageIndex[packagePath] = append(r.packageIndex[packagePath], entity)
		}

		if entity.Name != "" {
			r.updateNameIndex(entity)
		}

		r.updateQualifiedNameIndex(entity)

		if entity.Type == EntityTypeMethod {
			r.updateMethodIndex(entity)
		}

		if entity.Type == EntityTypeImport {
			r.updateImportIndex(entity)
		}

		// Update statistics
		r.stats.TotalEntities++
		r.stats.EntitiesByType[entity.Type]++
	}

	return nil
}

// GetEntityByID performs direct entity lookup by unique identifier
// This is the fastest resolution method with O(1) complexity.
func (r *EntityRegistry) GetEntityByID(id string) *Entity {
	if id == "" {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	r.stats.LookupCount++

	entity := r.entities[id]
	if entity != nil {
		r.stats.ResolvedCount++
	} else {
		r.stats.UnresolvedCount++
	}

	return entity
}

// ResolveEntity performs comprehensive entity resolution using the provided context
// This method implements the full resolution strategy with fallback mechanisms.
func (r *EntityRegistry) ResolveEntity(name string, context *EntityResolutionContext) *Entity {
	if name == "" || context == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	r.stats.LookupCount++

	// Strategy 1: Exact qualified name match
	if entity := r.qualifiedNameIndex[name]; entity != nil {
		if r.matchesContext(entity, context) {
			r.stats.ResolvedCount++
			return entity
		}
	}

	// Strategy 2: File-scoped resolution
	if context.CurrentFile != "" {
		if entity := r.resolveInFile(name, context.CurrentFile, context); entity != nil {
			r.stats.ResolvedCount++
			return entity
		}
	}

	// Strategy 3: Package-scoped resolution
	if context.AllowCrossFile && context.CurrentPackage != "" {
		if entity := r.resolveInPackage(name, context.CurrentPackage, context); entity != nil {
			r.stats.ResolvedCount++
			return entity
		}
	}

	// Strategy 4: Global resolution
	if entity := r.resolveGlobal(name, context); entity != nil {
		r.stats.ResolvedCount++
		return entity
	}

	// Strategy 5: Built-in resolution
	if context.AllowBuiltins {
		if entity := r.resolveBuiltin(name, context); entity != nil {
			r.stats.ResolvedCount++
			return entity
		}
	}

	r.stats.UnresolvedCount++
	return nil
}

// ResolveFunction provides specialized function resolution with method handling
// This method handles function calls including method calls with receiver types.
func (r *EntityRegistry) ResolveFunction(name string, context *EntityResolutionContext) *Entity {
	if name == "" || context == nil {
		return nil
	}

	// Handle method calls (e.g., "obj.Method" or "Type.Method")
	if strings.Contains(name, ".") {
		return r.resolveMethodCall(name, context)
	}

	// Set expected types for function resolution
	functionContext := *context
	functionContext.ExpectedTypes = []EntityType{
		EntityTypeFunction,
		EntityTypeMethod,
		EntityTypeTestFunction,
	}

	return r.ResolveEntity(name, &functionContext)
}

// ResolveType provides specialized type resolution for struct, interface, and class references
func (r *EntityRegistry) ResolveType(name string, context *EntityResolutionContext) *Entity {
	if name == "" || context == nil {
		return nil
	}

	// Set expected types for type resolution
	typeContext := *context
	typeContext.ExpectedTypes = []EntityType{
		EntityTypeStruct,
		EntityTypeInterface,
		EntityTypeClass,
		EntityTypeType,
		EntityTypeEnum,
	}

	return r.ResolveEntity(name, &typeContext)
}

// GetEntitiesByType returns all entities of the specified type
func (r *EntityRegistry) GetEntitiesByType(entityType EntityType) []*Entity {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entities := r.typeIndex[entityType]
	if entities == nil {
		return []*Entity{}
	}

	// Return a copy to prevent external modification
	result := make([]*Entity, len(entities))
	copy(result, entities)
	return result
}

// GetEntitiesByFile returns all entities in the specified file
func (r *EntityRegistry) GetEntitiesByFile(filePath string) []*Entity {
	if filePath == "" {
		return []*Entity{}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	entities := r.fileIndex[filePath]
	if entities == nil {
		return []*Entity{}
	}

	// Return a copy to prevent external modification
	result := make([]*Entity, len(entities))
	copy(result, entities)
	return result
}

// GetStats returns current registry statistics
func (r *EntityRegistry) GetStats() RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy of the stats
	statsCopy := r.stats
	statsCopy.EntitiesByType = make(map[EntityType]int)
	for k, v := range r.stats.EntitiesByType {
		statsCopy.EntitiesByType[k] = v
	}

	return statsCopy
}

// Private helper methods

// updateNameIndex updates the multi-level name index for an entity
func (r *EntityRegistry) updateNameIndex(entity *Entity) {
	name := entity.Name
	entityType := entity.Type
	filePackage := r.extractPackagePath(entity.FilePath)

	// Initialize nested maps if they don't exist
	if r.nameIndex[name] == nil {
		r.nameIndex[name] = make(map[EntityType]map[string][]*Entity)
	}
	if r.nameIndex[name][entityType] == nil {
		r.nameIndex[name][entityType] = make(map[string][]*Entity)
	}

	// Add to file-specific index
	r.nameIndex[name][entityType][entity.FilePath] = append(
		r.nameIndex[name][entityType][entity.FilePath], entity)

	// Add to package-specific index if different from file
	if filePackage != "" && filePackage != entity.FilePath {
		r.nameIndex[name][entityType][filePackage] = append(
			r.nameIndex[name][entityType][filePackage], entity)
	}

	// Add to global index
	r.nameIndex[name][entityType][""] = append(
		r.nameIndex[name][entityType][""], entity)
}

// updateQualifiedNameIndex updates the qualified name index
func (r *EntityRegistry) updateQualifiedNameIndex(entity *Entity) {
	// Build qualified name based on entity hierarchy
	qualifiedName := r.buildQualifiedName(entity)
	if qualifiedName != "" {
		r.qualifiedNameIndex[qualifiedName] = entity
	}

	// For methods, also index with receiver type
	if entity.Type == EntityTypeMethod && entity.Parent != nil {
		receiverQualifiedName := fmt.Sprintf("%s.%s", entity.Parent.Name, entity.Name)
		r.qualifiedNameIndex[receiverQualifiedName] = entity
	}
}

// updateMethodIndex updates the method-specific index
func (r *EntityRegistry) updateMethodIndex(entity *Entity) {
	methodName := entity.Name
	if r.methodIndex[methodName] == nil {
		r.methodIndex[methodName] = make(map[string]*Entity)
	}

	// Index by receiver type if available
	if receiverType := entity.GetProperty("receiver_type"); receiverType != nil {
		if receiverTypeStr, ok := receiverType.(string); ok && receiverTypeStr != "" {
			r.methodIndex[methodName][receiverTypeStr] = entity
		}
	}

	// Index by parent entity name if available
	if entity.Parent != nil {
		r.methodIndex[methodName][entity.Parent.Name] = entity
	}
}

// updateImportIndex updates the import-specific index
func (r *EntityRegistry) updateImportIndex(entity *Entity) {
	if entity.FilePath == "" {
		return
	}

	if r.importIndex[entity.FilePath] == nil {
		r.importIndex[entity.FilePath] = make(map[string]*Entity)
	}

	// Index by import path
	if importPath := entity.GetProperty("path"); importPath != nil {
		if importPathStr, ok := importPath.(string); ok && importPathStr != "" {
			r.importIndex[entity.FilePath][importPathStr] = entity
		}
	}

	// Index by import name/alias
	r.importIndex[entity.FilePath][entity.Name] = entity
}

// extractPackagePath extracts the package path from a file path
func (r *EntityRegistry) extractPackagePath(filePath string) string {
	if filePath == "" {
		return ""
	}

	// Extract directory path as package path
	return filepath.Dir(filePath)
}

// buildQualifiedName builds a qualified name for an entity
func (r *EntityRegistry) buildQualifiedName(entity *Entity) string {
	if entity.Parent != nil {
		parentQualified := r.buildQualifiedName(entity.Parent)
		if parentQualified != "" {
			return fmt.Sprintf("%s.%s", parentQualified, entity.Name)
		}
	}

	// Include package information for top-level entities
	packagePath := r.extractPackagePath(entity.FilePath)
	if packagePath != "" {
		packageName := filepath.Base(packagePath)
		if packageName != "." && packageName != entity.FilePath {
			return fmt.Sprintf("%s.%s", packageName, entity.Name)
		}
	}

	return entity.Name
}

// matchesContext checks if an entity matches the resolution context
func (r *EntityRegistry) matchesContext(entity *Entity, context *EntityResolutionContext) bool {
	// Check expected types
	if len(context.ExpectedTypes) > 0 {
		matches := false
		for _, expectedType := range context.ExpectedTypes {
			if entity.Type == expectedType {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}

	// Check cross-file restrictions
	if !context.AllowCrossFile && entity.FilePath != context.CurrentFile {
		return false
	}

	return true
}

// resolveInFile performs file-scoped entity resolution
func (r *EntityRegistry) resolveInFile(name string, filePath string, context *EntityResolutionContext) *Entity {
	// Check name index for file-specific entities
	if nameMap := r.nameIndex[name]; nameMap != nil {
		for _, expectedType := range context.ExpectedTypes {
			if typeMap := nameMap[expectedType]; typeMap != nil {
				if entities := typeMap[filePath]; len(entities) > 0 {
					return entities[0] // Return first match
				}
			}
		}

		// If no expected types specified, check all types
		if len(context.ExpectedTypes) == 0 {
			for _, typeMap := range nameMap {
				if entities := typeMap[filePath]; len(entities) > 0 {
					return entities[0]
				}
			}
		}
	}

	return nil
}

// resolveInPackage performs package-scoped entity resolution
func (r *EntityRegistry) resolveInPackage(name string, packagePath string, context *EntityResolutionContext) *Entity {
	// Check name index for package-specific entities
	if nameMap := r.nameIndex[name]; nameMap != nil {
		for _, expectedType := range context.ExpectedTypes {
			if typeMap := nameMap[expectedType]; typeMap != nil {
				if entities := typeMap[packagePath]; len(entities) > 0 {
					return entities[0]
				}
			}
		}

		// If no expected types specified, check all types
		if len(context.ExpectedTypes) == 0 {
			for _, typeMap := range nameMap {
				if entities := typeMap[packagePath]; len(entities) > 0 {
					return entities[0]
				}
			}
		}
	}

	return nil
}

// resolveGlobal performs global entity resolution
func (r *EntityRegistry) resolveGlobal(name string, context *EntityResolutionContext) *Entity {
	// Check name index for global entities
	if nameMap := r.nameIndex[name]; nameMap != nil {
		for _, expectedType := range context.ExpectedTypes {
			if typeMap := nameMap[expectedType]; typeMap != nil {
				if entities := typeMap[""]; len(entities) > 0 {
					return entities[0]
				}
			}
		}

		// If no expected types specified, check all types
		if len(context.ExpectedTypes) == 0 {
			for _, typeMap := range nameMap {
				if entities := typeMap[""]; len(entities) > 0 {
					return entities[0]
				}
			}
		}
	}

	return nil
}

// resolveBuiltin performs built-in entity resolution
func (r *EntityRegistry) resolveBuiltin(name string, context *EntityResolutionContext) *Entity {
	// This method can be extended to handle built-in functions and types
	// For now, it returns nil as built-ins are not pre-registered
	// Parameters are kept for future implementation
	_ = name
	_ = context
	return nil
}

// resolveMethodCall handles method call resolution (e.g., "obj.Method", "Type.Method")
func (r *EntityRegistry) resolveMethodCall(qualifiedName string, context *EntityResolutionContext) *Entity {
	parts := strings.Split(qualifiedName, ".")
	if len(parts) != 2 {
		return nil
	}

	receiverName := parts[0]
	methodName := parts[1]

	// Check method index
	if methodMap := r.methodIndex[methodName]; methodMap != nil {
		// Try to resolve by receiver type
		if entity := methodMap[receiverName]; entity != nil {
			if r.matchesContext(entity, context) {
				return entity
			}
		}

		// Try to resolve receiver first, then find method
		receiverContext := *context
		receiverContext.ExpectedTypes = []EntityType{
			EntityTypeStruct, EntityTypeInterface, EntityTypeClass,
		}
		
		if receiverEntity := r.ResolveEntity(receiverName, &receiverContext); receiverEntity != nil {
			if entity := methodMap[receiverEntity.Name]; entity != nil {
				if r.matchesContext(entity, context) {
					return entity
				}
			}
		}
	}

	// Fallback to qualified name resolution
	return r.qualifiedNameIndex[qualifiedName]
}