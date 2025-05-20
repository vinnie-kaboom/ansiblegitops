 
# AnsibleGitOps

A simplistic GitOps service written in Go, inspired by Flux, for managing Ansible playbooks.

## Prerequisites
- Go 1.21+
- Ansible
- Git
- A Git repository with Ansible playbooks and inventory

## Setup
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/ansiblegitops-repo.git
   cd ansiblegitops-repo
   ```

## Up the Repository Structure: 

git clone https://github.com/yourusername/ansible-repo.git
cd ansible-repo
mkdir -p inventory playbooks

## Create an Inventory File:

inventory/hosts:
```ini
[webservers]
localhost ansible_connection=local
```

## Create a Sample Playbook: playbooks/site.yml:
```
---
- hosts: webservers
  tasks:
    - name: Create a test file
      file:
        path: /tmp/testfile.txt
        state: touch
        mode: '0644'
```      
## config.yaml:
```
git:
  url: "https://github.com/vinnie-kaboom/ansible-repo.git"
  branch: "main"
  poll_interval: 300
ansible:
  playbook_dir: "playbooks"
  inventory_file: "inventory/hosts"
```

## Step 4: Build and Run the Service
1.  Build the Binary: From the ansiblegitops directory:
    go build -o ansiblegitops ./cmd/ansiblegitops
2.  Run the Service:
    ./ansiblegitops

## Verify Output

1. Check the logs for messages like:
    ```json
    {"level":"info","message":"No changes in repository"}
    {"level":"info","commit":"abc123","message":"New commit detected, running playbooks"}
    {"level":"info","playbook":".../playbooks/site.yml","message":"Running Ansible playbook"}
    {"level":"info","commit":"abc123","message":"Reconciliation completed"}
    ```

2. Verify `/tmp/testfile.txt` exists on the target host (localhost in this case):
    ```bash
    ls -l /tmp/testfile.txt
    ```

## Test GitOps Workflow
1. Update the Playbook: Modify playbooks/site.yml in the Git repository to add a new task:
    ```yaml
    ---
    - hosts: webservers
      tasks:
        - name: Create a test file
          file:
            path: /tmp/testfile.txt
            state: touch
            mode: '0644'
        - name: Add content to test file
          copy:
            content: "Hello from AnsibleGitOps!"
            dest: /tmp/testfile.txt
    ```

2. Commit and Push:
    ```bash
    git add playbooks/site.yml
    git commit -m "Update playbook to add content"
    git push origin main
    ```


 ## Observe Reconciliation:
	•  Within 5 minutes (or sooner if you restart the service), the service will detect the new commit.
	•  It will run the updated playbook, adding content to /tmp/testfile.txt.
	•  Check the file:  cat /tmp/testfile.txt    
