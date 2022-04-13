package sv

import (
	"context"
	"errors"
	"fmt"
	"github.com/antihax/optional"
	"github.com/vballestra/gobb-cli/bitbucket"
	"net/http"
	"time"
)

type BitBucketSv struct {
	ctx       context.Context
	client    *bitbucket.APIClient
	repoSlug  string
	workspace string
}

func NewBitBucketSv(username string, password string, repoSlug string, workspace string) Sv {
	cfg := bitbucket.NewConfiguration()
	cfg.HTTPClient = &http.Client{}
	auth := bitbucket.BasicAuth{UserName: username, Password: password}
	ctx := context.WithValue(context.Background(), bitbucket.ContextBasicAuth, auth)
	return &BitBucketSv{ctx: ctx, client: bitbucket.NewAPIClient(cfg), repoSlug: repoSlug, workspace: workspace}
}

func (b *BitBucketSv) ListPullRequests(prsQuery string) (<-chan PullRequest, error) {
	vars := bitbucket.PullrequestsApiRepositoriesWorkspaceRepoSlugPullrequestsGetOpts{
		State: optional.NewString("ACTIVE"),
	}
	if len(prsQuery) > 0 {
		vars.Q = optional.NewString(prsQuery)
	}
	prs, resp, err := b.client.PullrequestsApi.RepositoriesWorkspaceRepoSlugPullrequestsGet(b.ctx, b.repoSlug, b.workspace, &vars)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Status code = %d", resp.StatusCode))
	}

	res := make(chan PullRequest)

	go func() {
		for pr := range Paginate[bitbucket.Pullrequest, bitbucket.PaginatedPullrequests](b.ctx, PaginatedPullrequests{&prs}) {
			res <- BitbucketPullRequestWrapper{&pr}
		}
		close(res)
	}()

	return res, nil
}

type BitbucketPullRequestWrapper struct {
	*bitbucket.Pullrequest
}

func (b BitbucketPullRequestWrapper) GetState() string {
	return b.State
}

func (b BitbucketPullRequestWrapper) GetCreatedOn() time.Time {
	return b.CreatedOn
}

func (b BitbucketPullRequestWrapper) GetAuthor() Author {
	return BitbucketAuthorWrapper{b.Author}
}

type BitbucketAuthorWrapper struct {
	*bitbucket.Account
}

func (b BitbucketAuthorWrapper) GetDisplayName() string {
	return b.DisplayName
}

func (b BitbucketPullRequestWrapper) GetId() interface{} {
	return b.Id
}

func (b BitbucketPullRequestWrapper) GetTitle() string {
	return b.Title
}

func (b BitbucketPullRequestWrapper) GetBranch() Branch {
	data := b.Source.Branch.(map[string]interface{})
	return BitBucketBranchWrapper{&data}
}

type BitBucketBranchWrapper struct {
	data *map[string]interface{}
}

func (b BitBucketBranchWrapper) GetName() string {
	return (*b.data)["name"].(string)
}

func (b *BitBucketSv) getPullRequest(id string) PullRequest {
	//TODO implement me
	panic("implement me")
}

type PaginatedPullrequests struct {
	*bitbucket.PaginatedPullrequests
}

func (p PaginatedPullrequests) GetContainer() *bitbucket.PaginatedPullrequests {
	return p.PaginatedPullrequests
}

func (p PaginatedPullrequests) GetNext() string {
	return p.Next
}

func (p PaginatedPullrequests) GetPages() int32 {
	return p.Size
}

func (p PaginatedPullrequests) GetValues() []bitbucket.Pullrequest {
	return p.Values
}

type PaginatedPullRequestComments struct {
	*bitbucket.PaginatedPullrequestComments
}

func (p PaginatedPullRequestComments) GetContainer() *bitbucket.PaginatedPullrequestComments {
	return p.PaginatedPullrequestComments
}

func (p PaginatedPullRequestComments) GetNext() string {
	return p.Next
}

func (p PaginatedPullRequestComments) GetPages() int32 {
	return p.Size
}

func (p PaginatedPullRequestComments) GetValues() []bitbucket.PullrequestComment {
	return p.Values
}
