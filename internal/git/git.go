// Package git wraps the handful of git plumbing commands the tool needs to
// read what is about to be committed.
package git

import (
	"errors"
	"os/exec"
	"strings"
)

// ErrNoChanges is returned when there is nothing staged (or nothing at all,
// depending on the mode) to describe.
var ErrNoChanges = errors.New("no changes to commit")

// Repo points at a working tree. An empty Dir means the current directory.
type Repo struct {
	Dir string
}

func (r Repo) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if r.Dir != "" {
		cmd.Dir = r.Dir
	}
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", errors.New(strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

// Diff returns the textual diff to summarise. When includeUnstaged is false it
// looks only at the staging area, which is what you usually want right before a
// commit.
func (r Repo) Diff(includeUnstaged bool) (string, error) {
	args := []string{"diff", "--staged", "--no-color"}
	if includeUnstaged {
		args = []string{"diff", "HEAD", "--no-color"}
	}
	diff, err := r.run(args...)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(diff) == "" {
		return "", ErrNoChanges
	}
	return diff, nil
}

// ChangedFiles lists the paths touched by the diff, used to give the model a
// quick overview before it reads the full patch.
func (r Repo) ChangedFiles(includeUnstaged bool) ([]string, error) {
	args := []string{"diff", "--staged", "--name-only"}
	if includeUnstaged {
		args = []string{"diff", "HEAD", "--name-only"}
	}
	out, err := r.run(args...)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// Commit records the staged changes with the given message.
func (r Repo) Commit(message string) error {
	_, err := r.run("commit", "-m", message)
	return err
}
