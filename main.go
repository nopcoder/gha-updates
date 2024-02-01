package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

type Workflow struct {
	Jobs map[string]Job `yaml:"jobs"`
}

type Job struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Uses string `yaml:"uses,omitempty"`
}

// parseGitHubActions read and parse the YAML file, returning a map of unique "uses"
func parseGitHubActions(filename string) (map[string]bool, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML into our Workflow struct
	var workflow Workflow
	err = yaml.Unmarshal(data, &workflow)
	if err != nil {
		return nil, err
	}

	// Iterate through the jobs and steps to find unique "uses"
	uniqueUses := make(map[string]bool)
	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			if step.Uses != "" {
				uniqueUses[step.Uses] = true
			}
		}
	}

	return uniqueUses, nil
}

type ActionsUpdater struct {
	repos map[string]string
}

func NewActionsUpdater() *ActionsUpdater {
	return &ActionsUpdater{
		repos: make(map[string]string),
	}
}

func (u *ActionsUpdater) Scan(filename string) error {
	uniqueUses, err := parseGitHubActions(filename)
	if err != nil {
		return err
	}

	// Iterate through the unique "uses" and check for updates
	for use := range uniqueUses {
		latest, err := u.Update(use)
		if err != nil {
			return err
		}
		if latest != "" {
			fmt.Printf("- %s -> %s\n", use, latest)
		}
	}

	return nil
}

func (u *ActionsUpdater) Update(use string) (string, error) {
	// split use to repository and tag
	repo, tag, ok := strings.Cut(use, "@")
	if !ok || repo == "" || tag == "" {
		return "", nil
	}
	// re-format the repository name by using the owner and repo name
	safeRepo := repo
	repoParts := strings.Split(repo, "/")
	if len(repoParts) > 2 {
		safeRepo = strings.Join(repoParts[:2], "/")
	}

	latest, found := u.repos[safeRepo]
	if !found {
		latestTag, err := u.getLatestTag(safeRepo)
		if err != nil {
			return "", err
		}

		latest = latestTag
		u.repos[safeRepo] = latestTag
	}

	if semver.Compare(latest, tag) > 0 {
		return latest, nil
	}
	return "", nil
}

func (u *ActionsUpdater) getLatestTag(repo string) (string, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", err
	}
	out, err := exec.Command(gitPath, "-c", "versionsort.suffix=-", "ls-remote", "--tags", "--sort=v:refname", "https://github.com/"+repo+".git").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			fmt.Printf("[DEBUG] stderr: %s\n", string(exitErr.Stderr))
		}
		return "", err
	}

	tags := parseOutputTags(string(out))
	var latest string
	if len(tags) > 0 {
		latest = tags[len(tags)-1]
	}
	return latest, nil
}

func parseOutputTags(out string) []string {
	var tags []string
	lines := strings.Split(strings.ReplaceAll(out, "\r", ""), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if !strings.HasPrefix(fields[1], "refs/tags/") {
			continue
		}
		tag := fields[1][len("refs/tags/"):]
		if tag == "" {
			continue
		}
		tags = append(tags, tag)
	}
	return tags
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <workflow.yaml>\n", os.Args[0])
		os.Exit(1)
	}

	au := NewActionsUpdater()
	for _, fn := range os.Args[1:] {
		fmt.Printf("File: %s\n", fn)
		err := au.Scan(fn)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
	}
}
