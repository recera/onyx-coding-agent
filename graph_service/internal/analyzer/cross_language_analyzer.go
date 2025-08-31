package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/onyx/onyx-tui/graph_service/internal/entities"
)

// CrossLanguageAnalyzer analyzes relationships between Python, Go, and TypeScript code
type CrossLanguageAnalyzer struct {
	pythonAnalyzer     *PythonAnalyzer
	goAnalyzer         *EnhancedGoAnalyzer
	typescriptAnalyzer *TypeScriptAnalyzer

	allFiles        map[string]*entities.File
	allEntities     map[string]*entities.Entity
	crossReferences []*entities.Relationship

	httpEndpoints map[string]*EndpointInfo
	apiCalls      map[string]*APICallInfo
}

// EndpointInfo represents HTTP endpoint information
type EndpointInfo struct {
	Method   string
	Path     string
	Language string
	File     string
}

// APICallInfo represents API call information
type APICallInfo struct {
	Method   string
	Target   string
	Language string
	File     string
}

// ProjectAnalysis contains cross-language analysis results
type ProjectAnalysis struct {
	Files         map[string]*entities.File
	Entities      map[string]*entities.Entity
	Relationships []*entities.Relationship
	HTTPEndpoints map[string]*EndpointInfo
	APICalls      map[string]*APICallInfo
	Stats         *CrossLanguageStats
}

// CrossLanguageStats contains analysis statistics
type CrossLanguageStats struct {
	TotalFiles       int
	PythonFiles      int
	GoFiles          int
	TypeScriptFiles  int
	CrossReferences  int
	HTTPEndpoints    int
	APICalls         int
	TestFiles        int
	TestCoverage     float64
}

// NewCrossLanguageAnalyzer creates a new cross-language analyzer
func NewCrossLanguageAnalyzer() *CrossLanguageAnalyzer {
	return &CrossLanguageAnalyzer{
		pythonAnalyzer:     NewPythonAnalyzer(),
		goAnalyzer:         NewEnhancedGoAnalyzer(),
		typescriptAnalyzer: NewTypeScriptAnalyzer(),
		allFiles:           make(map[string]*entities.File),
		allEntities:        make(map[string]*entities.Entity),
		crossReferences:    make([]*entities.Relationship, 0),
		httpEndpoints:      make(map[string]*EndpointInfo),
		apiCalls:           make(map[string]*APICallInfo),
	}
}

// AnalyzeProject analyzes a mixed Python/Go/TypeScript project
func (cla *CrossLanguageAnalyzer) AnalyzeProject(projectPath string) (*ProjectAnalysis, error) {
	fmt.Println("Starting cross-language project analysis...")

	pythonFiles := []string{}
	goFiles := []string{}
	typescriptFiles := []string{}
	testFiles := 0

	// Discover files
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Check if it's a test file
		isTest := strings.Contains(path, ".test.") || strings.Contains(path, ".spec.") ||
			strings.Contains(path, "_test.") || strings.Contains(path, "__tests__/")
		if isTest {
			testFiles++
		}

		ext := filepath.Ext(path)
		switch ext {
		case ".py":
			pythonFiles = append(pythonFiles, path)
		case ".go":
			goFiles = append(goFiles, path)
		case ".ts", ".tsx", ".js", ".jsx":
			typescriptFiles = append(typescriptFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover files: %w", err)
	}

	// Analyze Python files
	for _, filePath := range pythonFiles {
		err := cla.analyzePythonFile(filePath)
		if err != nil {
			fmt.Printf("Warning: Failed to analyze Python file %s: %v\n", filePath, err)
		}
	}

	// Analyze Go files
	for _, filePath := range goFiles {
		err := cla.analyzeGoFile(filePath)
		if err != nil {
			fmt.Printf("Warning: Failed to analyze Go file %s: %v\n", filePath, err)
		}
	}

	// Analyze TypeScript files
	for _, filePath := range typescriptFiles {
		err := cla.analyzeTypeScriptFile(filePath)
		if err != nil {
			fmt.Printf("Warning: Failed to analyze TypeScript file %s: %v\n", filePath, err)
		}
	}

	// Detect cross-language patterns
	cla.detectHTTPEndpoints()
	cla.detectAPICallPatterns()
	cla.buildCrossLanguageRelationships()
	cla.buildTestCoverageRelationships()

	return cla.createProjectAnalysis(len(pythonFiles), len(goFiles), len(typescriptFiles), testFiles), nil
}

// analyzePythonFile analyzes a single Python file
func (cla *CrossLanguageAnalyzer) analyzePythonFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	file, relationships, err := cla.pythonAnalyzer.AnalyzeFile(filePath, content)
	if err != nil {
		return err
	}

	cla.allFiles[filePath] = file

	for _, entity := range file.GetAllEntities() {
		cla.allEntities[entity.ID] = entity
	}

	for _, rel := range relationships {
		cla.crossReferences = append(cla.crossReferences, rel)
	}

	return nil
}

