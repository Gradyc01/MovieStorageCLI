package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

const model = "gemini-3.5-flash-lite"

// Session holds one natural-language interaction's growing history so
// multi-step exchanges (clarification -> answer -> action) share context.
type Session struct {
	client   *genai.Client
	executor *Executor
	history  []*genai.Content
}

func NewSession(client *genai.Client, executor *Executor) *Session {
	return &Session{client: client, executor: executor}
}

// Handle takes one line of user input, drives the function-calling loop to
// completion (including any clarification round-trips), and returns
// Gemini's final natural-language reply.
func (s *Session) Handle(ctx context.Context, userText string) (string, error) {
	s.history = append(s.history, genai.NewContentFromText(userText, genai.RoleUser))

	for {
		resp, err := s.client.Models.GenerateContent(ctx, model, s.history, &genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(SystemPrompt, genai.RoleUser),
			Tools:             Tools(),
		})
		if err != nil {
			return "", fmt.Errorf("gemini request failed: %w", err)
		}

		calls := resp.FunctionCalls()
		if len(calls) == 0 {
			reply := resp.Text()
			if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
				s.history = append(s.history, resp.Candidates[0].Content)
			} else {
				s.history = append(s.history, genai.NewContentFromText(reply, genai.RoleModel))
			}
			return reply, nil
		}

		// Append the model's own returned content object, not a hand-built
		// one. Thinking-capable Gemini models attach an opaque
		// thought_signature to function-call parts and require it echoed
		// back on the next turn; genai.NewContentFromFunctionCall builds a
		// fresh Content from scratch and silently drops that field, which
		// is what caused "Function call is missing a thought_signature."
		// Handle one call per turn for simplicity and predictability.
		call := calls[0]
		s.history = append(s.history, resp.Candidates[0].Content)

		if call.Name == "ask_clarification" {
			question, _ := call.Args["question"].(string)
			options := toStringSlice(call.Args["options"])

			answer := PromptMultipleChoice(question, options)

			s.history = append(s.history, genai.NewContentFromFunctionResponse(
				call.Name,
				map[string]any{"answer": answer},
				genai.RoleUser,
			))
			continue // let Gemini act on the user's answer next turn
		}

		result, execErr := s.executor.Run(call.Name, call.Args)
		responsePayload := map[string]any{}
		if execErr != nil {
			responsePayload["error"] = execErr.Error()
		} else {
			responsePayload["result"] = result
		}

		s.history = append(s.history, genai.NewContentFromFunctionResponse(
			call.Name,
			responsePayload,
			genai.RoleUser,
		))
		// Loop again so Gemini can summarize the (possibly errored) result.
	}
}

// toStringSlice defensively converts a decoded JSON []any (from the
// function-call args) into []string.
func toStringSlice(v any) []string {
	raw, ok := v.([]any)
	if !ok {
		// Some SDK versions may already hand back []string, or raw JSON.
		if b, err := json.Marshal(v); err == nil {
			var out []string
			if json.Unmarshal(b, &out) == nil {
				return out
			}
		}
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
