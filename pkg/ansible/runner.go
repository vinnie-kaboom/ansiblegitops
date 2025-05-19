package ansible

import (
    "os/exec"
    "path/filepath"
    "github.com/rs/zerolog/log"
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
    playbookPath := filepath.Join(r.repoPath, r.playbookDir)
    inventoryPath := filepath.Join(r.repoPath, r.inventoryFile)
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