// analyzeGoFile analyzes a single Go file
func (cla *CrossLanguageAnalyzer) analyzeGoFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	file, relationships, err := cla.goAnalyzer.AnalyzeFile(filePath, content)
	if err != nil {
		return err
	}

	cla.allFiles[filePath] = file

	for _, entity := range file.GetAllEntities() {
		cla.allEntities[entity.ID] = entity
	}

	for _, rel := range relationships {
		cla.crossReferences = append(cla.crossReferences, rel)
	}

	return nil
}

// analyzeTypeScriptFile analyzes a single TypeScript/JavaScript file
func (cla *CrossLanguageAnalyzer) analyzeTypeScriptFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	file, relationships, err := cla.typescriptAnalyzer.AnalyzeFile(filePath, content)
	if err != nil {
		return err
	}

	cla.allFiles[filePath] = file

	for _, entity := range file.GetAllEntities() {
		cla.allEntities[entity.ID] = entity
	}

	for _, rel := range relationships {
		cla.crossReferences = append(cla.crossReferences, rel)
	}

	return nil
}

// detectHTTPEndpoints detects HTTP endpoints in all languages
func (cla *CrossLanguageAnalyzer) detectHTTPEndpoints() {
	fmt.Println("Detecting HTTP endpoints...")

	for filePath, file := range cla.allFiles {
		content := string(file.Content)

		switch file.Language {
		case "python":
			cla.detectPythonEndpoints(filePath, content)
		case "go":
			cla.detectGoEndpoints(filePath, content)
		case "typescript", "javascript":
			cla.detectTypeScriptEndpoints(filePath, content)
		}
	}
}

// detectPythonEndpoints detects Python Flask/FastAPI endpoints
func (cla *CrossLanguageAnalyzer) detectPythonEndpoints(filePath, content string) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`@app\.route\s*\(\s*["']([^"']+)["']\s*,?\s*methods\s*=\s*\[["'](\w+)["']\]`),
		regexp.MustCompile(`@app\.(get|post|put|delete)\s*\(\s*["']([^"']+)["']\s*\)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				method := "GET"
				path := match[1]

				if len(match) > 2 && match[2] != "" {
					method = strings.ToUpper(match[2])
				}
				if match[1] != "" && match[2] != "" {
					path = match[2]
					method = strings.ToUpper(match[1])
				}

				key := fmt.Sprintf("%s:%s", method, path)
				cla.httpEndpoints[key] = &EndpointInfo{
					Method:   method,
					Path:     path,
					Language: "python",
					File:     filePath,
				}
			}
		}
	}
}

// detectGoEndpoints detects Go HTTP endpoints
func (cla *CrossLanguageAnalyzer) detectGoEndpoints(filePath, content string) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\.GET\s*\(\s*["']([^"']+)["']\s*,`),
		regexp.MustCompile(`\.POST\s*\(\s*["']([^"']+)["']\s*,`),
		regexp.MustCompile(`\.PUT\s*\(\s*["']([^"']+)["']\s*,`),
		regexp.MustCompile(`\.DELETE\s*\(\s*["']([^"']+)["']\s*,`),
		regexp.MustCompile(`http\.HandleFunc\s*\(\s*["']([^"']+)["']\s*,`),
	}

	methods := []string{"GET", "POST", "PUT", "DELETE", "ANY"}

	for i, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				method := methods[i]
				path := match[1]

				key := fmt.Sprintf("%s:%s", method, path)
				cla.httpEndpoints[key] = &EndpointInfo{
					Method:   method,
					Path:     path,
					Language: "go",
					File:     filePath,
				}
			}
		}
	}
}

