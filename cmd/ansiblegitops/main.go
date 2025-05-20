package main

import (
	"github.com/vinnie-kaboom/ansiblegitops/pkg/ansible"
	"github.com/vinnie-kaboom/ansiblegitops/pkg/config"
	"github.com/vinnie-kaboom/ansiblegitops/pkg/git"
	"github.com/vinnie-kaboom/ansiblegitops/pkg/reconciler"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	tempDir := filepath.Join(os.TempDir(), "ansiblegitops", "repo")

	gitClient, err := git.NewClient(
		cfg.Git.URL,
		cfg.Git.Branch,
		tempDir,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Git client: %v", err)
	}

	ansibleRunner := ansible.NewRunner(
		gitClient.Path(),
		cfg.Ansible.PlaybookDir,
		cfg.Ansible.InventoryFile,
	)
	repoReconciler := reconciler.NewReconciler(gitClient, ansibleRunner)
	interval := time.Duration(cfg.Git.PollInterval) * time.Second
	repoReconciler.Run(interval)
}
