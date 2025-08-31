import { tool } from 'ai';
import { z } from 'zod';

// Store pending requests with their resolve functions
const pendingRequests = new Map<string, (result: any) => void>();

// Generate unique request IDs
let requestCounter = 0;
function generateRequestId(): string {
  return `cypher_${Date.now()}_${requestCounter++}`;
}

// Handle incoming cypher_result messages from the Go TUI
export function handleCypherResult(data: any) {
  const { request_id, result, error } = data;
  const resolver = pendingRequests.get(request_id);
  
  if (resolver) {
    pendingRequests.delete(request_id);
    if (error) {
      resolver({ error });
    } else {
      resolver({ result });
    }
  }
}

export function createRunCypherTool(sendMessage: (message: any) => void) {
  return tool({
    description: 'Execute a Cypher query against the code graph database to find relationships between code entities',
    inputSchema: z.object({
      query: z.string().describe('The Cypher query to execute against the graph database'),
    }),
    execute: async ({ query }) => {
      const requestId = generateRequestId();
      
      // Send tool_call message to display in UI
      sendMessage({
        type: 'tool_call',
        data: { toolName: 'run_cypher', args: { query } }
      });
      
      // Send the run_cypher message to the Go TUI
      sendMessage({
        type: 'run_cypher',
        data: { 
          query,
          request_id: requestId
        }
      });
      
      // Log for debugging
      console.error(`[Tool] Running Cypher query: ${query}`);
      
      // Create a promise that will be resolved when we receive the response
      const resultPromise = new Promise<any>((resolve) => {
        pendingRequests.set(requestId, resolve);
        
        // Set a timeout in case we don't get a response
        setTimeout(() => {
          if (pendingRequests.has(requestId)) {
            pendingRequests.delete(requestId);
            resolve({ error: 'Query timeout - no response from graph database' });
          }
        }, 30000); // 30 second timeout
      });
      
      // Wait for the response
      const response = await resultPromise;
      
      if (response.error) {
        return { 
          error: response.error,
          query
        };
      }
      
      return {
        query,
        result: response.result,
        resultCount: response.result ? response.result.split('\n').filter((line: string) => line.trim()).length : 0
      };
    }
  });
}