package git

import (
	"errors"
	"os"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Client struct {
	repo   *git.Repository
	path   string
	url    string
	branch string
}

// NewClient creates a new Git client by cloning a repository or opening an existing one.
// It takes the repository URL, branch name, and local path where the repository should be cloned.
// Returns a configured Client instance or an error if the setup fails.
func NewClient(url, branch, clonePath string) (*Client, error) {
	// 1. Create clone directory
	if err := os.MkdirAll(clonePath, 0755); err != nil {
		return nil, err
	}

	// 2. Attempt to clone repository
	repo, err := git.PlainClone(clonePath, false, &git.CloneOptions{
		URL:           url,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	})

	// 3. Handle existing repository
	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		repo, err = git.PlainOpen(clonePath)
	}

	// 4. Check for errors
	if err != nil {
		return nil, err
	}

	// 5. Create and return new client
	return &Client{repo: repo, path: clonePath, url: url, branch: branch}, nil
}

// Pull fetches and integrates changes from the remote repository.
// Returns the new commit hash, a boolean indicating if changes were detected,
// and any error encountered during the process.
// The boolean is true if the repository was updated, false if it was already up-to-date.
func (c *Client) Pull() (string, bool, error) {
	// 1. Get repository worktree
	wt, err := c.repo.Worktree()
	if err != nil {
		return "", false, err
	}

	// 2. Get current HEAD
	head, err := c.repo.Head()
	if err != nil {
		return "", false, err
	}
	oldCommit := head.Hash().String()

	// 3. Pull from remote
	err = wt.Pull(&git.PullOptions{RemoteName: "origin"})

	// 4. Handle up-to-date case
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return oldCommit, false, nil
	}

	// 5. Handle pull errors
	if err != nil {
		return "", false, err
	}

	// 6. Get new HEAD
	head, err = c.repo.Head()
	if err != nil {
		return "", false, err
	}
	newCommit := head.Hash().String()

	// 7. Return results
	return newCommit, oldCommit != newCommit, nil
}

// Path returns the local filesystem path where the repository is cloned.
// This is a simple getter method that returns the path field of the Client.

func (c *Client) Path() string {
	// 1. Return repository path
	return c.path
}
