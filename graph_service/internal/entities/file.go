package entities

import (
	"path/filepath"

	ts "github.com/tree-sitter/go-tree-sitter"
)

// File represents a source code file with its associated entities and metadata
type File struct {
	Path      string             // Full path to the file
	Name      string             // File name (basename)
	Extension string             // File extension
	Language  string             // Programming language
	Tree      *ts.Tree           // Tree-sitter parse tree
	Content   []byte             // File content
	Entities  map[string]*Entity // Map of entity ID to Entity
	Functions []*Entity          // All functions in this file
	Classes   []*Entity          // All classes in this file
	Methods   []*Entity          // All methods in this file
	Imports   []*Entity          // All imports in this file
	Variables []*Entity          // All variables in this file
}

// NewFile creates a new File instance
func NewFile(path string, language string, tree *ts.Tree, content []byte) *File {
	name := filepath.Base(path)
	ext := filepath.Ext(path)

	return &File{
		Path:      path,
		Name:      name,
		Extension: ext,
		Language:  language,
		Tree:      tree,
		Content:   content,
		Entities:  make(map[string]*Entity),
		Functions: make([]*Entity, 0),
		Classes:   make([]*Entity, 0),
		Methods:   make([]*Entity, 0),
		Imports:   make([]*Entity, 0),
		Variables: make([]*Entity, 0),
	}
}

// AddEntity adds an entity to this file and categorizes it
func (f *File) AddEntity(entity *Entity) {
	f.Entities[entity.ID] = entity

	// Categorize the entity
	switch entity.Type {
	case EntityTypeFunction:
		if entity.IsMethod() {
			f.Methods = append(f.Methods, entity)
		} else {
			f.Functions = append(f.Functions, entity)
		}
	case EntityTypeMethod:
		f.Methods = append(f.Methods, entity)
	case EntityTypeClass:
		f.Classes = append(f.Classes, entity)
	case EntityTypeImport:
		f.Imports = append(f.Imports, entity)
	case EntityTypeVariable:
		f.Variables = append(f.Variables, entity)
	}
}

// GetEntity retrieves an entity by ID
func (f *File) GetEntity(id string) *Entity {
	return f.Entities[id]
}

// GetEntitiesByType returns all entities of a specific type
func (f *File) GetEntitiesByType(entityType EntityType) []*Entity {
	var result []*Entity
	for _, entity := range f.Entities {
		if entity.Type == entityType {
			result = append(result, entity)
		}
	}
	return result
}

// GetEntitiesByName returns all entities with a specific name
func (f *File) GetEntitiesByName(name string) []*Entity {
	var result []*Entity
	for _, entity := range f.Entities {
		if entity.Name == name {
			result = append(result, entity)
		}
	}
	return result
}

// GetAllEntities returns all entities in the file
func (f *File) GetAllEntities() []*Entity {
	var result []*Entity
	for _, entity := range f.Entities {
		result = append(result, entity)
	}
	return result
}

// GetStats returns statistics about entities in this file
func (f *File) GetStats() FileStats {
	return FileStats{
		TotalEntities: len(f.Entities),
		Functions:     len(f.Functions),
		Classes:       len(f.Classes),
		Methods:       len(f.Methods),
		Imports:       len(f.Imports),
		Variables:     len(f.Variables),
	}
}

// FileStats contains statistics about a file's entities
type FileStats struct {
	TotalEntities int
	Functions     int
	Classes       int
	Methods       int
	Imports       int
	Variables     int
}
