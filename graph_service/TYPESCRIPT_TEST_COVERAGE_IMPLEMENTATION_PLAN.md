# TypeScript Test Coverage Implementation Plan

## Executive Summary

This document outlines a comprehensive plan to implement full test coverage analysis for TypeScript codebases in the Goru graph service. The implementation will bring TypeScript test analysis to parity with existing Go and Python support, enabling complete test coverage metrics, test-to-code relationships, and cross-language test analysis.

## Current State Analysis

### Existing Test Coverage Infrastructure

The codebase already has robust test coverage support for:

1. **Go Language**
   - Detection of `_test.go` files
   - Recognition of Test*, Benchmark*, and Example* functions
   - Support for Go's built-in testing package
   - Assertion counting for testify and standard testing.T methods

2. **Python Language**
   - Detection of `test_*.py`, `*_test.py`, and `tests.py` files
   - Support for pytest, unittest, nose, and doctest frameworks
   - Recognition of test classes inheriting from TestCase
   - Decorator-based test detection (@pytest.mark, @unittest, etc.)

3. **Database Schema**
   - Complete test entity types (TestFunction, TestCase, TestSuite, Assertion, Mock, Fixture)
   - Test relationship types (TESTS, COVERS, MOCKS, ASSERTS, etc.)
   - Coverage metrics and reporting infrastructure

### TypeScript Analysis Gaps

1. **No Test Detection**: TypeScript analyzer lacks any test file or function detection
2. **Missing Framework Support**: No recognition of Jest, Mocha, Vitest, or Jasmine patterns
3. **No Cross-Language Integration**: TypeScript not included in cross-language analyzer
4. **Incomplete API Matching**: TypeScript API calls don't link to backend endpoints

## Implementation Requirements

### Phase 1: Core Test Detection

#### 1.1 Test File Pattern Recognition

```go
// File patterns to detect:
- *.test.ts, *.test.tsx
- *.spec.ts, *.spec.tsx  
- *.test.js, *.test.jsx (for JavaScript)
- *.spec.js, *.spec.jsx
- __tests__/*.ts, __tests__/*.tsx directories
- Files containing describe/it blocks
```

#### 1.2 Test Framework Detection

**Jest Detection:**
- Import patterns: `from '@jest/globals'`, `import jest`
- Global functions: `describe`, `it`, `test`, `expect`
- Setup/teardown: `beforeEach`, `afterEach`, `beforeAll`, `afterAll`
- Mocking: `jest.fn()`, `jest.mock()`, `jest.spyOn()`

**Mocha Detection:**
- Import patterns: `import { describe, it } from 'mocha'`
- Global patterns: `suite`, `test`, `setup`, `teardown`
- Assertion libraries: `chai`, `should`, `expect`

**Vitest Detection:**
- Import patterns: `from 'vitest'`, `import { vi }`
- Similar to Jest but with `vi` instead of `jest`
- Concurrent test support: `describe.concurrent`

**Jasmine Detection:**
- Global functions: `describe`, `it`, `expect`, `spyOn`
- Setup patterns: `beforeEach`, `afterEach`

**Testing Library Detection:**
- React: `@testing-library/react`
- DOM: `@testing-library/dom`
- User events: `@testing-library/user-event`

#### 1.3 Test Entity Extraction

```go
type TypeScriptTestPatterns struct {
    // Test suite patterns
    DescribeBlock    *regexp.Regexp // describe('...', () => {})
    SuiteBlock       *regexp.Regexp // suite('...', () => {})
    
    // Test case patterns  
    ItBlock          *regexp.Regexp // it('...', () => {})
    TestBlock        *regexp.Regexp // test('...', () => {})
    TestEachPattern  *regexp.Regexp // test.each([...])
    
    // Assertion patterns
    ExpectCall       *regexp.Regexp // expect(...).toBe(...)
    AssertCall       *regexp.Regexp // assert.equal(...)
    
    // Mock patterns
    JestMock         *regexp.Regexp // jest.fn(), jest.mock()
    ViMock           *regexp.Regexp // vi.fn(), vi.mock()
    SinonStub        *regexp.Regexp // sinon.stub()
    
    // Hook patterns
    BeforeHooks      *regexp.Regexp // beforeEach, beforeAll
    AfterHooks       *regexp.Regexp // afterEach, afterAll
}
```

