package ansible

import (
	"github.com/rs/zerolog/log"
	"os/exec"
	"path/filepath"
)

type Runner struct {
	repoPath      string
	playbookDir   string
	inventoryFile string
}

func NewRunner(repoPath, playbookDir, inventoryFile string) *Runner {
	return &Runner{
		repoPath:      repoPath,
		playbookDir:   playbookDir,
		inventoryFile: inventoryFile,
	}
}

func (r *Runner) RunPlaybooks() error {
	inventoryPath := filepath.Join(r.repoPath, r.inventoryFile)

	// First check connectivity using ansible ping
	log.Info().Msg("Verifying host connectivity with ansible ping")
	pingCmd := exec.Command("ansible", "all", "-i", inventoryPath, "-m", "ping")
	output, err := pingCmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Host connectivity check failed")
		return err
	}
	log.Info().Str("output", string(output)).Msg("Host connectivity check successful")

	// Continue with playbook execution
	playbookPath := filepath.Join(r.repoPath, r.playbookDir)
	entries, err := filepath.Glob(filepath.Join(playbookPath, "*.yml"))
	if err != nil {
		return err
	}

	for _, playbook := range entries {
		log.Info().Str("playbook", playbook).Msg("Running Ansible playbook")
		cmd := exec.Command("ansible-playbook", "-i", inventoryPath, playbook)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error().Err(err).Str("output", string(output)).Msg("Playbook execution failed")
			return err
		}
		log.Info().Str("output", string(output)).Msg("Playbook executed successfully")
	}
	return nil
}
