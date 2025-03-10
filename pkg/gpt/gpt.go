package gpt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type GPTResponse struct {
	Summary          string   `json:"summary"`
	NewContributions string   `json:"novel_contributions"`
	Tags             []string `json:"tags"`
}

func escapeStringForJSON(input string) (string, error) {
	bytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	escaped := string(bytes)
	if len(escaped) < 2 {
		return "", errors.New("invalid escaped string")
	}
	// Remove the surrounding quotes
	return escaped[1 : len(escaped)-1], nil
}

// GetGPTInfoForPaper calls the GPT API to enrich the paper summary.
func GetGPTInfoForPaper(apiKey, abstract string) (GPTResponse, error) {
	client := openai.NewClient(apiKey)
	friendlyAbstract, err := escapeStringForJSON(abstract)
	if err != nil {
		return GPTResponse{}, err
	}
	ctx := context.Background()
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an AI specialized in scientific research summarization. You will be provided with the abstract of a research paper. Your task is to: 1. Generate a concise summary of the paper (1-2 sentences). 2. Clearly describe the novel contributions or findings the paper introduces. 3. Provide a list of relevant tags (5-10) that are specific and informative. The tags should focus on the paper's key topics and areas, particularly in the context of conditioning of musical generative models or related fields. Avoid general terms like 'AI' or 'machine learning' or any general tag that is obvious. Your response must always be in valid JSON format with the following structure: { \"summary\": \"Your concise summary here.\", \"novel_contributions\": \"Description of what new contributions the paper introduces.\", \"tags\": [\"specific_tag1\", \"specific_tag2\", \"specific_tag3\", \"...\"] } Ensure that all fields are filled accurately and that the response strictly adheres to the above structure.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: friendlyAbstract,
			},
		},
		Temperature: 0.7,
		MaxTokens:   300,
	})
	if err != nil {
		return GPTResponse{}, err
	}
	var gptResponse GPTResponse
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &gptResponse)
	if err != nil {
		return GPTResponse{}, fmt.Errorf("failed to unmarshal GPT response: %w", err)
	}
	return gptResponse, nil
}
