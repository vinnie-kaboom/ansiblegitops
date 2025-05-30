package git

import (
	"errors"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Client struct {
	repo   *git.Repository
	path   string
	url    string
	branch string
}

func NewClient(url, branch, clonePath string) (*Client, error) {
	if err := os.MkdirAll(clonePath, 0755); err != nil {
		return nil, err
	}

	// Clone options
	cloneOpts := &git.CloneOptions{
		URL:           url,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	}

	repo, err := git.PlainClone(clonePath, false, cloneOpts)
	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		repo, err = git.PlainOpen(clonePath)
	}
	if err != nil {
		return nil, err
	}
	return &Client{repo: repo, path: clonePath, url: url, branch: branch}, nil
}

func (c *Client) Pull() (string, bool, error) {
	// Check if repository exists
	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		// Repository doesn't exist, re-clone it
		cloneOpts := &git.CloneOptions{
			URL:           c.url,
			ReferenceName: plumbing.NewBranchReferenceName(c.branch),
			SingleBranch:  true,
		}

		repo, err := git.PlainClone(c.path, false, cloneOpts)
		if err != nil {
			return "", false, err
		}
		c.repo = repo
		return "", true, nil
	}

	// Repository exists, try to pull with retries
	var lastErr error
	for i := 0; i < 3; i++ { // Try up to 3 times
		wt, err := c.repo.Worktree()
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
			continue
		}

		head, err := c.repo.Head()
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		oldCommit := head.Hash().String()

		pullOpts := &git.PullOptions{
			RemoteName: "origin",
		}

		err = wt.Pull(pullOpts)
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return oldCommit, false, nil
		}
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		head, err = c.repo.Head()
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		newCommit := head.Hash().String()
		return newCommit, oldCommit != newCommit, nil
	}

	return "", false, lastErr
}

func (c *Client) Path() string {
	return c.path
}
