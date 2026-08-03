package renderers

import (
	"strings"
	"testing"

	"github.com/ollama/ollama/api"
)

// A thinking-token budget is enforced by the runner's sampler, which only
// engages when it sees the opening tag in generated output. Prefilling that tag
// after a tool response therefore disables the budget for every turn of an
// agent loop, so the prefill has to go when a budget is in force.
func TestGemma4ToolResponsePrefillRespectsBudget(t *testing.T) {
	msgs := []api.Message{
		{Role: "user", Content: "list the directory"},
		{Role: "assistant", ToolCalls: []api.ToolCall{{Function: api.ToolCallFunction{Name: "list_dir"}}}},
		{Role: "tool", ToolName: "list_dir", Content: "a.txt"},
	}

	const openTag = "<|channel>thought\n"

	got, err := RenderWithRenderer("gemma4", msgs, nil, &api.ThinkValue{Value: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(got, openTag) {
		t.Errorf("without a budget the thinking tag should still be prefilled, tail = %q", promptTail(got))
	}

	for _, budget := range []any{2048, "medium"} {
		got, err := RenderWithRenderer("gemma4", msgs, nil, &api.ThinkValue{Value: budget})
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(got, openTag) {
			t.Errorf("think=%v: tag prefilled, the budget would never engage; tail = %q", budget, promptTail(got))
		}
	}
}

func promptTail(s string) string {
	if len(s) > 48 {
		return s[len(s)-48:]
	}
	return s
}
