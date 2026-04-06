package ai

import (
	"context"

	"google.golang.org/genai"
)

type Gemini struct {
	client *genai.Client
}

func NewGemini(ctx context.Context, apiKey string) (Llm, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	return &Gemini{client: client}, nil
}

const temperature float32 = 0.8

func (g *Gemini) Ask(ctx context.Context, prompt string) (string, error) {
	t := temperature
	result, err := g.client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash-lite",
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			Temperature: &t,
		},
	)
	if err != nil {
		return "", err
	}
	return result.Text(), nil
}
