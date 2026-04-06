package ai

import "context"

type Llm interface {
	Ask(ctx context.Context, prompt string) (string, error)
}
