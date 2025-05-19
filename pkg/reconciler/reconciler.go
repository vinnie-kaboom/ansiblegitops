package reconciler

import (
    "github.com/rs/zerolog/log"
    "github.com/yourusername/ansiblegitops/pkg/ansible"
    "github.com/yourusername/ansiblegitops/pkg/git"
    "time"
)

type Reconciler struct {
    gitClient  *git.Client
    ansible    *ansible.Runner
    lastCommit string
}

func NewReconciler(gitClient *git.Client, ansibleRunner *ansible.Runner) *Reconciler {
    return &Reconciler{
        gitClient: gitClient,
        ansible:   ansibleRunner,
    }
}

func (r *Reconciler) Reconcile() error {
    commit, changed, err := r.gitClient.Pull()
    if err != nil {
        return err
    }
    if !changed && commit == r.lastCommit {
        log.Info().Msg("No changes in repository")
        return nil
    }
    log.Info().Str("commit", commit).Msg("New commit detected, running playbooks")
    if err := r.ansible.RunPlaybooks(); err != nil {
        return err
    }
    r.lastCommit = commit
    log.Info().Str("commit", commit).Msg("Reconciliation completed")
    return nil
}

func (r *Reconciler) Run(interval time.Duration) {
    for {
        if err := r.Reconcile(); err != nil {
            log.Error().Err(err).Msg("Reconciliation failed")
        }
        time.Sleep(interval)
    }
}
