#!/usr/bin/env python3
"""
Python Service for Phase 3 Cross-Language Integration Demo

This service demonstrates how Python and Go services can work together
in a microservices architecture with HTTP APIs and shared data processing.
"""

from flask import Flask, request, jsonify
import asyncio
import aiohttp
import json
import time
from dataclasses import dataclass, asdict
from typing import List, Dict, Any, Optional
import requests

app = Flask(__name__)

@dataclass
class ProcessingJob:
    id: str
    action: str
    data: str
    timestamp: float
    status: str = "pending"
    result: Optional[str] = None

@dataclass
class AnalysisResult:
    job_id: str
    entities_found: int
    relationships_found: int
    language: str
    processing_time: float
    patterns: List[str]

# In-memory storage for demo purposes
processing_jobs: Dict[str, ProcessingJob] = {}
analysis_results: Dict[str, AnalysisResult] = {}

# Configuration for calling Go services
GO_SERVICE_URL = "http://localhost:8080"

@app.route("/health", methods=["GET"])
def health_check():
    """Health check endpoint"""
    return jsonify({"status": "healthy", "service": "python_analyzer"})

@app.route("/api/process", methods=["POST"])
def process_data():
    """
    Main processing endpoint that demonstrates Python analysis capabilities.
    This would be called by the Go service to process data.
    """
    try:
        data = request.get_json()
        
        if not data:
            return jsonify({"error": "No data provided"}), 400
        
        action = data.get("action", "analyze")
        input_data = data.get("data", "")
        
        # Create processing job
        job = ProcessingJob(
            id=f"job_{int(time.time())}",
            action=action,
            data=input_data,
            timestamp=time.time()
        )
        
        processing_jobs[job.id] = job
        
        # Process the data based on action
        if action == "analyze":
            result = analyze_code_structure(input_data)
        elif action == "extract_patterns":
            result = extract_design_patterns(input_data)
        elif action == "generate_report":
            result = generate_analysis_report(input_data)
        else:
            result = {"error": f"Unknown action: {action}"}
        
        # Update job status
        job.status = "completed"
        job.result = json.dumps(result)
        
        return jsonify({
            "job_id": job.id,
            "status": job.status,
            "result": result
        })
        
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route("/api/analyze/<language>", methods=["POST"])
def analyze_by_language(language: str):
    """
    Language-specific analysis endpoint.
    Demonstrates how Python can provide specialized analysis for different languages.
    """
    try:
        data = request.get_json()
        code_content = data.get("code", "")
        
        start_time = time.time()
        
        if language.lower() == "python":
            result = analyze_python_code(code_content)
        elif language.lower() == "go":
            result = analyze_go_code(code_content)
        else:
            return jsonify({"error": f"Unsupported language: {language}"}), 400
        
        processing_time = time.time() - start_time
        
        # Create analysis result
        analysis = AnalysisResult(
            job_id=f"analysis_{int(time.time())}",
            entities_found=result.get("entities_count", 0),
            relationships_found=result.get("relationships_count", 0),
            language=language,
            processing_time=processing_time,
            patterns=result.get("patterns", [])
        )
        
        analysis_results[analysis.job_id] = analysis
        
        return jsonify({
            "analysis_id": analysis.job_id,
            "language": language,
            "processing_time": processing_time,
            "results": asdict(analysis),
            "details": result
        })
        
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route("/api/cross-language/sync", methods=["POST"])
def synchronize_with_go_service():
    """
    Demonstrates calling the Go service from Python.
    This creates bidirectional communication between services.
    """
    try:
        data = request.get_json()
        
        # Call Go service API
        go_response = call_go_service("/api/analyze", data)
        
        # Process the response from Go service
        if go_response:
            # Combine Python and Go analysis results
            python_analysis = analyze_code_structure(data.get("code", ""))
            
            combined_result = {
                "python_analysis": python_analysis,
                "go_analysis": go_response,
                "cross_language_insights": generate_cross_language_insights(
                    python_analysis, go_response
                ),
                "synchronization_status": "success"
            }
            
            return jsonify(combined_result)
        else:
            return jsonify({"error": "Failed to communicate with Go service"}), 500
            
    except Exception as e:
        return jsonify({"error": str(e)}), 500

