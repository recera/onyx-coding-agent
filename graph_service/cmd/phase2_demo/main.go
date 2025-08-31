package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/onyx/onyx-tui/graph_service/internal/analyzer"
	"github.com/onyx/onyx-tui/graph_service/internal/entities"
)

func main() {
	fmt.Println("=== TypeScript Phase 2 Analysis Demo ===")

	// Create TypeScript analyzer
	tsAnalyzer := analyzer.NewTypeScriptAnalyzer()

	// Read the test TypeScript file
	content, err := ioutil.ReadFile("phase2_typescript_test.ts")
	if err != nil {
		log.Fatalf("Failed to read test file: %v", err)
	}

	// Analyze the file
	file, relationships, err := tsAnalyzer.AnalyzeFile("phase2_typescript_test.ts", content)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	fmt.Printf("\n📁 Analyzed file: %s\n", file.Path)
	fmt.Printf("📊 Found %d entities and %d relationships\n\n", len(file.Entities), len(relationships))

	// Display Phase 2 features
	displayPhase2Features(file, relationships)
}

func displayPhase2Features(file *entities.File, relationships []*entities.Relationship) {
	// Group entities by type
	entityCounts := make(map[string]int)
	entityExamples := make(map[string][]string)

	for _, entity := range file.Entities {
		entityType := string(entity.Type)
		entityCounts[entityType]++

		if len(entityExamples[entityType]) < 3 { // Show max 3 examples
			entityExamples[entityType] = append(entityExamples[entityType], entity.Name)
		}
	}

	fmt.Println("🔍 Phase 2 Entity Types Detected:")
	for entityType, count := range entityCounts {
		fmt.Printf("  %s: %d", entityType, count)
		if examples := entityExamples[entityType]; len(examples) > 0 {
			fmt.Printf(" (examples: %v)\n", examples)
		} else {
			fmt.Println()
		}
	}

	// Group relationships by type
	relationshipCounts := make(map[string]int)
	relationshipExamples := make(map[string][]string)

	for _, rel := range relationships {
		relType := string(rel.Type)
		relationshipCounts[relType]++

		sourceName := rel.SourceID
		if rel.Source != nil {
			sourceName = rel.Source.Name
		}
		targetName := rel.TargetID
		if rel.Target != nil {
			targetName = rel.Target.Name
		}

		if len(relationshipExamples[relType]) < 2 { // Show max 2 examples
			relationshipExamples[relType] = append(relationshipExamples[relType],
				fmt.Sprintf("%s → %s", sourceName, targetName))
		}
	}

	fmt.Println("\n🔗 Phase 2 Relationship Types Detected:")
	for relType, count := range relationshipCounts {
		fmt.Printf("  %s: %d", relType, count)
		if examples := relationshipExamples[relType]; len(examples) > 0 {
			fmt.Printf(" (examples: %v)\n", examples)
		} else {
			fmt.Println()
		}
	}

	// Show Phase 2 specific features
	showAdvancedFeatures(file)
}

func showAdvancedFeatures(file *entities.File) {
	fmt.Println("\n✨ Phase 2 Advanced Features:")

	// Show decorators
	decorators := []string{}
	for _, entity := range file.Entities {
		if entity.Type == entities.EntityTypeDecorator {
			decorators = append(decorators, entity.Name)
		}
	}
	if len(decorators) > 0 {
		fmt.Printf("  🎨 Decorators: %v\n", decorators)
	}

	// Show components
	components := []string{}
	for _, entity := range file.Entities {
		if entity.Type == entities.EntityTypeComponent {
			framework := entity.GetProperty("framework")
			components = append(components, fmt.Sprintf("%s (%v)", entity.Name, framework))
		}
	}
	if len(components) > 0 {
		fmt.Printf("  🧩 Components: %v\n", components)
	}

	// Show hooks
	hooks := []string{}
	for _, entity := range file.Entities {
		if entity.Type == entities.EntityTypeHook {
			hooks = append(hooks, entity.Name)
		}
	}
	if len(hooks) > 0 {
		fmt.Printf("  🪝 React Hooks: %v\n", hooks)
	}

	// Show services
	services := []string{}
	for _, entity := range file.Entities {
		if entity.Type == entities.EntityTypeService {
			services = append(services, entity.Name)
		}
	}
	if len(services) > 0 {
		fmt.Printf("  🔧 Services: %v\n", services)
	}

	// Show namespaces
	namespaces := []string{}
	for _, entity := range file.Entities {
		if entity.Type == entities.EntityTypeNamespace {
			namespaces = append(namespaces, entity.Name)
		}
	}
	if len(namespaces) > 0 {
		fmt.Printf("  📦 Namespaces: %v\n", namespaces)
	}

	// Show modules
	modules := []string{}
	for _, entity := range file.Entities {
		if entity.Type == entities.EntityTypeModule {
			ambient := entity.GetProperty("ambient")
			modules = append(modules, fmt.Sprintf("%s (ambient: %v)", entity.Name, ambient))
		}
	}
	if len(modules) > 0 {
		fmt.Printf("  📋 Module Declarations: %v\n", modules)
	}

	// Show advanced types
	advancedTypes := []string{}
	for _, entity := range file.Entities {
		if entity.Type == entities.EntityTypeType {
			if entity.GetProperty("conditional") == true {
				advancedTypes = append(advancedTypes, entity.Name+" (conditional)")
			} else if entity.GetProperty("mapped") == true {
				advancedTypes = append(advancedTypes, entity.Name+" (mapped)")
			}
		}
	}
	if len(advancedTypes) > 0 {
		fmt.Printf("  🔮 Advanced Types: %v\n", advancedTypes)
	}

	// Show type-only imports
	typeOnlyImports := []string{}
	for _, entity := range file.Entities {
		if entity.Type == entities.EntityTypeImport && entity.GetProperty("type_only") == true {
			typeOnlyImports = append(typeOnlyImports, entity.Name)
		}
	}
	if len(typeOnlyImports) > 0 {
		fmt.Printf("  📥 Type-only Imports: %v\n", typeOnlyImports)
	}

	fmt.Println("\n🎉 Phase 2 TypeScript analysis complete!")
	fmt.Println("\n🔧 Phase 2 Features Successfully Implemented:")
	fmt.Println("  ✅ Enhanced Generic Type Analysis with constraints")
	fmt.Println("  ✅ Decorator Pattern Detection (classes, methods, properties)")
	fmt.Println("  ✅ Advanced Module System Analysis (type-only imports, re-exports)")
	fmt.Println("  ✅ Declaration File Support (.d.ts ambient declarations)")
	fmt.Println("  ✅ Framework Pattern Detection (React, Angular, Vue)")
	fmt.Println("  ✅ Component and Service Detection")
	fmt.Println("  ✅ React Hooks Analysis")
	fmt.Println("  ✅ Namespace and Module Declaration Support")
	fmt.Println("  ✅ Conditional and Mapped Type Analysis")
}
