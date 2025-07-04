package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fraser-isbester/git-snap/pkg/types"
)

type GitClient struct {
	repoPath string
}

func NewGitClient(repoPath string) *GitClient {
	return &GitClient{repoPath: repoPath}
}

func (g *GitClient) GetCommits(since time.Time) ([]types.EnrichedCommit, error) {
	args := []string{
		"log",
		"--format=%H|%an|%ae|%cn|%ct|%s|%P",
		"--numstat",
		"--since=" + since.Format("2006-01-02"),
		"--all",
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	return g.parseCommits(string(output))
}

func (g *GitClient) GetCommitsByRange(fromCommit, toCommit string) ([]types.EnrichedCommit, error) {
	args := []string{
		"log",
		"--format=%H|%an|%ae|%cn|%ct|%s|%P",
		"--numstat",
		fmt.Sprintf("%s..%s", fromCommit, toCommit),
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	return g.parseCommits(string(output))
}

func (g *GitClient) GetCommitDetails(sha string) (*types.EnrichedCommit, error) {
	args := []string{
		"show",
		"--format=%H|%an|%ae|%cn|%ct|%s|%P",
		"--numstat",
		"--no-patch",
		sha,
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit details: %w", err)
	}

	commits, err := g.parseCommits(string(output))
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("commit not found: %s", sha)
	}

	return &commits[0], nil
}

func (g *GitClient) parseCommits(output string) ([]types.EnrichedCommit, error) {
	lines := strings.Split(output, "\n")
	var commits []types.EnrichedCommit

	var currentCommit *types.EnrichedCommit

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "|") && !strings.Contains(line, "\t") {
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
			}

			parts := strings.Split(line, "|")
			if len(parts) < 6 {
				continue
			}

			timestamp, err := strconv.ParseInt(parts[4], 10, 64)
			if err != nil {
				continue
			}

			parents := []string{}
			if len(parts) > 6 && parts[6] != "" {
				parents = strings.Split(parts[6], " ")
			}

			currentCommit = &types.EnrichedCommit{
				SHA:         parts[0],
				Author:      parts[1],
				AuthorEmail: parts[2],
				Committer:   parts[3],
				Timestamp:   time.Unix(timestamp, 0),
				Message:     parts[5],
				Parents:     parents,
				Repository:  g.getRepositoryName(),
				Branch:      g.getCurrentBranch(),
				Files:       []string{},
			}
		} else if currentCommit != nil {
			parts := strings.Split(line, "\t")
			if len(parts) == 3 {
				additions, _ := strconv.Atoi(parts[0])
				deletions, _ := strconv.Atoi(parts[1])
				filename := parts[2]

				currentCommit.Additions += additions
				currentCommit.Deletions += deletions
				currentCommit.Files = append(currentCommit.Files, filename)
			}
		}
	}

	if currentCommit != nil {
		commits = append(commits, *currentCommit)
	}

	return commits, nil
}

func (g *GitClient) getRepositoryName() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return filepath.Base(g.repoPath)
	}

	url := strings.TrimSpace(string(output))

	re := regexp.MustCompile(`([^/]+)\.git$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}

	return filepath.Base(g.repoPath)
}

func (g *GitClient) getCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(output))
}

func (g *GitClient) IsGitRepository() bool {
	_, err := os.Stat(filepath.Join(g.repoPath, ".git"))
	return err == nil
}

func (g *GitClient) GetBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-r")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var branches []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "->") {
			continue
		}

		if strings.HasPrefix(line, "origin/") {
			branches = append(branches, strings.TrimPrefix(line, "origin/"))
		}
	}

	return branches, nil
}

func (g *GitClient) GetCommitsBetweenDates(start, end time.Time) ([]types.EnrichedCommit, error) {
	args := []string{
		"log",
		"--format=%H|%an|%ae|%cn|%ct|%s|%P",
		"--numstat",
		"--since=" + start.Format("2006-01-02"),
		"--until=" + end.Format("2006-01-02"),
		"--all",
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	return g.parseCommits(string(output))
}

func FindGitRepository(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	for {
		gitPath := filepath.Join(absPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return absPath, nil
		}

		parent := filepath.Dir(absPath)
		if parent == absPath {
			break
		}
		absPath = parent
	}

	return "", fmt.Errorf("not a git repository: %s", path)
}
