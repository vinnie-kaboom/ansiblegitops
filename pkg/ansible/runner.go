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

const (
	playbookFilename = "site.yml"
	testFilePath     = "/tmp/testfile.txt"
	executionDelay   = 500 * time.Millisecond
)

type Runner struct {
	repoPath      string
	playbookDir   string
	inventoryPath string
	initialized   bool
}

// NewRunner creates and initializes a new Ansible runner with specified repository
// and playbook directory paths. It validates the paths, initializes the runner,
// and returns an error if initialization fails.
func NewRunner(repoPath, playbookDir string, inventoryPath string) (*Runner, error) {
	runner := &Runner{
		repoPath:      strings.TrimSpace(repoPath),
		playbookDir:   strings.TrimSpace(playbookDir),
		inventoryPath: strings.TrimSpace(inventoryPath),
	}

	if err := runner.Init(); err != nil {
		return nil, err
	}

	log.Info().
		Str("repoPath", runner.repoPath).
		Str("playbookDir", runner.playbookDir).
		Str("inventoryPath", runner.inventoryPath).
		Str("fullPath", filepath.Join(runner.repoPath, runner.playbookDir)).
		Msg("Created new Runner with paths")

	return runner, nil
}

// Init verifies that Ansible and ansible-playbook are properly installed
// and available in the system. It performs the initialization only once
// and returns an error if the required tools are not available.
func (r *Runner) Init() error {
	if r.initialized {
		return nil
	}

	if err := verifyAnsibleInstallation(); err != nil {
		return fmt.Errorf("ansible initialization failed: %w", err)
	}

	if err := verifyAnsiblePlaybook(); err != nil {
		return err
	}

	r.initialized = true
	return nil
}

// The Run method is a core function that executes an Ansible playbook
// in four sequential steps:
//  1. Checks if Ansible is properly installed
//  2. Locates and validates the playbook file
//  3. Executes the playbook
//  4. Verifies the execution results
//
// Each step must succeed before moving to the next one, and any error
// immediately stops the process and returns the error.
func (r *Runner) Run() error {
	// 1. Initialize the runner
	if err := r.Init(); err != nil {
		return err
	}

	// 2. Get and verify playbook path
	playbookPath := r.getPlaybookPath()
	if err := r.verifyPlaybookExists(playbookPath); err != nil {
		return err
	}

	// 3. Execute the playbook
	if err := r.executePlaybook(playbookPath); err != nil {
		return err
	}

	// 4. Verify the execution result
	return r.verifyPlaybookResult()
}

// getPlaybookPath constructs and returns the full path to the Ansible
// playbook file by combining the repository path, playbook directory,
// and playbook filename.
func (r *Runner) getPlaybookPath() string {
	// 1. Construct and return full path to playbook
	return filepath.Join(r.repoPath, r.playbookDir, playbookFilename)
}

// verifyPlaybookExists checks if the Ansible playbook file exists at the specified path.
// Returns an error if the file is not found, nil otherwise.
func (r *Runner) verifyPlaybookExists(playbookPath string) error {
	// 1. Check if playbook file exists
	if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
		// 2. Log error and return if file doesn't exist
		log.Error().Str("path", playbookPath).Msg("Playbook file does not exist")
		return fmt.Errorf("playbook file does not exist: %s", playbookPath)
	}
	return nil
}

// executePlaybook executes an Ansible playbook at the specified path.
// It logs the execution process and handles any errors that occur during the playbook run.
// The function creates and executes the ansible-playbook command, captures its output,
// and returns an error if the execution fails.
func (r *Runner) executePlaybook(playbookPath string) error {
	// 1. Log the start of execution
	log.Info().Msg("Starting Ansible playbook execution")

	// 2. Create and execute the command
	cmd := createPlaybookCommand(playbookPath, r.inventoryPath)
	output, err := executeCommand(cmd)

	// 3. Handle errors with detailed logging
	if err != nil {
		log.Error().
			Err(err).
			Str("stdout", output.stdout).
			Str("stderr", output.stderr).
			Msg("Ansible playbook execution failed")
		return fmt.Errorf("ansible-playbook execution failed: %w", err)
	}

	return nil
}

