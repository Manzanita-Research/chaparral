package publisher

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// ErrNoChanges is returned when CommitAndPush finds nothing to commit.
var ErrNoChanges = errors.New("no changes to commit")

// CommitAndPush stages the written files, commits with a structured message,
// and pushes to origin using system git. Returns ErrNoChanges if the working
// tree is clean.
func CommitAndPush(brandRepoPath string, writtenFiles []WrittenFile, version string) error {
	repo, err := git.PlainOpen(brandRepoPath)
	if err != nil {
		return fmt.Errorf("opening repo at %s: %w", brandRepoPath, err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("getting worktree: %w", err)
	}

	// Stage each written file
	for _, f := range writtenFiles {
		if _, err := w.Add(f.Path); err != nil {
			return fmt.Errorf("staging %s: %w", f.Path, err)
		}
	}

	// Commit
	msg := fmt.Sprintf("publish marketplace v%s", version)
	_, err = w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "chaparral",
			Email: "chaparral@local",
			When:  time.Now(),
		},
	})
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			return ErrNoChanges
		}
		return fmt.Errorf("committing: %w", err)
	}

	// Push using system git — uses whatever auth the user already has configured
	cmd := exec.Command("git", "push", "origin")
	cmd.Dir = brandRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pushing: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

// RemoteURL returns the origin remote URL for display in confirmation prompts.
func RemoteURL(brandRepoPath string) (string, error) {
	repo, err := git.PlainOpen(brandRepoPath)
	if err != nil {
		return "", fmt.Errorf("opening repo at %s: %w", brandRepoPath, err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return "", fmt.Errorf("no origin remote: %w", err)
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("origin remote has no URLs")
	}

	return urls[0], nil
}