def analyze_code_structure(code_content: str) -> Dict[str, Any]:
    """
    Analyze code structure and extract entities.
    This simulates Python's code analysis capabilities.
    """
    # Simulate code analysis
    lines = code_content.split('\n')
    
    entities = []
    relationships = []
    patterns = []
    
    # Simple pattern detection
    for i, line in enumerate(lines):
        line = line.strip()
        
        # Detect functions
        if line.startswith('func ') or line.startswith('def '):
            func_name = line.split('(')[0].split()[-1]
            entities.append({
                "type": "function",
                "name": func_name,
                "line": i + 1
            })
        
        # Detect types/classes
        elif line.startswith('type ') or line.startswith('class '):
            type_name = line.split()[1].split('(')[0].split('[')[0]
            entities.append({
                "type": "type",
                "name": type_name,
                "line": i + 1
            })
        
        # Detect channels (Go-specific)
        elif 'chan ' in line or 'make(chan' in line:
            patterns.append("channel_usage")
        
        # Detect goroutines (Go-specific)
        elif line.startswith('go '):
            patterns.append("goroutine_launch")
            relationships.append({
                "type": "launches",
                "source": "current_function",
                "target": line.split()[1].split('(')[0],
                "line": i + 1
            })
    
    # Detect common patterns
    if 'select {' in code_content:
        patterns.append("select_pattern")
    if 'sync.WaitGroup' in code_content:
        patterns.append("wait_group_pattern")
    if 'context.Context' in code_content:
        patterns.append("context_pattern")
    
    return {
        "entities_count": len(entities),
        "relationships_count": len(relationships),
        "entities": entities,
        "relationships": relationships,
        "patterns": list(set(patterns)),
        "analysis_type": "structure_analysis"
    }

def analyze_python_code(code_content: str) -> Dict[str, Any]:
    """Analyze Python-specific code patterns"""
    result = analyze_code_structure(code_content)
    
    # Add Python-specific analysis
    python_patterns = []
    
    if 'async def' in code_content:
        python_patterns.append("async_function")
    if 'await ' in code_content:
        python_patterns.append("await_usage")
    if '@' in code_content:
        python_patterns.append("decorator_usage")
    if 'yield' in code_content:
        python_patterns.append("generator_pattern")
    
    result["python_patterns"] = python_patterns
    result["language"] = "python"
    
    return result

def analyze_go_code(code_content: str) -> Dict[str, Any]:
    """Analyze Go-specific code patterns"""
    result = analyze_code_structure(code_content)
    
    # Add Go-specific analysis
    go_patterns = []
    
    if '[' in code_content and ']' in code_content and 'func' in code_content:
        go_patterns.append("generics_usage")
    if 'interface{' in code_content:
        go_patterns.append("interface_definition")
    if 'defer ' in code_content:
        go_patterns.append("defer_usage")
    if 'recover()' in code_content:
        go_patterns.append("panic_recovery")
    
    result["go_patterns"] = go_patterns
    result["language"] = "go"
    
    return result

def extract_design_patterns(code_content: str) -> Dict[str, Any]:
    """Extract design patterns from code"""
    patterns = []
    
    # Worker pool pattern
    if all(keyword in code_content for keyword in ['chan', 'goroutine', 'worker']):
        patterns.append("worker_pool")
    
    # Pipeline pattern
    if 'chan' in code_content and 'select' in code_content:
        patterns.append("pipeline")
    
    # Producer-consumer pattern
    if 'Producer' in code_content and 'Consumer' in code_content:
        patterns.append("producer_consumer")
    
    # Singleton pattern
    if 'sync.Once' in code_content or ('once' in code_content.lower() and 'do' in code_content.lower()):
        patterns.append("singleton")
    
    return {
        "design_patterns": patterns,
        "pattern_count": len(patterns),
        "analysis_type": "design_patterns"
    }

