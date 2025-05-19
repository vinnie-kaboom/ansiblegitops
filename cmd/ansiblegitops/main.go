package main

import (
    "github.com/rs/zerolog/log"
    "github.com/yourusername/ansiblegitops/pkg/ansible"
    "github.com/yourusername/ansiblegitops/pkg/config"
    "github.com/yourusername/ansiblegitops/pkg/git"
    "github.com/yourusername/ansiblegitops/pkg/reconciler"
    "time"
)

func main() {
    cfg, err := config.LoadConfig("config.yaml")
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to load config")
    }
    gitClient, err := git.NewClient(
        cfg.Git.URL,
        cfg.Git.Branch,
        "/tmp/ansiblegitops/repo",
    )
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to initialize Git client")
    }
    ansibleRunner := ansible.NewRunner(
        gitClient.Path(),
        cfg.Ansible.PlaybookDir,
        cfg.Ansible.InventoryFile,
    )
    reconciler := reconciler.NewReconciler(gitClient, ansibleRunner)
    interval := time.Duration(cfg.Git.PollInterval) * time.Second
    reconciler.Run(interval)
}