### Phase 2: Enhanced Test Analysis

#### 2.1 Test Type Classification

```go
type TestType string

const (
    UnitTest        TestType = "unit"
    IntegrationTest TestType = "integration"
    ComponentTest   TestType = "component"
    E2ETest         TestType = "e2e"
    SnapshotTest    TestType = "snapshot"
    PerformanceTest TestType = "performance"
)
```

#### 2.2 Assertion Detection and Counting

**Jest/Vitest Assertions:**
- `expect(x).toBe(y)`
- `expect(x).toEqual(y)`
- `expect(x).toMatch(pattern)`
- `expect(x).toThrow()`
- `expect(x).resolves.toBe()`
- `expect(x).rejects.toThrow()`
- `expect(component).toMatchSnapshot()`

**Chai Assertions:**
- `expect(x).to.equal(y)`
- `x.should.equal(y)`
- `assert.equal(x, y)`

**Testing Library Assertions:**
- `screen.getByRole()`
- `screen.findByText()`
- `waitFor(() => ...)`

#### 2.3 Mock and Spy Detection

```go
type MockInfo struct {
    Type       string   // "function", "module", "class", "api"
    Target     string   // What is being mocked
    Framework  string   // "jest", "sinon", "vitest"
    Location   Position // Where in code
}
```

#### 2.4 Test Coverage Relationship Building

```go
// Determine what each test is testing
func determineTestTarget(testEntity *Entity) string {
    // 1. Parse test name for hints
    //    "should calculate sum correctly" -> calculateSum
    //    "UserService.getUser returns user" -> UserService.getUser
    
    // 2. Analyze imports in test file
    //    import { calculateSum } from './math'
    
    // 3. Look for direct function calls in test body
    //    const result = calculateSum(2, 3)
    
    // 4. Check mocked modules
    //    jest.mock('./userService')
    
    // 5. Component rendering
    //    render(<Button />)
}
```

### Phase 3: Cross-Language Integration

#### 3.1 TypeScript in Cross-Language Analyzer

```go
// Extend CrossLanguageAnalyzer to include TypeScript
type CrossLanguageAnalyzer struct {
    pythonAnalyzer     *PythonAnalyzer
    goAnalyzer         *EnhancedGoAnalyzer
    typescriptAnalyzer *TypeScriptAnalyzer // NEW
    
    // ... existing fields
}

// Analyze TypeScript files
func (cla *CrossLanguageAnalyzer) analyzeTypeScriptFile(filePath string) error {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return err
    }
    
    file, relationships, err := cla.typescriptAnalyzer.AnalyzeFile(filePath, content)
    // ... process entities and relationships
}
```

#### 3.2 API Call to Endpoint Matching

```go
// Match TypeScript API calls to backend endpoints
func (cla *CrossLanguageAnalyzer) matchTypeScriptAPICallsToEndpoints() {
    for _, apiCall := range cla.typescriptAPICalls {
        endpoint := cla.findMatchingEndpoint(apiCall.URL, apiCall.Method)
        if endpoint != nil {
            cla.createCrossLanguageRelationship(apiCall, endpoint)
        }
    }
}

// Enhanced URL pattern matching
func matchURLPattern(apiPath, endpointPath string) bool {
    // Handle path parameters
    // /api/users/123 matches /api/users/:id
    // /api/products?category=electronics matches /api/products
    
    // Extract base path
    apiBase := extractBasePath(apiPath)
    endpointBase := extractBasePath(endpointPath)
    
    // Convert endpoint pattern to regex
    pattern := convertEndpointToRegex(endpointBase)
    
    return pattern.MatchString(apiBase)
}
```

#### 3.3 Test Coverage Across Languages