def generate_analysis_report(code_content: str) -> Dict[str, Any]:
    """Generate comprehensive analysis report"""
    structure_analysis = analyze_code_structure(code_content)
    pattern_analysis = extract_design_patterns(code_content)
    
    return {
        "structure": structure_analysis,
        "patterns": pattern_analysis,
        "summary": {
            "total_entities": structure_analysis["entities_count"],
            "total_relationships": structure_analysis["relationships_count"],
            "design_patterns": len(pattern_analysis["design_patterns"]),
            "complexity_score": calculate_complexity_score(structure_analysis, pattern_analysis)
        },
        "analysis_type": "comprehensive_report"
    }

def calculate_complexity_score(structure: Dict, patterns: Dict) -> float:
    """Calculate code complexity score"""
    base_score = structure["entities_count"] * 0.1
    relationship_score = structure["relationships_count"] * 0.2
    pattern_score = len(patterns["design_patterns"]) * 0.5
    
    return round(base_score + relationship_score + pattern_score, 2)

def generate_cross_language_insights(python_result: Dict, go_result: Dict) -> Dict[str, Any]:
    """Generate insights from cross-language analysis"""
    insights = []
    
    # Compare entity counts
    py_entities = python_result.get("entities_count", 0)
    go_entities = go_result.get("entities_count", 0)
    
    if py_entities > go_entities:
        insights.append("Python service has more complex structure")
    elif go_entities > py_entities:
        insights.append("Go service has more complex structure")
    else:
        insights.append("Services have similar complexity")
    
    # Compare patterns
    py_patterns = set(python_result.get("patterns", []))
    go_patterns = set(go_result.get("patterns", []))
    
    common_patterns = py_patterns.intersection(go_patterns)
    if common_patterns:
        insights.append(f"Common patterns: {', '.join(common_patterns)}")
    
    unique_py = py_patterns - go_patterns
    if unique_py:
        insights.append(f"Python-specific patterns: {', '.join(unique_py)}")
    
    unique_go = go_patterns - py_patterns
    if unique_go:
        insights.append(f"Go-specific patterns: {', '.join(unique_go)}")
    
    return {
        "insights": insights,
        "common_patterns": list(common_patterns),
        "python_unique": list(unique_py),
        "go_unique": list(unique_go),
        "similarity_score": len(common_patterns) / max(len(py_patterns) + len(go_patterns) - len(common_patterns), 1)
    }

def call_go_service(endpoint: str, data: Dict) -> Optional[Dict]:
    """Call Go service API"""
    try:
        url = f"{GO_SERVICE_URL}{endpoint}"
        response = requests.post(url, json=data, timeout=10)
        
        if response.status_code == 200:
            return response.json()
        else:
            print(f"Go service call failed: {response.status_code}")
            return None
            
    except requests.exceptions.RequestException as e:
        print(f"Failed to call Go service: {e}")
        return None

@app.route("/api/jobs/<job_id>", methods=["GET"])
def get_job_status(job_id: str):
    """Get processing job status"""
    if job_id in processing_jobs:
        job = processing_jobs[job_id]
        return jsonify(asdict(job))
    else:
        return jsonify({"error": "Job not found"}), 404

@app.route("/api/results/<analysis_id>", methods=["GET"])
def get_analysis_results(analysis_id: str):
    """Get analysis results"""
    if analysis_id in analysis_results:
        result = analysis_results[analysis_id]
        return jsonify(asdict(result))
    else:
        return jsonify({"error": "Analysis not found"}), 404

if __name__ == "__main__":
    print("Starting Python Analysis Service...")
    print("This service demonstrates cross-language integration with Go")
    print("Available endpoints:")
    print("  POST /api/process - Process data")
    print("  POST /api/analyze/<language> - Language-specific analysis")
    print("  POST /api/cross-language/sync - Synchronize with Go service")
    print("  GET  /health - Health check")
    
    app.run(host="0.0.0.0", port=8000, debug=True) 