package covertool

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/xanzy/go-gitlab"
)

type Tool struct {
	projectID string
	cli       *gitlab.Client
}

func New(baseUrl, token, projectID string) (*Tool, error) {
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(baseUrl))
	if err != nil {
		return nil, err
	}

	return &Tool{
		projectID: projectID,
		cli:       client,
	}, nil
}

func (t *Tool) getLatestCommitFromRef(ctx context.Context, ref string) (string, error) {
	commit, _, err := t.cli.Commits.GetCommit(t.projectID, ref, gitlab.WithContext(ctx))
	if err != nil {
		return "", err
	}
	return commit.ID, nil
}

// CommitStatus represents a GitLab commit status.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/commits.html#get-the-status-of-a-commit
type CommitStatus struct {
	ID           int           `json:"id"`
	SHA          string        `json:"sha"`
	Ref          string        `json:"ref"`
	Status       string        `json:"status"`
	CreatedAt    *time.Time    `json:"created_at"`
	StartedAt    *time.Time    `json:"started_at"`
	FinishedAt   *time.Time    `json:"finished_at"`
	Name         string        `json:"name"`
	AllowFailure bool          `json:"allow_failure"`
	Coverage     *float64      `json:"coverage"`
	Author       gitlab.Author `json:"author"`
	Description  string        `json:"description"`
	TargetURL    string        `json:"target_url"`
}

func (t *Tool) GetCommitStatuses(pid string, sha string, opt *gitlab.GetCommitStatusesOptions, options ...gitlab.RequestOptionFunc) ([]*CommitStatus, *gitlab.Response, error) {
	u := fmt.Sprintf("projects/%s/repository/commits/%s/statuses", gitlab.PathEscape(pid), url.PathEscape(sha))

	req, err := t.cli.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var cs []*CommitStatus
	resp, err := t.cli.Do(req, &cs)
	if err != nil {
		return nil, resp, err
	}

	return cs, resp, err
}

func (t *Tool) Read(ctx context.Context, pipeline, ref string) (coverage float64, error error) {
	sha, err := t.getLatestCommitFromRef(ctx, ref)
	if err != nil {
		return 0, fmt.Errorf("error get latest commit hash from %q: %w", ref, err)
	}

	statusList, _, err := t.GetCommitStatuses(
		t.projectID, sha, &gitlab.GetCommitStatusesOptions{
			Name: &pipeline,
			All:  gitlab.Bool(true),
		}, gitlab.WithContext(ctx))
	if err != nil {
		return 0, fmt.Errorf("error get commit status: %w", err)
	}

	for _, status := range statusList {
		if status.Coverage == nil {
			continue
		}
		if *status.Coverage > coverage {
			coverage = *status.Coverage
		}
	}

	return coverage, nil
}

func (t *Tool) Write(ctx context.Context, pipeline, ref, sha string, coverage float64) (err error) {
	if sha == "" {
		sha, err = t.getLatestCommitFromRef(ctx, ref)
		if err != nil {
			return fmt.Errorf("error get latest commit hash from %q: %w", ref, err)
		}
	}
	_, _, err = t.cli.Commits.SetCommitStatus(
		t.projectID, sha, &gitlab.SetCommitStatusOptions{
			State:    gitlab.Success,
			Ref:      &ref,
			Name:     &pipeline,
			Coverage: &coverage,
		}, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("error set commit status: %w", err)
	}
	return nil
}
