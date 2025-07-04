package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fraser-isbester/git-snap/pkg/types"
	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	client *github.Client
	owner  string
	repo   string
}

func NewGitHubClient(token, owner, repo string) *GitHubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &GitHubClient{
		client: client,
		owner:  owner,
		repo:   repo,
	}
}

func (gc *GitHubClient) GetCommits(since time.Time) ([]types.EnrichedCommit, error) {
	ctx := context.Background()

	opts := &github.CommitsListOptions{
		Since: since,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allCommits []types.EnrichedCommit

	for {
		commits, resp, err := gc.client.Repositories.ListCommits(ctx, gc.owner, gc.repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list commits: %w", err)
		}

		for _, commit := range commits {
			enrichedCommit, err := gc.enrichCommit(commit)
			if err != nil {
				continue
			}
			allCommits = append(allCommits, *enrichedCommit)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allCommits, nil
}

func (gc *GitHubClient) GetCommitDetails(sha string) (*types.EnrichedCommit, error) {
	ctx := context.Background()

	commit, _, err := gc.client.Repositories.GetCommit(ctx, gc.owner, gc.repo, sha, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	return gc.enrichCommit(commit)
}

func (gc *GitHubClient) enrichCommit(commit *github.RepositoryCommit) (*types.EnrichedCommit, error) {
	if commit == nil || commit.SHA == nil {
		return nil, fmt.Errorf("invalid commit data")
	}

	enriched := &types.EnrichedCommit{
		SHA:        *commit.SHA,
		Repository: gc.repo,
		Files:      []string{},
		Parents:    []string{},
	}

	if commit.Commit != nil {
		if commit.Commit.Author != nil {
			if commit.Commit.Author.Name != nil {
				enriched.Author = *commit.Commit.Author.Name
			}
			if commit.Commit.Author.Email != nil {
				enriched.AuthorEmail = *commit.Commit.Author.Email
			}
			if commit.Commit.Author.Date != nil {
				enriched.Timestamp = commit.Commit.Author.Date.Time
			}
		}

		if commit.Commit.Committer != nil && commit.Commit.Committer.Name != nil {
			enriched.Committer = *commit.Commit.Committer.Name
		}

		if commit.Commit.Message != nil {
			enriched.Message = *commit.Commit.Message
		}
	}

	if commit.Author != nil && commit.Author.Login != nil {
		enriched.Author = *commit.Author.Login
	}

	if commit.Stats != nil {
		if commit.Stats.Additions != nil {
			enriched.Additions = *commit.Stats.Additions
		}
		if commit.Stats.Deletions != nil {
			enriched.Deletions = *commit.Stats.Deletions
		}
	}

	for _, file := range commit.Files {
		if file.Filename != nil {
			enriched.Files = append(enriched.Files, *file.Filename)
		}
	}

	for _, parent := range commit.Parents {
		if parent.SHA != nil {
			enriched.Parents = append(enriched.Parents, *parent.SHA)
		}
	}

	prNumber := gc.extractPRNumber(enriched.Message)
	if prNumber != nil {
		enriched.PRNumber = prNumber
	}

	return enriched, nil
}

func (gc *GitHubClient) extractPRNumber(message string) *int {
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Merge pull request #") {
			parts := strings.Split(line, "#")
			if len(parts) > 1 {
				prStr := strings.Fields(parts[1])[0]
				if prNum, err := strconv.Atoi(prStr); err == nil {
					return &prNum
				}
			}
		}
	}
	return nil
}

func (gc *GitHubClient) GetPullRequests(since time.Time) ([]github.PullRequest, error) {
	ctx := context.Background()

	opts := &github.PullRequestListOptions{
		State: "all",
		Sort:  "updated",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allPRs []github.PullRequest

	for {
		prs, resp, err := gc.client.PullRequests.List(ctx, gc.owner, gc.repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list pull requests: %w", err)
		}

		for _, pr := range prs {
			if pr.UpdatedAt != nil && pr.UpdatedAt.After(since) {
				allPRs = append(allPRs, *pr)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

func (gc *GitHubClient) GetCommitsBetweenDates(start, end time.Time) ([]types.EnrichedCommit, error) {
	ctx := context.Background()

	opts := &github.CommitsListOptions{
		Since: start,
		Until: end,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allCommits []types.EnrichedCommit

	for {
		commits, resp, err := gc.client.Repositories.ListCommits(ctx, gc.owner, gc.repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list commits: %w", err)
		}

		for _, commit := range commits {
			enrichedCommit, err := gc.enrichCommit(commit)
			if err != nil {
				continue
			}
			allCommits = append(allCommits, *enrichedCommit)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allCommits, nil
}

func (gc *GitHubClient) TagCommit(sha, tag string) error {
	ctx := context.Background()

	_, _, err := gc.client.Git.CreateTag(ctx, gc.owner, gc.repo, &github.Tag{
		Tag:     &tag,
		Message: &tag,
		Object: &github.GitObject{
			SHA:  &sha,
			Type: github.String("commit"),
		},
	})

	return err
}

func (gc *GitHubClient) CreateLabel(name, color, description string) error {
	ctx := context.Background()

	label := &github.Label{
		Name:        &name,
		Color:       &color,
		Description: &description,
	}

	_, _, err := gc.client.Issues.CreateLabel(ctx, gc.owner, gc.repo, label)
	return err
}

func ParseGitHubURL(url string) (owner, repo string, err error) {
	url = strings.TrimSpace(url)

	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.TrimPrefix(url, "git@github.com:")
		url = strings.TrimSuffix(url, ".git")
		parts := strings.Split(url, "/")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}

	if strings.HasPrefix(url, "https://github.com/") {
		url = strings.TrimPrefix(url, "https://github.com/")
		url = strings.TrimSuffix(url, ".git")
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}

	return "", "", fmt.Errorf("invalid GitHub URL: %s", url)
}
