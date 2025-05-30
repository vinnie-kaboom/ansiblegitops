package reconciler

import (
	"log"
	"os"
	"time"
)

type GitClient interface {
	Pull() (string, bool, error)
	Path() string
}

type AnsibleRunner interface {
	Run() error
}

type Reconciler struct {
	git     GitClient
	ansible AnsibleRunner
}

func NewReconciler(git GitClient, ansible AnsibleRunner) *Reconciler {
	// Set up more verbose logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	return &Reconciler{
		git:     git,
		ansible: ansible,
	}
}

func (r *Reconciler) Run(interval time.Duration) {
	log.Printf("Starting reconciliation loop with interval: %v", interval)
	log.Printf("Git repository path: %s", r.git.Path())

	// Check if we have necessary permissions
	if err := os.MkdirAll("/tmp", 0755); err != nil {
		log.Printf("Warning: Failed to verify /tmp directory permissions: %v", err)
	}

	for {
		log.Println("Beginning reconciliation cycle")

		log.Println("Pulling latest changes from Git repository")
		commit, changed, err := r.git.Pull()
		if err != nil {
			log.Printf("Error pulling repository: %v", err)
			time.Sleep(interval)
			continue
		}

		if changed {
			log.Printf("Repository was recreated or updated. Commit: %s", commit)
		} else {
			log.Println("Git pull completed successfully")
		}

		log.Println("Running Ansible playbook")
		if err := r.ansible.Run(); err != nil {
			log.Printf("Error running Ansible playbook: %v", err)
			time.Sleep(interval)
			continue
		}
		log.Println("Ansible playbook completed successfully")

		// Verify file creation
		if _, err := os.Stat("/tmp/testfile.txt"); os.IsNotExist(err) {
			log.Printf("Warning: /tmp/testfile.txt was not created after playbook run")
		} else if err == nil {
			log.Printf("Success: /tmp/testfile.txt exists")
		} else {
			log.Printf("Error checking /tmp/testfile.txt: %v", err)
		}

		log.Printf("Reconciliation cycle completed. Waiting %v before next cycle", interval)
		time.Sleep(interval)
	}
}
