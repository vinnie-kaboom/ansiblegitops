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

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	log.Println("Starting ansiblegitops...")
	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Cannot get working directory: %v", err)
	} else {
		log.Printf("Current working directory: %s", pwd)
	}

	// Load configuration
	log.Println("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Configuration loaded successfully: Git URL: %s, Branch: %s, Playbook Dir: %s",
		cfg.Git.URL, cfg.Git.Branch, cfg.Ansible.PlaybookDir)

	tempDir := filepath.Join(os.TempDir(), "ansiblegitops", "repo")
	log.Printf("Using temporary directory: %s", tempDir)

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatalf("Failed to create temporary directory: %v", err)
	}

	log.Println("Initializing Git client...")
	rawGitClient, err := git.NewClient(
		cfg.Git.URL,
		cfg.Git.Branch,
		tempDir,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Git client: %v", err)
	}

	// Wrap the git client to match the interface
	gitClient := &gitClientWrapper{rawGitClient}
	log.Println("Git client initialized successfully")

	log.Println("Setting up Ansible runner...")
	rawAnsibleRunner, err := ansible.NewRunner(
		gitClient.Path(),
		cfg.Ansible.PlaybookDir,
	)
	if err != nil {
		log.Fatalf("Failed to set up Ansible runner: %v", err)
	}
	// Wrap the ansible runner to match the interface
	ansibleRunner := &ansibleRunnerWrapper{rawAnsibleRunner}
	log.Println("Ansible runner configured")

	log.Println("Creating repository reconciler...")
	repoReconciler := reconciler.NewReconciler(gitClient, ansibleRunner)
	interval := time.Duration(cfg.Git.PollInterval) * time.Second
	log.Printf("Starting reconciliation loop with interval: %v", interval)

	repoReconciler.Run(interval)
}

type gitClientWrapper struct {
	*git.Client
}

func (w *gitClientWrapper) Pull() error {
	_, _, err := w.Client.Pull()
	return err
}

type ansibleRunnerWrapper struct {
	*ansible.Runner
}

func (w *ansibleRunnerWrapper) Run() error {
	return w.Runner.Run()
}
