// Command ai-commit-gen reads your staged git changes and drafts a Conventional
// Commits message using a local (or any OpenAI-compatible) language model.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/gaurav-oberoi/ai-commit-gen/internal/commit"
	"github.com/gaurav-oberoi/ai-commit-gen/internal/git"
	"github.com/gaurav-oberoi/ai-commit-gen/internal/llm"
)

func main() {
	var (
		model    = flag.String("model", "", "model name (defaults to $OPENAI_MODEL or llama3.2)")
		baseURL  = flag.String("base-url", "", "OpenAI-compatible endpoint (defaults to $OPENAI_BASE_URL or local Ollama)")
		all      = flag.Bool("all", false, "describe staged + unstaged changes instead of only staged")
		doCommit = flag.Bool("commit", false, "create the commit with the generated message instead of just printing it")
	)
	flag.Parse()

	if err := run(*model, *baseURL, *all, *doCommit); err != nil {
		fmt.Fprintln(os.Stderr, "ai-commit-gen:", err)
		os.Exit(1)
	}
}

func run(model, baseURL string, all, doCommit bool) error {
	repo := git.Repo{}

	diff, err := repo.Diff(all)
	if errors.Is(err, git.ErrNoChanges) {
		hint := "stage some changes first (git add ...)"
		if all {
			hint = "make some changes first"
		}
		return fmt.Errorf("nothing to describe — %s", hint)
	}
	if err != nil {
		return err
	}

	files, err := repo.ChangedFiles(all)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client := llm.New(baseURL, model)
	message, err := commit.Generate(ctx, client, files, diff)
	if err != nil {
		return err
	}

	if !doCommit {
		fmt.Println(message)
		return nil
	}

	if err := repo.Commit(message); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	fmt.Println("committed:")
	fmt.Println(message)
	return nil
}
