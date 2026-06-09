package commit

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeLLM struct {
	reply string
	err   error
	gotUs string
}

func (f *fakeLLM) Complete(_ context.Context, _, user string) (string, error) {
	f.gotUs = user
	return f.reply, f.err
}

func TestClean(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "feat: add login", "feat: add login"},
		{"fenced", "```\nfix: handle nil pointer\n```", "fix: handle nil pointer"},
		{"fenced with lang", "```text\nchore: bump deps\n```", "chore: bump deps"},
		{"wrapping quotes", "\"docs: update readme\"", "docs: update readme"},
		{"surrounding whitespace", "  \nrefactor: extract helper\n  ", "refactor: extract helper"},
		{"keeps body", "feat: x\n\nlonger body here", "feat: x\n\nlonger body here"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Clean(tc.in); got != tc.want {
				t.Fatalf("Clean(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBuildTruncatesLargeDiff(t *testing.T) {
	huge := strings.Repeat("a", maxDiffBytes+500)
	out := Build([]string{"big.txt"}, huge)
	if !strings.Contains(out, "diff truncated") {
		t.Fatal("expected large diff to be truncated")
	}
	if !strings.Contains(out, "big.txt") {
		t.Fatal("expected changed file to be listed")
	}
}

func TestGenerateCleansReply(t *testing.T) {
	f := &fakeLLM{reply: "```\nfeat(auth): add token refresh\n```"}
	got, err := Generate(context.Background(), f, []string{"auth.go"}, "some diff")
	if err != nil {
		t.Fatal(err)
	}
	if got != "feat(auth): add token refresh" {
		t.Fatalf("unexpected message: %q", got)
	}
	if !strings.Contains(f.gotUs, "auth.go") {
		t.Fatal("prompt should mention changed files")
	}
}

func TestGenerateRejectsEmpty(t *testing.T) {
	f := &fakeLLM{reply: "   "}
	if _, err := Generate(context.Background(), f, nil, "diff"); err == nil {
		t.Fatal("expected error on empty model reply")
	}
}

func TestGeneratePropagatesError(t *testing.T) {
	f := &fakeLLM{err: errors.New("boom")}
	if _, err := Generate(context.Background(), f, nil, "diff"); err == nil {
		t.Fatal("expected error to propagate")
	}
}
