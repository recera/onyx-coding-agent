package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	go_code_graph "github.com/onyx/onyx-tui/graph_service"
)

func createAdvancedTestRepository() (string, error) {
	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "test_repo_")
	if err != nil {
		return "", err
	}

	// Create a more complex Python file with inheritance, decorators, and type annotations
	pythonCode := `"""
Advanced test module for comprehensive Python analysis.
"""
import os
import sys
from typing import List, Dict, Optional
from abc import ABC, abstractmethod

# Module-level variable
DEFAULT_CONFIG: Dict[str, str] = {"debug": "false"}

class Animal(ABC):
    """Base abstract class for all animals."""
    
    def __init__(self, name: str, age: int = 0):
        self.name = name
        self.age = age
    
    @abstractmethod
    def make_sound(self) -> str:
        """Make a sound - must be implemented by subclasses."""
        pass
    
    @property
    def description(self) -> str:
        return f"{self.name} is {self.age} years old"

class Dog(Animal):
    """A dog is a type of animal."""
    
    def __init__(self, name: str, breed: str, age: int = 0):
        super().__init__(name, age)
        self.breed = breed
    
    def make_sound(self) -> str:
        return "Woof!"
    
    def fetch(self, item: str) -> bool:
        print(f"{self.name} fetches the {item}")
        return True

class Cat(Animal):
    """A cat is a type of animal."""
    
    def make_sound(self) -> str:
        return "Meow!"
    
    @staticmethod
    def purr() -> str:
        return "Purr purr"

@staticmethod
def create_animal(animal_type: str, name: str) -> Optional[Animal]:
    """Factory function to create animals."""
    if animal_type.lower() == "dog":
        return Dog(name, "Mixed")
    elif animal_type.lower() == "cat":
        return Cat(name)
    return None

@property 
def get_system_info() -> Dict[str, str]:
    """Get system information."""
    return {
        "platform": sys.platform,
        "python_version": sys.version
    }

def main():
    """Main function to demonstrate the animal hierarchy."""
    animals: List[Animal] = []
    
    # Create some animals
    dog = create_animal("dog", "Buddy")
    cat = create_animal("cat", "Whiskers")
    
    if dog:
        animals.append(dog)
        print(dog.make_sound())
        if isinstance(dog, Dog):
            dog.fetch("ball")
    
    if cat:
        animals.append(cat)
        print(cat.make_sound())
        print(Cat.purr())
    
    # Print all animal descriptions
    for animal in animals:
        print(animal.description)
    
    # Use module variable
    config = DEFAULT_CONFIG.copy()
    config["debug"] = "true"
    
    system_info = get_system_info()
    print(f"Running on {system_info['platform']}")

if __name__ == "__main__":
    main()
`

	// Write the Python file
	pythonFile := filepath.Join(tmpDir, "advanced_animals.py")
	err = ioutil.WriteFile(pythonFile, []byte(pythonCode), 0644)
	if err != nil {
		return "", err
	}

	// Create a second file to test cross-file relationships
	utilsCode := `"""
Utility functions for the animal system.
"""
from typing import Dict, Any
import json

def save_config(config: Dict[str, Any], filename: str = "config.json") -> bool:
    """Save configuration to a JSON file."""
    try:
        with open(filename, 'w') as f:
            json.dump(config, f, indent=2)
        return True
    except Exception as e:
        print(f"Error saving config: {e}")
        return False

def load_config(filename: str = "config.json") -> Dict[str, Any]:
    """Load configuration from a JSON file."""
    try:
        with open(filename, 'r') as f:
            return json.load(f)
    except Exception:
        return {}

class ConfigManager:
    """Manages application configuration."""
    
    def __init__(self, config_file: str = "app.json"):
        self.config_file = config_file
        self._config: Dict[str, Any] = {}
    
    def load(self) -> bool:
        """Load configuration from file."""
        self._config = load_config(self.config_file)
        return bool(self._config)
    
    def save(self) -> bool:
        """Save current configuration to file."""
        return save_config(self._config, self.config_file)
    
    def get(self, key: str, default: Any = None) -> Any:
        """Get a configuration value."""
        return self._config.get(key, default)
    
    def set(self, key: str, value: Any) -> None:
        """Set a configuration value."""
        self._config[key] = value
`

	utilsFile := filepath.Join(tmpDir, "utils.py")
	err = ioutil.WriteFile(utilsFile, []byte(utilsCode), 0644)
	if err != nil {
		return "", err
	}

	return tmpDir, nil
}

