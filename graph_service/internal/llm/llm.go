package llm

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// LLMClient manages the interaction with the OpenAI API.
type LLMClient struct {
	client *openai.Client
}

// NewLLMClient creates a new OpenAI client, authenticating with an API key
// from the OPENAI_API_KEY environment variable.
func NewLLMClient() (*LLMClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	return &LLMClient{client: openai.NewClient(apiKey)}, nil
}

// GenerateQuery uses the LLM to generate a Cypher query from a user question.
func (c *LLMClient) GenerateQuery(question, schema string) (string, error) {
	prompt := fmt.Sprintf(SystemPrompt, schema, question)

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	return resp.Choices[0].Message.Content, nil
}