```go
// Determine if TypeScript tests cover backend code
func findCrossLanguageTestCoverage(tsTest *Entity) []*Entity {
    covered := []*Entity{}
    
    // 1. Find API calls in test
    apiCalls := extractAPICallsFromTest(tsTest)
    
    // 2. Match to backend endpoints
    for _, call := range apiCalls {
        endpoint := findEndpointHandler(call)
        if endpoint != nil {
            covered = append(covered, endpoint)
            
            // 3. Find what the endpoint calls
            backendCalls := findFunctionCalls(endpoint)
            covered = append(covered, backendCalls...)
        }
    }
    
    return covered
}
```

### Phase 4: Framework-Specific Enhancements

#### 4.1 React Component Testing

```go
type ReactTestInfo struct {
    ComponentTested   string
    RenderMethod      string // "render", "shallow", "mount"
    UserInteractions  []string // "click", "type", "hover"
    PropsProvided     map[string]interface{}
    StateChanges      []string
}

// Detect React Testing Library patterns
func detectReactTestingPatterns(node *ts.Node) *ReactTestInfo {
    // render(<Component prop="value" />)
    // fireEvent.click(button)
    // userEvent.type(input, 'text')
    // screen.getByRole('button')
    // waitFor(() => expect(...))
}
```

#### 4.2 Angular Component Testing

```go
type AngularTestInfo struct {
    ComponentTested  string
    TestBedConfig    map[string]interface{}
    ServiceMocks     []string
    FixtureUsed      bool
}

// Detect Angular testing patterns
func detectAngularTestingPatterns(node *ts.Node) *AngularTestInfo {
    // TestBed.configureTestingModule({...})
    // fixture = TestBed.createComponent(Component)
    // fixture.detectChanges()
    // inject([Service], (service) => {...})
}
```

#### 4.3 Vue Component Testing

```go
type VueTestInfo struct {
    ComponentTested string
    MountMethod     string // "mount", "shallowMount"
    PropsData       map[string]interface{}
    SlotsProvided   []string
}

// Detect Vue Test Utils patterns
func detectVueTestingPatterns(node *ts.Node) *VueTestInfo {
    // mount(Component, { propsData: {...} })
    // wrapper.find('.selector')
    // wrapper.emitted()
    // wrapper.vm.$nextTick()
}
```

#### 4.4 API/Integration Testing

```go
type APITestInfo struct {
    Endpoint      string
    Method        string
    RequestBody   interface{}
    ExpectedCode  int
    Assertions    []string
}

// Detect API testing patterns
func detectAPITestingPatterns(node *ts.Node) *APITestInfo {
    // supertest: request(app).get('/api/users').expect(200)
    // fetch: await fetch('/api/users', { method: 'POST' })
    // axios: await axios.post('/api/users', data)
}
```

## Implementation Plan

### Week 1-2: Core Test Detection

**Day 1-3: Test File Detection**
- [ ] Add `isTypeScriptTestFile()` function
- [ ] Implement test file pattern matching
- [ ] Add configuration for custom test patterns

**Day 4-6: Test Framework Detection**
- [ ] Implement framework detection from imports
- [ ] Add framework-specific pattern recognition
- [ ] Create framework priority system

**Day 7-10: Test Entity Extraction**
- [ ] Implement `enhanceTestFunction()` for TypeScript
- [ ] Add describe/it block parsing
- [ ] Extract test metadata (name, type, framework)

### Week 3-4: Test Analysis Enhancement

**Day 11-13: Assertion Counting**
- [ ] Implement assertion pattern matching
- [ ] Support multiple assertion libraries
- [ ] Add assertion type classification

**Day 14-16: Mock and Spy Detection**
- [ ] Detect jest.mock() patterns
- [ ] Identify spied/stubbed functions
- [ ] Track mock implementations

**Day 17-20: Test-to-Code Relationships**
- [ ] Implement `determineTestTarget()`
- [ ] Build TESTS relationships
- [ ] Add COVERS relationships for indirect testing

### Week 5-6: Cross-Language Integration