// verifyPlaybookResult checks if the playbook execution was successful by verifying
// the test file creation and its contents. It includes a delay to allow for file
// system operations to complete.
func (r *Runner) verifyPlaybookResult() error {
	// 1. Wait for file system operations to complete
	time.Sleep(executionDelay)

	// 2. Verify test file exists
	if err := verifyFileExists(testFilePath); err != nil {
		return err
	}

	// 3. Read and verify file contents
	content, err := os.ReadFile(testFilePath)
	if err != nil {
		// 4. Handle file reading errors
		log.Error().Err(err).Msg("Error reading testfile.txt")
		return fmt.Errorf("error reading testfile.txt: %w", err)
	}

	// 5. Log file contents for verification
	log.Info().
		Str("content", string(content)).
		Msg("Content of created file")

	return nil
}

// commandOutput represents the output from a command execution,
// containing both stdout and stderr streams.
type commandOutput struct {
	stdout string
	stderr string
}

// verifyAnsibleInstallation checks if Ansible is installed and accessible
// in the system by running the 'ansible --version' command.
// Returns an error if Ansible is not available.
func verifyAnsibleInstallation() error {
	// 1. Execute ansible --version command
	output, err := executeCommand(exec.Command("ansible", "--version"))
	if err != nil {
		// 2. Handle command execution failure with detailed logging
		log.Error().
			Err(err).
			Str("stderr", output.stderr).
			Msg("Ansible is not installed or not accessible")
		return fmt.Errorf("ansible is not installed or not accessible: %w", err)
	}

	// 3. Log version information
	logVersionInfo("Ansible", output.stdout)
	return nil
}

// verifyAnsiblePlaybook verifies that the ansible-playbook command is available
// by executing 'ansible-playbook --version'. Returns an error if the command
// is not accessible.
func verifyAnsiblePlaybook() error {
	// 1. Execute ansible-playbook --version command
	output, err := executeCommand(exec.Command("ansible-playbook", "--version"))
	if err != nil {
		// 2. Handle command execution failure with detailed logging
		log.Error().
			Err(err).
			Str("stderr", output.stderr).
			Msg("ansible-playbook command is not available")
		return fmt.Errorf("ansible-playbook command is not available: %w", err)
	}

	// 3. Log version information
	logVersionInfo("Ansible Playbook", output.stdout)
	return nil
}

// createPlaybookCommand builds an exec.Cmd for running an Ansible playbook
// with the necessary arguments and environment variables.
// Takes the playbook path and inventory file path from configuration.
func createPlaybookCommand(playbookPath, inventoryPath string) *exec.Cmd {
	// 1. Create ansible-playbook command with configurable inventory
	cmd := exec.Command("ansible-playbook",
		"-i", inventoryPath,
		"-v",
		playbookPath)

	// 2. Add environment variables for configuration
	cmd.Env = append(os.Environ(), "ANSIBLE_HOST_KEY_CHECKING=False")
	return cmd
}

// executeCommand runs a given exec.Cmd and captures both stdout and stderr.
// Returns the captured output and any error that occurred during execution.
func executeCommand(cmd *exec.Cmd) (commandOutput, error) {
	// 1. Set up output buffers
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 2. Execute the command
	err := cmd.Run()

	// 3. Return captured output
	return commandOutput{
		stdout: stdout.String(),
		stderr: stderr.String(),
	}, err
}

// verifyFileExists checks if a file exists at the specified path.
// Returns an error if the file doesn't exist or if there are permission issues.
func verifyFileExists(path string) error {
	// 1. Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 2. Handle non-existent file
		log.Warn().Msgf("%s was not created after playbook run", path)
		return fmt.Errorf("expected file %s was not created", path)
	} else if err != nil {
		// 3. Handle other errors
		log.Error().Err(err).Msgf("Error checking for %s", path)
		return fmt.Errorf("error checking for %s: %w", path, err)
	}
	return nil
}

// logVersionInfo extracts and logs version information from command output
// for a specified component.
func logVersionInfo(component, output string) {
	// 1. Extract version from output
	version := strings.Split(output, "\n")[0]

	// 2. Log component version information
	log.Info().
		Str("version", version).
		Msgf("%s installation verified", component)
}
