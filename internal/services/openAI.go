package services

import (
	"bufio"
	"chat-ai-backend/utils"
	"encoding/json"
	"net/http"
	"strings"
)

type OpenAIService struct {
	OpenAIKey string
	OpenAIUrl string
	Client    *http.Client
}

// Constructor
func NewOpenAIService(url string, openaiKey string) *OpenAIService {
	return &OpenAIService{
		OpenAIKey: openaiKey,
		OpenAIUrl: url,
		Client:    &http.Client{Timeout: 0}, // no timeout for streaming
	}
}

// GenerateAIResponse sends a message to the LLM and streams the response
func (s *OpenAIService) GenerateAIResponse(message, conversationID string, responseChan chan<- string) {
	defer close(responseChan)

	payload := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "user", "content": message},
		},
		"stream": true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		utils.Logger.Printf("Failed to marshal payload: %v\n", err)
		responseChan <- "Error encoding request"
		return
	}

	req, err := http.NewRequest("POST", s.OpenAIUrl, strings.NewReader(string(jsonData)))
	if err != nil {
		utils.Logger.Printf("Failed to create request: %v\n", err)
		responseChan <- "Error creating request"
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Add Authorization header if OpenAIKey is set
	if s.OpenAIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.OpenAIKey)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		utils.Logger.Printf("HTTP request failed: %v\n", err)
		responseChan <- "Error connecting to LLM service"
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.Logger.Printf("Non-OK HTTP status: %v\n", resp.Status)
		responseChan <- "Error: LLM service returned an error"
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and [DONE]
		if line == "" || line == "data: [DONE]" {
			continue
		}

		// Trim "data: " prefix unconditionally
		line = strings.TrimPrefix(line, "data: ")

		// Parse JSON and extract content
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			utils.Logger.Printf("Failed to unmarshal chunk: %v\n", err)
			continue
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				responseChan <- content
			}
		}
	}

	if err := scanner.Err(); err != nil {
		utils.Logger.Printf("Error reading response body: %v\n", err)
	}
}
