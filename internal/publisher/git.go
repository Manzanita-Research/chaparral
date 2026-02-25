package publisher

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// ErrNoChanges is returned when CommitAndPush finds nothing to commit.
var ErrNoChanges = errors.New("no changes to commit")

// CommitAndPush stages the written files, commits with a structured message,
// and pushes to origin. Returns ErrNoChanges if the working tree is clean.
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

	// Push
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN is not set — needed to push to GitHub")
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &githttp.BasicAuth{
			Username: "token",
			Password: token,
		},
	})
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil
		}
		return fmt.Errorf("pushing: %w", err)
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
