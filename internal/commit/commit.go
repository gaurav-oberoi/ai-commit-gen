// Package commit turns a git diff into a Conventional Commits message by
// prompting a model and cleaning up whatever it sends back.
package commit

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// maxDiffBytes caps how much patch text we forward to the model. Huge diffs add
// little signal and risk blowing the context window, so we trim and tell the
// model the patch was shortened.
const maxDiffBytes = 6000

// Generator depends only on the small surface of an LLM client, which keeps it
// trivial to swap a fake in tests.
type Generator interface {
	Complete(ctx context.Context, system, user string) (string, error)
}

const systemPrompt = `You write git commit messages that follow the Conventional Commits specification.
Rules:
- First line: <type>(<optional scope>): <summary>, max 72 characters, imperative mood, no trailing period.
- type is one of: feat, fix, docs, style, refactor, perf, test, build, ci, chore.
- Optionally add a blank line then a short body explaining the why, wrapped at ~72 characters.
- Output ONLY the commit message. No code fences, no preamble, no quotes.`

// Build assembles the user prompt from the changed files and the (possibly
// trimmed) diff.
func Build(files []string, diff string) string {
	var b strings.Builder
	b.WriteString("Write a commit message for the following change.\n\n")
	if len(files) > 0 {
		b.WriteString("Files changed:\n")
		for _, f := range files {
			fmt.Fprintf(&b, "  - %s\n", f)
		}
		b.WriteString("\n")
	}
	b.WriteString("Diff:\n")
	if len(diff) > maxDiffBytes {
		diff = diff[:maxDiffBytes] + "\n... (diff truncated)"
	}
	b.WriteString(diff)
	return b.String()
}

var fencePattern = regexp.MustCompile("(?s)^```[a-zA-Z]*\\n(.*?)\\n?```$")

// Clean strips the wrappers models like to add (code fences, surrounding
// quotes, stray blank lines) so the result is ready to hand to `git commit`.
func Clean(raw string) string {
	out := strings.TrimSpace(raw)
	if m := fencePattern.FindStringSubmatch(out); m != nil {
		out = strings.TrimSpace(m[1])
	}
	out = strings.Trim(out, "`")
	// Drop a single layer of wrapping quotes around the whole message.
	if len(out) >= 2 && out[0] == '"' && out[len(out)-1] == '"' && !strings.Contains(out[1:len(out)-1], "\"") {
		out = out[1 : len(out)-1]
	}
	return strings.TrimSpace(out)
}

// Generate runs the full pipeline: prompt the model, then clean its reply.
func Generate(ctx context.Context, g Generator, files []string, diff string) (string, error) {
	raw, err := g.Complete(ctx, systemPrompt, Build(files, diff))
	if err != nil {
		return "", err
	}
	msg := Clean(raw)
	if msg == "" {
		return "", fmt.Errorf("model returned an empty message")
	}
	return msg, nil
}
