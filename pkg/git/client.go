package git

import (
    "gopkg.in/src-d/go-git.v4"
    "gopkg.in/src-d/go-git.v4/plumbing"
    "os"
    "path/filepath"
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
    repo, err := git.PlainClone(clonePath, false, &git.CloneOptions{
        URL:           url,
        ReferenceName: plumbing.NewBranchReferenceName(branch),
        SingleBranch:  true,
    })
    if err == git.ErrRepositoryAlreadyExists {
        repo, err = git.PlainOpen(clonePath)
    }
    if err != nil {
        return nil, err
    }
    return &Client{repo: repo, path: clonePath, url: url, branch: branch}, nil
}

func (c *Client) Pull() (string, bool, error) {
    wt, err := c.repo.Worktree()
    if err != nil {
        return "", false, err
    }
    head, err := c.repo.Head()
    if err != nil {
        return "", false, err
    }
    oldCommit := head.Hash().String()
    err = wt.Pull(&git.PullOptions{RemoteName: "origin"})
    if err == git.NoErrAlreadyUpToDate {
        return oldCommit, false, nil
    }
    if err != nil {
        return "", false, err
    }
    head, err = c.repo.Head()
    if err != nil {
        return "", false, err
    }
    newCommit := head.Hash().String()
    return newCommit, oldCommit != newCommit, nil
}

func (c *Client) Path() string {
    return c.path
}
