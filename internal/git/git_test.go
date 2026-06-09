package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// newTestRepo creates a throwaway git repo so the helpers run against real git.
func newTestRepo(t *testing.T) Repo {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return Repo{Dir: dir}
}

func stage(t *testing.T, r Repo, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(r.Dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", name)
	cmd.Dir = r.Dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
}

func TestDiffNoChanges(t *testing.T) {
	r := newTestRepo(t)
	if _, err := r.Diff(false); !errors.Is(err, ErrNoChanges) {
		t.Fatalf("expected ErrNoChanges, got %v", err)
	}
}

func TestDiffAndChangedFiles(t *testing.T) {
	r := newTestRepo(t)
	stage(t, r, "hello.txt", "hello world\n")

	diff, err := r.Diff(false)
	if err != nil {
		t.Fatal(err)
	}
	if diff == "" {
		t.Fatal("expected a non-empty diff")
	}

	files, err := r.ChangedFiles(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "hello.txt" {
		t.Fatalf("unexpected files: %v", files)
	}
}

func TestCommit(t *testing.T) {
	r := newTestRepo(t)
	stage(t, r, "a.txt", "content\n")
	if err := r.Commit("feat: add a"); err != nil {
		t.Fatal(err)
	}
	// After committing the staging area is clean again.
	if _, err := r.Diff(false); !errors.Is(err, ErrNoChanges) {
		t.Fatalf("expected clean tree after commit, got %v", err)
	}
}
