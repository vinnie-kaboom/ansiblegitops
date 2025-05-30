package ansible

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Add this function to check Ansible installation
func checkAnsibleInstallation() error {
	cmd := exec.Command("ansible", "--version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Error().
			Err(err).
			Str("stderr", stderr.String()).
			Msg("Ansible is not installed or not accessible")
		return fmt.Errorf("ansible is not installed or not accessible: %w", err)
	}

	version := strings.Split(stdout.String(), "\n")[0]
	log.Info().
		Str("version", version).
		Msg("Ansible installation verified")

	return nil
}

type Runner struct {
	repoPath    string
	playbookDir string
	initialized bool
}

func (r *Runner) Init() error {
	if r.initialized {
		return nil
	}

	// Check Ansible installation
	if err := checkAnsibleInstallation(); err != nil {
		return fmt.Errorf("ansible initialization failed: %w", err)
	}

	// Check ansible-playbook command specifically
	cmd := exec.Command("ansible-playbook", "--version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Error().
			Err(err).
			Str("stderr", stderr.String()).
			Msg("ansible-playbook command is not available")
		return fmt.Errorf("ansible-playbook command is not available: %w", err)
	}

	version := strings.Split(stdout.String(), "\n")[0]
	log.Info().
		Str("version", version).
		Msg("ansible-playbook command verified")

	r.initialized = true
	return nil
}

func (r *Runner) Run() error {
	// Ensure ansible is initialized
	if err := r.Init(); err != nil {
		return err
	}

	log.Info().Msg("Starting Ansible playbook execution")

	playbookPath := filepath.Join(r.repoPath, r.playbookDir, "site.yml")

	// Check if playbook file exists
	if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
		log.Error().Str("path", playbookPath).Msg("Playbook file does not exist")
		return fmt.Errorf("playbook file does not exist: %s", playbookPath)
	}

	// Use explicit localhost connection
	cmd := exec.Command("ansible-playbook",
		"-i", "localhost,", // Use explicit localhost inventory
		"--connection", "local", // Force local connection
		"-v", // Add verbosity
		playbookPath)

	// Set environment variable to avoid host key checking
	cmd.Env = append(os.Environ(), "ANSIBLE_HOST_KEY_CHECKING=False")

	// Create pipes for stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Always log the output regardless of error
	log.Info().
		Str("command", cmd.String()).
		Str("stdout", stdout.String()).
		Str("stderr", stderr.String()).
		Msg("Ansible playbook execution details")

	if err != nil {
		log.Error().
			Err(err).
			Str("stdout", stdout.String()).
			Str("stderr", stderr.String()).
			Msg("Ansible playbook execution failed")
		return fmt.Errorf("ansible-playbook execution failed: %w", err)
	}

	// Add a small delay to ensure file operations are complete
	time.Sleep(500 * time.Millisecond)

	// Check if the file was created
	if _, err := os.Stat("/tmp/testfile.txt"); os.IsNotExist(err) {
		log.Warn().Msg("/tmp/testfile.txt was not created after playbook run")
		return fmt.Errorf("expected file /tmp/testfile.txt was not created")
	} else if err != nil {
		log.Error().Err(err).Msg("Error checking for testfile.txt")
		return fmt.Errorf("error checking for testfile.txt: %w", err)
	}

	// Check file contents
	content, err := os.ReadFile("/tmp/testfile.txt")
	if err != nil {
		log.Error().Err(err).Msg("Error reading testfile.txt")
		return fmt.Errorf("error reading testfile.txt: %w", err)
	}

	log.Info().
		Str("content", string(content)).
		Msg("Content of created file")

	return nil
}

func NewRunner(repoPath, playbookDir string) (*Runner, error) {
	runner := &Runner{
		repoPath:    strings.TrimSpace(repoPath),
		playbookDir: strings.TrimSpace(playbookDir),
	}

	// Initialize right away
	if err := runner.Init(); err != nil {
		return nil, err
	}

	log.Info().
		Str("repoPath", runner.repoPath).
		Str("playbookDir", runner.playbookDir).
		Str("fullPath", filepath.Join(runner.repoPath, runner.playbookDir)).
		Msg("Created new Runner with paths")

	return runner, nil
}

// WrapRunner creates a new ansibleRunnerWrapper
func WrapRunner(runner *Runner) *ansibleRunnerWrapper {
	return &ansibleRunnerWrapper{Runner: runner}
}

type ansibleRunnerWrapper struct {
	*Runner
}

func (w *ansibleRunnerWrapper) Run() error {
	return w.Runner.Run()
}
