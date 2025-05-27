package reconciler

import (
	"log"
	"os"
	"time"
)

type GitClient interface {
	Pull() error
	Path() string
}

type AnsibleRunner interface {
	Run() error
}

type Reconciler struct {
	git     GitClient
	ansible AnsibleRunner
}

// NewReconciler creates a new Reconciler instance that manages the synchronization
// between a Git repository and Ansible playbook execution.
// It takes a GitClient for repository operations and an AnsibleRunner for playbook execution.
func NewReconciler(git GitClient, ansible AnsibleRunner) *Reconciler {
	// 1. Configure detailed logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	// 2. Create and return new reconciler instance
	return &Reconciler{
		git:     git,
		ansible: ansible,
	}
}

// Run starts the continuous reconciliation loop that monitors the Git repository
// for changes and executes Ansible playbooks when updates are detected.
// The interval parameter determines how frequently the repository is checked for changes.
func (r *Reconciler) Run(interval time.Duration) {
	// 1. Initialize logging and display startup information
	log.Printf("Starting reconciliation loop with interval: %v", interval)
	log.Printf("Git repository path: %s", r.git.Path())

	// 2. Verify temporary directory permissions
	if err := os.MkdirAll("/tmp", 0755); err != nil {
		log.Printf("Warning: Failed to verify /tmp directory permissions: %v", err)
	}

	// 3. Start reconciliation loop
	for {
		// 4. Pull latest changes from repository
		if err := r.git.Pull(); err != nil {
			log.Printf("Error pulling repository: %v", err)
			time.Sleep(interval)
			continue
		}

		// 5. Execute Ansible playbook
		if err := r.ansible.Run(); err != nil {
			log.Printf("Error running Ansible playbook: %v", err)
			time.Sleep(interval)
			continue
		}

		// 6. Verify playbook execution results
		if _, err := os.Stat("/tmp/testfile.txt"); os.IsNotExist(err) {
			log.Printf("Warning: /tmp/testfile.txt was not created after playbook run")
		} else if err == nil {
			log.Printf("Success: /tmp/testfile.txt exists")
		} else {
			log.Printf("Error checking /tmp/testfile.txt: %v", err)
		}

		// 7. Wait for next cycle
		log.Printf("Reconciliation cycle completed. Waiting %v before next cycle", interval)
		time.Sleep(interval)
	}
}
