package main

import (
	"fmt"
	"log"

	"github.com/onyx/onyx-tui/graph_service/internal/analyzer"
	"github.com/onyx/onyx-tui/graph_service/internal/db"
)

func main() {
	fmt.Println("Testing Go Language Support...")

	// Initialize database
	database, err := db.NewKuzuDatabase("test_go.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create schema
	err = database.CreateSchema()
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// Sample Go code for testing
	sampleGoCode := `package main

import "fmt"

type Person struct {
	Name string
	Age  int
}

func NewPerson(name string, age int) *Person {
	return &Person{Name: name, Age: age}
}

func (p *Person) String() string {
	return fmt.Sprintf("Person{Name: %s, Age: %d}", p.Name, p.Age)
}

func main() {
	person := NewPerson("Alice", 30)
	fmt.Println(person.String())
}
`

	// Test the Go analyzer
	fmt.Println("Testing Go Analyzer...")
	goAnalyzer := analyzer.NewGoAnalyzer()
	file, relationships, err := goAnalyzer.AnalyzeFile("test.go", []byte(sampleGoCode))
	if err != nil {
		log.Fatalf("Failed to analyze Go file: %v", err)
	}

	fmt.Printf("Successfully analyzed Go file: %s\n", file.Path)
	fmt.Printf("Language: %s\n", file.Language)
	fmt.Printf("Entities found: %d\n", len(file.GetAllEntities()))
	fmt.Printf("Relationships found: %d\n", len(relationships))

	// Print found entities
	fmt.Println("\nFound Entities:")
	allEntities := file.GetAllEntities()
	for _, entity := range allEntities {
		fmt.Printf("- %s: %s\n", entity.Type, entity.Name)
	}

	fmt.Println("Go Language Support Test Completed Successfully!")
}