// detectTypeScriptEndpoints detects TypeScript/Express endpoints
func (cla *CrossLanguageAnalyzer) detectTypeScriptEndpoints(filePath, content string) {
	// Express.js patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`app\.(get|post|put|delete|patch)\s*\(\s*["'\` + "`" + `]([^"'\` + "`" + `]+)["'\` + "`" + `]`),
		regexp.MustCompile(`router\.(get|post|put|delete|patch)\s*\(\s*["'\` + "`" + `]([^"'\` + "`" + `]+)["'\` + "`" + `]`),
		regexp.MustCompile(`@(Get|Post|Put|Delete|Patch)\s*\(\s*["'\` + "`" + `]([^"'\` + "`" + `]+)["'\` + "`" + `]`), // NestJS decorators
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				method := strings.ToUpper(match[1])
				path := match[2]

				key := fmt.Sprintf("%s:%s", method, path)
				cla.httpEndpoints[key] = &EndpointInfo{
					Method:   method,
					Path:     path,
					Language: "typescript",
					File:     filePath,
				}
			}
		}
	}
}

// detectAPICallPatterns detects API calls in all languages
func (cla *CrossLanguageAnalyzer) detectAPICallPatterns() {
	fmt.Println("Detecting API call patterns...")

	for filePath, file := range cla.allFiles {
		content := string(file.Content)

		patterns := []*regexp.Regexp{
			// Python requests
			regexp.MustCompile(`requests\.(get|post|put|delete)\s*\(\s*["']([^"']+)["']`),
			// Go http client
			regexp.MustCompile(`http\.(Get|Post|Put|Delete)\s*\(\s*["']([^"']+)["']`),
			// TypeScript/JavaScript fetch
			regexp.MustCompile(`fetch\s*\(\s*["'\` + "`" + `]([^"'\` + "`" + `]+)["'\` + "`" + `].*method:\s*["'\` + "`" + `](\w+)["'\` + "`" + `]`),
			// TypeScript/JavaScript axios
			regexp.MustCompile(`axios\.(get|post|put|delete|patch)\s*\(\s*["'\` + "`" + `]([^"'\` + "`" + `]+)["'\` + "`" + `]`),
		}

		for _, pattern := range patterns {
			matches := pattern.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					method := strings.ToUpper(match[1])
					target := match[2]

					key := fmt.Sprintf("%s:%s:%s", file.Language, method, target)
					cla.apiCalls[key] = &APICallInfo{
						Method:   method,
						Target:   target,
						Language: file.Language,
						File:     filePath,
					}
				}
			}
		}
	}
}

// buildCrossLanguageRelationships creates relationships between languages
func (cla *CrossLanguageAnalyzer) buildCrossLanguageRelationships() {
	fmt.Println("Building cross-language relationships...")

	// Match API calls with endpoints
	for callKey, apiCall := range cla.apiCalls {
		for endpointKey, endpoint := range cla.httpEndpoints {
			if cla.pathsMatch(apiCall.Target, endpoint.Path) && apiCall.Language != endpoint.Language {
				// Create cross-language relationship
				relID := fmt.Sprintf("cross_api_%s_%s", callKey, endpointKey)
				relationship := entities.NewRelationshipByID(
					relID,
					entities.RelationshipTypeCalls,
					callKey,     // Using call key as source ID
					endpointKey, // Using endpoint key as target ID
					entities.EntityTypeAPICall, // Source is an API call
					entities.EntityTypeEndpoint, // Target is an API endpoint
				)
				relationship.SetProperty("cross_language", true)
				relationship.SetProperty("api_method", apiCall.Method)
				relationship.SetProperty("api_path", apiCall.Target)
				relationship.SetProperty("source_language", apiCall.Language)
				relationship.SetProperty("target_language", endpoint.Language)

				cla.crossReferences = append(cla.crossReferences, relationship)
			}
		}
	}
}

// pathsMatch checks if API call path matches endpoint path
func (cla *CrossLanguageAnalyzer) pathsMatch(callPath, endpointPath string) bool {
	// Extract path from full URL
	if idx := strings.Index(callPath, "://"); idx != -1 {
		if idx2 := strings.Index(callPath[idx+3:], "/"); idx2 != -1 {
			callPath = callPath[idx+3+idx2:]
		}
	}

	// Remove query parameters
	if idx := strings.Index(callPath, "?"); idx != -1 {
		callPath = callPath[:idx]
	}

	return callPath == endpointPath ||
		strings.HasSuffix(callPath, endpointPath) ||
		strings.Contains(callPath, endpointPath)
}