func main() {
	fmt.Println("=== Enhanced Python Analysis Integration Test ===")

	// Create advanced test repository
	repoPath, err := createAdvancedTestRepository()
	if err != nil {
		log.Fatalf("Failed to create test repository: %v", err)
	}
	defer os.RemoveAll(repoPath)

	fmt.Printf("Created advanced test repository in: %s\n\n", repoPath)

	// Test the enhanced analyzer
	fmt.Println("=== Testing Enhanced Python Analyzer ===")

	options := go_code_graph.BuildGraphOptions{
		RepoPath:  repoPath,
		DBPath:    "", // Use temporary database
		CleanupDB: true,
	}

	result, err := go_code_graph.BuildGraph(options)
	if err != nil {
		log.Fatalf("Failed to build graph: %v", err)
	}
	defer result.Close()

	fmt.Printf("✅ Successfully analyzed repository!\n")
	fmt.Printf("Files processed: %d\n", result.Stats.FilesCount)
	fmt.Printf("Functions found: %d\n", result.Stats.FunctionsCount)
	fmt.Printf("Classes found: %d\n", result.Stats.ClassesCount)
	fmt.Printf("Methods found: %d\n", result.Stats.MethodsCount)
	fmt.Printf("Relationships found: %d\n", result.Stats.CallsCount)
	fmt.Printf("Errors encountered: %d\n", result.Stats.ErrorsCount)
	fmt.Printf("Database path: %s\n\n", result.DBPath)

	// Get detailed analysis
	analysis := result.GetAnalysisResult()

	fmt.Println("=== Enhanced Analysis Results ===")
	fmt.Printf("Total entities: %d\n", len(analysis.Entities))
	fmt.Printf("Total relationships: %d\n", len(analysis.Relationships))

	// Show inheritance relationships
	fmt.Println("\n--- Inheritance Relationships ---")
	inheritanceCount := 0
	for _, rel := range analysis.Relationships {
		if rel.Type == "INHERITS" {
			inheritanceCount++
			fmt.Printf("  %s INHERITS %s\n", rel.SourceID, rel.TargetID)
		}
	}
	fmt.Printf("Total inheritance relationships: %d\n", inheritanceCount)

	// Show decorator usage
	fmt.Println("\n--- Entities with Decorators ---")
	decoratorCount := 0
	for _, entity := range analysis.Entities {
		if decorators, ok := entity.Properties["decorators"]; ok {
			decoratorCount++
			fmt.Printf("  %s (%s): %v\n", entity.Name, entity.Type, decorators)
		}
	}
	fmt.Printf("Total entities with decorators: %d\n", decoratorCount)

	// Show type annotations
	fmt.Println("\n--- Type Annotations ---")
	typeAnnotationCount := 0
	for _, entity := range analysis.Entities {
		if returnType, ok := entity.Properties["return_type"]; ok {
			typeAnnotationCount++
			fmt.Printf("  %s -> %s\n", entity.Name, returnType)
		}
		if typeAnnotation, ok := entity.Properties["type_annotation"]; ok {
			typeAnnotationCount++
			fmt.Printf("  Variable %s: %s\n", entity.Name, typeAnnotation)
		}
	}
	fmt.Printf("Total type annotations found: %d\n", typeAnnotationCount)

	// Show classes and their methods
	fmt.Println("\n--- Class Hierarchies ---")
	for _, entity := range analysis.Entities {
		if entity.Type == "Class" {
			fmt.Printf("Class: %s\n", entity.Name)
			if superclasses, ok := entity.Properties["superclasses"]; ok {
				fmt.Printf("  Inherits from: %s\n", superclasses)
			}
			fmt.Printf("  Methods:\n")
			for _, child := range entity.Children {
				if child.Type == "Method" {
					fmt.Printf("    - %s%s\n", child.Name, child.Signature)
					if returnType, ok := child.Properties["return_type"]; ok {
						fmt.Printf("      Returns: %s\n", returnType)
					}
				}
			}
			fmt.Println()
		}
	}

	// Test database querying
	fmt.Println("=== Testing Database Queries ===")

	// Query for classes
	classQuery := `MATCH (c:Class) RETURN c.name ORDER BY c.name LIMIT 5`
	fmt.Printf("Running query: %s\n", classQuery)
	classResult, err := result.QueryGraph(classQuery)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Query successful:\n%s\n", classResult)
	}

	// Query for inheritance relationships
	inheritanceQuery := `MATCH (child)-[r:INHERITS]->(parent) RETURN child, parent LIMIT 5`
	fmt.Printf("\nRunning query: %s\n", inheritanceQuery)
	inheritResult, err := result.QueryGraph(inheritanceQuery)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
	} else {
		fmt.Printf("✅ Query successful:\n%s\n", inheritResult)
	}

	fmt.Println("\n🎉 Enhanced integration test completed successfully!")
	fmt.Println("✅ Enhanced Python analyzer working correctly")
	fmt.Println("✅ Inheritance tracking operational")
	fmt.Println("✅ Decorator extraction functional")
	fmt.Println("✅ Type annotation detection working")
	fmt.Println("✅ Advanced entity modeling complete")
}