**Day 21-23: TypeScript in Cross-Language Analyzer**
- [ ] Add TypeScript analyzer to CrossLanguageAnalyzer
- [ ] Implement TypeScript file processing
- [ ] Update project analysis statistics

**Day 24-26: API Call Matching**
- [ ] Enhance API call detection in TypeScript
- [ ] Implement URL pattern matching
- [ ] Create cross-language relationships

**Day 27-30: Cross-Language Test Coverage**
- [ ] Detect TypeScript tests of backend APIs
- [ ] Build test coverage relationships across languages
- [ ] Calculate cross-language coverage metrics

### Week 7-8: Framework-Specific Features

**Day 31-33: React Testing Support**
- [ ] Detect React Testing Library patterns
- [ ] Extract component test information
- [ ] Handle JSX in tests

**Day 34-36: Angular Testing Support**
- [ ] Detect Angular TestBed patterns
- [ ] Extract service mock information
- [ ] Handle dependency injection in tests

**Day 37-40: Vue & Other Frameworks**
- [ ] Detect Vue Test Utils patterns
- [ ] Add Svelte testing support
- [ ] Support custom testing frameworks

### Week 9-10: Cypher Queries and API

**Day 41-43: Test Coverage Queries**
- [ ] Create TypeScript-specific coverage queries
- [ ] Add framework-based coverage queries
- [ ] Implement test quality metrics

**Day 44-46: API Integration**
- [ ] Update AIAgentAPI for TypeScript tests
- [ ] Add TypeScript test metrics
- [ ] Create test recommendations

**Day 47-50: Testing and Documentation**
- [ ] Create comprehensive test suite
- [ ] Write usage documentation
- [ ] Add example queries

## Technical Implementation Details

### 1. TypeScript Analyzer Modifications

```go
// Add to typescript_analyzer.go

type TypeScriptAnalyzer struct {
    // ... existing fields
    
    // Test-specific fields
    testFramework    string
    testSuites      map[string]*TestSuiteInfo
    testCases       map[string]*TestCaseInfo
    assertions      map[string]*AssertionInfo
    mocks           map[string]*MockInfo
    testCoverage    map[string][]string // test -> covered entities
}

// Main enhancement function
func (ta *TypeScriptAnalyzer) enhanceTestEntities() {
    if !ta.isTestFile() {
        return
    }
    
    ta.detectTestFramework()
    ta.extractTestSuites()
    ta.extractTestCases()
    ta.countAssertions()
    ta.detectMocks()
    ta.buildTestRelationships()
}
```

### 2. Test Pattern Matching