// buildTestCoverageRelationships builds relationships between tests and code across languages
func (cla *CrossLanguageAnalyzer) buildTestCoverageRelationships() {
	fmt.Println("Building test coverage relationships...")

	// Find all test entities
	testEntities := []*entities.Entity{}
	testedEntities := map[string]bool{}
	
	for _, entity := range cla.allEntities {
		if entity.IsTest() {
			testEntities = append(testEntities, entity)
		}
	}

	// For each test, find what it tests
	for _, test := range testEntities {
		// Check test name for hints about what it tests
		testTarget := test.GetTestTarget()
		if testTarget != "" {
			// Find matching entity
			for _, entity := range cla.allEntities {
				if entity.Name == testTarget && !entity.IsTest() {
					rel := entities.NewRelationshipByID(
						test.ID + "_tests_" + entity.ID,
						entities.RelationshipType("TESTS"),
						test.ID,
						entity.ID,
						entities.EntityTypeTestFunction,
						entity.Type,
					)
					rel.SetProperty("cross_language", true)
					cla.crossReferences = append(cla.crossReferences, rel)
					testedEntities[entity.ID] = true
				}
			}
		}

		// For TypeScript tests, check if they test API endpoints
		if test.FilePath != "" && (strings.HasSuffix(test.FilePath, ".ts") || strings.HasSuffix(test.FilePath, ".tsx")) {
			// Look for API calls in test
			for _, apiCall := range cla.apiCalls {
				// Match API call to endpoint
				for _, endpoint := range cla.httpEndpoints {
					if cla.pathsMatch(apiCall.Target, endpoint.Path) && 
					   strings.EqualFold(apiCall.Method, endpoint.Method) {
						// Create cross-language test relationship
						for _, entity := range cla.allEntities {
							if entity.FilePath == endpoint.File {
								rel := entities.NewRelationshipByID(
									test.ID + "_tests_api_" + entity.ID,
									entities.RelationshipType("TESTS_API"),
									test.ID,
									entity.ID,
									entities.EntityTypeTestFunction,
									entity.Type,
								)
								rel.SetProperty("endpoint", endpoint.Path)
								rel.SetProperty("method", endpoint.Method)
								rel.SetProperty("cross_language", true)
								cla.crossReferences = append(cla.crossReferences, rel)
								testedEntities[entity.ID] = true
							}
						}
					}
				}
			}
		}
	}

	// Calculate test coverage
	totalEntities := 0
	for _, entity := range cla.allEntities {
		if !entity.IsTest() && !entity.IsTestFile() {
			totalEntities++
		}
	}

	coverage := 0.0
	if totalEntities > 0 {
		coverage = float64(len(testedEntities)) / float64(totalEntities) * 100
	}

	fmt.Printf("Test coverage: %.1f%% (%d/%d entities covered)\n", coverage, len(testedEntities), totalEntities)
}

// createProjectAnalysis creates the final analysis result
func (cla *CrossLanguageAnalyzer) createProjectAnalysis(pythonFiles, goFiles, typescriptFiles, testFiles int) *ProjectAnalysis {
	// Calculate test coverage
	totalEntities := 0
	testedEntities := 0
	for _, entity := range cla.allEntities {
		if !entity.IsTest() && !entity.IsTestFile() {
			totalEntities++
			// Check if entity has test relationships
			for _, rel := range cla.crossReferences {
				if (rel.Type == "TESTS" || rel.Type == "TESTS_API" || rel.Type == "COVERS") && 
				   rel.TargetID == entity.ID {
					testedEntities++
					break
				}
			}
		}
	}

	coverage := 0.0
	if totalEntities > 0 {
		coverage = float64(testedEntities) / float64(totalEntities) * 100
	}

	stats := &CrossLanguageStats{
		TotalFiles:      len(cla.allFiles),
		PythonFiles:     pythonFiles,
		GoFiles:         goFiles,
		TypeScriptFiles: typescriptFiles,
		CrossReferences: len(cla.crossReferences),
		HTTPEndpoints:   len(cla.httpEndpoints),
		APICalls:        len(cla.apiCalls),
		TestFiles:       testFiles,
		TestCoverage:    coverage,
	}

	return &ProjectAnalysis{
		Files:         cla.allFiles,
		Entities:      cla.allEntities,
		Relationships: cla.crossReferences,
		HTTPEndpoints: cla.httpEndpoints,
		APICalls:      cla.apiCalls,
		Stats:         stats,
	}
}
