package ai

import (
	"context"
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client wraps the Anthropic Claude API client with OAuth support
type Client struct {
	tokenGetter func() (string, error) // Gets fresh OAuth access token
}

// NewClient creates a new Claude API client with a token getter function
// This is the recommended approach using OAuth tokens from PocketBase
func NewClient(tokenGetter func() (string, error)) *Client {
	return &Client{
		tokenGetter: tokenGetter,
	}
}

// NewClientWithAPIKey creates a new Claude API client with a direct API key
// This is a simpler alternative to OAuth, useful for development/testing
// For production, prefer NewClient with OAuth tokens (more secure, per-user)
func NewClientWithAPIKey(apiKey string) *Client {
	return &Client{
		tokenGetter: func() (string, error) {
			if apiKey == "" {
				return "", fmt.Errorf("API key is empty")
			}
			return apiKey, nil
		},
	}
}

// Ask sends a prompt to Claude and returns the response
func (c *Client) Ask(ctx context.Context, prompt string, opts ...MessageOption) (string, error) {
	// Get fresh OAuth token
	token, err := c.tokenGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get OAuth token: %w", err)
	}

	// Create HTTP client with OAuth transport
	httpClient := &http.Client{
		Transport: &oauthTransport{token: token},
	}

	// Create Anthropic client
	client := anthropic.NewClient(option.WithHTTPClient(httpClient))

	// Apply options to params
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	for _, opt := range opts {
		opt(&params)
	}

	// Send message
	message, err := client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("Claude API call failed: %w", err)
	}

	// Extract text response
	var response string
	for _, block := range message.Content {
		response += block.Text
	}

	return response, nil
}

// AskStreaming sends a prompt to Claude and streams the response
func (c *Client) AskStreaming(ctx context.Context, prompt string, callback func(chunk string) error, opts ...MessageOption) error {
	// Get fresh OAuth token
	token, err := c.tokenGetter()
	if err != nil {
		return fmt.Errorf("failed to get OAuth token: %w", err)
	}

	// Create HTTP client with OAuth transport
	httpClient := &http.Client{
		Transport: &oauthTransport{token: token},
	}

	// Create Anthropic client
	client := anthropic.NewClient(option.WithHTTPClient(httpClient))

	// Apply options to params
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	for _, opt := range opts {
		opt(&params)
	}

	// Create streaming request
	stream := client.Messages.NewStreaming(ctx, params)

	// Process stream
	for stream.Next() {
		event := stream.Current()

		switch eventVariant := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch deltaVariant := eventVariant.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				if err := callback(deltaVariant.Text); err != nil {
					return fmt.Errorf("callback error: %w", err)
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("streaming error: %w", err)
	}

	return nil
}

// MessageOption allows customizing the Claude API request
type MessageOption func(*anthropic.MessageNewParams)

// WithSystemPrompt sets the system prompt for the request
func WithSystemPrompt(system string) MessageOption {
	return func(params *anthropic.MessageNewParams) {
		params.System = []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: system,
			},
		}
	}
}

// WithModel sets a specific Claude model
func WithModel(model anthropic.Model) MessageOption {
	return func(params *anthropic.MessageNewParams) {
		params.Model = model
	}
}

// WithMaxTokens sets the maximum tokens for the response
func WithMaxTokens(maxTokens int) MessageOption {
	return func(params *anthropic.MessageNewParams) {
		params.MaxTokens = int64(maxTokens)
	}
}

// WithTemperature sets the temperature for response generation
func WithTemperature(temperature float64) MessageOption {
	return func(params *anthropic.MessageNewParams) {
		params.Temperature = anthropic.F(temperature)
	}
}

// oauthTransport is an HTTP transport that injects OAuth Bearer token
type oauthTransport struct {
	token string
}

func (t *oauthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	req = req.Clone(req.Context())

	// Remove API key header if present
	req.Header.Del("x-api-key")

	// Set OAuth Bearer token
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")

	return http.DefaultTransport.RoundTrip(req)
}