```go
// Comprehensive test patterns
var TypeScriptTestPatterns = struct {
    // File patterns
    TestFilePatterns []string
    
    // Framework patterns
    JestImports     *regexp.Regexp
    MochaImports    *regexp.Regexp
    VitestImports   *regexp.Regexp
    
    // Test structure patterns
    DescribeBlock   *regexp.Regexp
    ItBlock         *regexp.Regexp
    TestBlock       *regexp.Regexp
    
    // Assertion patterns
    ExpectPattern   *regexp.Regexp
    AssertPattern   *regexp.Regexp
    
    // Setup/teardown patterns
    BeforePatterns  *regexp.Regexp
    AfterPatterns   *regexp.Regexp
}{
    TestFilePatterns: []string{
        `.*\.test\.(ts|tsx|js|jsx)$`,
        `.*\.spec\.(ts|tsx|js|jsx)$`,
        `.*/__tests__/.*\.(ts|tsx|js|jsx)$`,
    },
    JestImports: regexp.MustCompile(`from ['"]@?jest`),
    DescribeBlock: regexp.MustCompile(`describe\s*\(\s*['"\` + "`" + `]([^'"\` + "`" + `]+)['"\` + "`" + `]\s*,\s*(?:async\s+)?\(\s*\)\s*=>\s*\{`),
    ExpectPattern: regexp.MustCompile(`expect\s*\([^)]+\)\s*\.(\w+)`),
}
```

### 3. Test Coverage Calculation

```go
// Calculate TypeScript test coverage
func calculateTypeScriptTestCoverage(project *ProjectAnalysis) *CoverageReport {
    report := &CoverageReport{
        Language: "TypeScript",
        Metrics:  make(map[string]interface{}),
    }
    
    // Count test vs production entities
    var testEntities, productionEntities int
    for _, entity := range project.Entities {
        if entity.IsTest() {
            testEntities++
        } else if entity.Language == "TypeScript" {
            productionEntities++
        }
    }
    
    // Calculate coverage
    coveredEntities := findCoveredEntities(project)
    coverage := float64(len(coveredEntities)) / float64(productionEntities) * 100
    
    report.Metrics["total_tests"] = testEntities
    report.Metrics["total_production"] = productionEntities
    report.Metrics["covered_entities"] = len(coveredEntities)
    report.Metrics["coverage_percentage"] = coverage
    
    // Framework breakdown
    frameworkCounts := make(map[string]int)
    for _, entity := range project.Entities {
        if entity.IsTest() {
            framework := entity.GetTestFramework()
            frameworkCounts[framework]++
        }
    }
    report.Metrics["frameworks"] = frameworkCounts
    
    return report
}
```

### 4. Cypher Query Examples

```cypher
-- Find all TypeScript test files
MATCH (f:File)
WHERE f.path =~ '.*\\.test\\.(ts|tsx)$' OR f.path =~ '.*\\.spec\\.(ts|tsx)$'
RETURN f.path, f.language

-- Get test coverage for a TypeScript function
MATCH (func:Function {name: $functionName, language: 'TypeScript'})
OPTIONAL MATCH (test:TestFunction)-[:TESTS]->(func)
RETURN func.name, func.file_path, 
       COUNT(test) as direct_test_count,
       COLLECT(test.name) as test_names

-- Find untested TypeScript components
MATCH (c:Class)
WHERE c.file_path =~ '.*\\.(tsx|jsx)$' 
  AND NOT EXISTS((test:TestFunction)-[:TESTS]->(c))
  AND c.name =~ '.*Component$'
RETURN c.name, c.file_path

-- Cross-language test coverage
MATCH (tsTest:TestFunction)-[:CALLS_API]->(endpoint:Endpoint)
MATCH (endpoint)-[:HANDLED_BY]->(handler:Function)
WHERE tsTest.language = 'TypeScript' 
  AND handler.language IN ['Go', 'Python']
RETURN tsTest.name, endpoint.path, handler.name, handler.language

-- Test framework statistics
MATCH (test:TestFunction)
WHERE test.language = 'TypeScript'
RETURN test.test_framework as framework, 
       COUNT(test) as count,
       AVG(test.assertion_count) as avg_assertions
ORDER BY count DESC

-- Component test coverage
MATCH (component:Class)-[:HAS_PROPS]->(prop:Prop)
WHERE component.file_path =~ '.*\\.tsx$'
OPTIONAL MATCH (test:TestFunction)-[:TESTS]->(component)
RETURN component.name, 
       COUNT(DISTINCT prop) as prop_count,
       COUNT(DISTINCT test) as test_count,
       CASE WHEN COUNT(test) > 0 THEN 'covered' ELSE 'uncovered' END as status
```

## Testing Strategy

### Unit Tests

```go
func TestTypeScriptTestDetection(t *testing.T) {
    tests := []struct {
        name     string
        filePath string
        content  string
        expected bool
    }{
        {
            name:     "Jest test file",
            filePath: "math.test.ts",
            content:  "describe('Math', () => { it('adds', () => { expect(1+1).toBe(2) }) })",
            expected: true,
        },
        {
            name:     "Regular TypeScript file",
            filePath: "math.ts",
            content:  "export function add(a: number, b: number) { return a + b }",
            expected: false,
        },
    }
    
    analyzer := NewTypeScriptAnalyzer()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            file, _, _ := analyzer.AnalyzeFile(tt.filePath, []byte(tt.content))
            hasTests := false
            for _, entity := range file.Entities {
                if entity.IsTest() {
                    hasTests = true
                    break
                }
            }
            assert.Equal(t, tt.expected, hasTests)
        })
    }
}
```

### Integration Tests

```go
func TestTypeScriptCrossLanguageCoverage(t *testing.T) {
    // Create test project structure
    projectDir := createTestProject(t, map[string]string{
        "frontend/api.test.ts": `
            describe('User API', () => {
                it('fetches users', async () => {
                    const response = await fetch('/api/users')
                    expect(response.status).toBe(200)
                })
            })
        `,
        "backend/server.go": `
            func handleUsers(w http.ResponseWriter, r *http.Request) {
                users := getUsersFromDB()
                json.NewEncoder(w).Encode(users)
            }
        `,
    })
    
    // Analyze project
    analyzer := NewCrossLanguageAnalyzer()
    analysis, err := analyzer.AnalyzeProject(projectDir)
    require.NoError(t, err)
    
    // Verify cross-language relationships
    assert.Greater(t, len(analysis.Relationships), 0)
    
    // Find test coverage across languages
    var crossLangCoverage int
    for _, rel := range analysis.Relationships {
        if rel.Source.Language == "TypeScript" && 
           rel.Target.Language == "Go" &&
           rel.Type == "TESTS_API" {
            crossLangCoverage++
        }
    }
    assert.Greater(t, crossLangCoverage, 0)
}
```

## Success Metrics

### Functional Metrics
- ✅ Detection rate: >95% of test files correctly identified
- ✅ Framework accuracy: >90% correct framework detection
- ✅ Assertion counting: ±10% accuracy compared to manual count
- ✅ Test target matching: >80% of tests correctly linked to code
- ✅ Cross-language linking: >75% of API tests linked to endpoints

### Performance Metrics
- ⚡ Analysis speed: <100ms per test file
- ⚡ Memory usage: <50MB for 1000 test files
- ⚡ Query performance: <500ms for coverage calculations

### Quality Metrics
- 📊 Test coverage: >90% of implementation code tested
- 📊 Documentation: 100% of public APIs documented
- 📊 Example coverage: Examples for all major use cases

## Risk Mitigation

### Technical Risks

1. **Pattern Matching Complexity**
   - Risk: Regex patterns may not catch all test variations
   - Mitigation: Use AST parsing as primary method, regex as fallback

2. **Performance Impact**
   - Risk: Test analysis may slow down overall processing
   - Mitigation: Implement caching and incremental analysis

3. **Framework Evolution**
   - Risk: Testing frameworks may change their APIs
   - Mitigation: Design extensible pattern system, version detection

### Implementation Risks

1. **Scope Creep**
   - Risk: Feature requests beyond initial scope
   - Mitigation: Strict phase boundaries, defer enhancements

2. **Integration Complexity**
   - Risk: Cross-language integration more complex than expected
   - Mitigation: Start with simple API matching, enhance iteratively

## Maintenance Plan

### Regular Updates
- Monthly: Review and update test framework patterns
- Quarterly: Add support for new frameworks
- Yearly: Major version compatibility updates

### Monitoring
- Track detection accuracy metrics
- Monitor performance benchmarks
- Collect user feedback on coverage accuracy

### Documentation
- Maintain pattern documentation
- Update example queries
- Create troubleshooting guide

## Conclusion

This implementation plan provides a comprehensive roadmap for adding TypeScript test coverage support to the Goru graph service. The phased approach ensures systematic development while maintaining system stability. With proper execution, this will provide TypeScript developers with the same powerful test coverage analysis currently available for Go and Python, plus enhanced cross-language capabilities unique to modern full-stack applications.

The implementation focuses on:
1. **Completeness**: Supporting all major TypeScript testing frameworks
2. **Accuracy**: Precise test detection and relationship building
3. **Integration**: Seamless cross-language test coverage analysis
4. **Performance**: Efficient processing of large TypeScript codebases
5. **Extensibility**: Easy addition of new frameworks and patterns

Upon completion, users will be able to run Cypher queries to get comprehensive test coverage metrics for their TypeScript repositories, understand test-to-code relationships, and analyze testing patterns across their entire multi-language codebase.