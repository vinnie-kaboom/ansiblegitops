package main

import (
	"fmt"
	"github.com/vinnie-kaboom/ansiblegitops/pkg/ansible"
	"github.com/vinnie-kaboom/ansiblegitops/pkg/config"
	"github.com/vinnie-kaboom/ansiblegitops/pkg/git"
	"github.com/vinnie-kaboom/ansiblegitops/pkg/reconciler"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	tempDirName    = "ansiblegitops"
	repoDirName    = "repo"
	dirPermissions = 0755
)

type AppError struct {
	Stage string
	Err   error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %v", e.Stage, e.Err)
}

type App struct {
	cfg            *config.Config
	gitClient      reconciler.GitClient
	ansibleRunner  reconciler.AnsibleRunner
	repoReconciler *reconciler.Reconciler
	tempDir        string
}

func NewApp() *App {
	return &App{}
}

func (a *App) Initialize() error {
	if err := a.setupLogging(); err != nil {
		return &AppError{"logging setup", err}
	}

	if err := a.loadConfig(); err != nil {
		return &AppError{"config loading", err}
	}

	if err := a.setupTempDir(); err != nil {
		return &AppError{"temp directory setup", err}
	}

	if err := a.setupGitClient(); err != nil {
		return &AppError{"git client setup", err}
	}

	if err := a.setupAnsibleRunner(); err != nil {
		return &AppError{"ansible runner setup", err}
	}

	a.setupReconciler()
	return nil
}

func (a *App) setupLogging() error {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Cannot get working directory: %v", err)
	} else {
		log.Printf("Current working directory: %s", pwd)
	}
	return nil
}

func (a *App) loadConfig() error {
	var err error
	a.cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	log.Printf("Configuration loaded successfully: Git URL: %s, Branch: %s, Playbook Dir: %s",
		a.cfg.Git.URL, a.cfg.Git.Branch, a.cfg.Ansible.PlaybookDir)
	return nil
}

func (a *App) setupTempDir() error {
	a.tempDir = filepath.Join(os.TempDir(), tempDirName, repoDirName)
	log.Printf("Using temporary directory: %s", a.tempDir)
	return os.MkdirAll(a.tempDir, dirPermissions)
}

func (a *App) setupGitClient() error {
	rawGitClient, err := git.NewClient(
		a.cfg.Git.URL,
		a.cfg.Git.Branch,
		a.tempDir,
	)
	if err != nil {
		return err
	}
	a.gitClient = &gitClientWrapper{rawGitClient}
	log.Println("Git client initialized successfully")
	return nil
}

func (a *App) setupAnsibleRunner() error {
	rawAnsibleRunner, err := ansible.NewRunner(
		a.gitClient.Path(),
		a.cfg.Ansible.PlaybookDir,
	)
	if err != nil {
		return err
	}
	a.ansibleRunner = &ansibleRunnerWrapper{rawAnsibleRunner}
	log.Println("Ansible runner configured")
	return nil
}

func (a *App) setupReconciler() {
	a.repoReconciler = reconciler.NewReconciler(a.gitClient, a.ansibleRunner)
	log.Println("Repository reconciler created")
}

func (a *App) Run() {
	interval := time.Duration(a.cfg.Git.PollInterval) * time.Second
	log.Printf("Starting reconciliation loop with interval: %v", interval)
	a.repoReconciler.Run(interval)
}

func main() {
	log.Println("Starting ansiblegitops...")

	app := NewApp()
	if err := app.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	app.Run()
}

// Wrapper types remain unchanged
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
