package covertool

import (
	"context"

	"github.com/xanzy/go-gitlab"
)

const (
	PipelineName = "alauda-pipeline-cover"
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

func (t *Tool) Read(ctx context.Context, ref string) (coverage float64, error error) {
	sha, err := t.getLatestCommitFromRef(ctx, ref)
	if err != nil {
		return coverage, err
	}

	statusList, _, err := t.cli.Commits.GetCommitStatuses(t.projectID, sha, &gitlab.GetCommitStatusesOptions{
		All: gitlab.Bool(true),
	}, gitlab.WithContext(ctx))
	for _, status := range statusList {
		if status.Name != PipelineName {
			continue
		}
		if status.Coverage > coverage {
			coverage = status.Coverage
		}
	}

	return coverage, nil
}

func (t *Tool) Write(ctx context.Context, ref string, sha string, coverage float64) (err error) {
	if sha == "" {
		sha, err = t.getLatestCommitFromRef(ctx, ref)
		if err != nil {
			return err
		}
	}
	_, _, err = t.cli.Commits.SetCommitStatus(t.projectID, sha, &gitlab.SetCommitStatusOptions{
		State:    gitlab.Success,
		Ref:      gitlab.String(ref),
		Name:     gitlab.String("alauda-pipeline-cover"),
		Coverage: &coverage,
	}, gitlab.WithContext(ctx))
	if err != nil {
		return err
	}
	return nil
}
