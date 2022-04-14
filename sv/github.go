package sv

import (
	"context"
	"errors"
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	gh "github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
	"time"
)

type GitHubSv struct {
	ctx    context.Context
	client *gh.Client
	owner  string
	repo   string
}

func (g *GitHubSv) GetPullRequest(id string) (PullRequest, error) {
	//TODO implement me
	panic("implement me")
}

func NewGitHubSv(token string) Sv {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	cl := gh.NewClient(tc)
	return &GitHubSv{
		ctx:    ctx,
		client: cl,
		owner:  "Latch",
		repo:   "latch-cortex",
	}
}

func (g *GitHubSv) ListPullRequests(query string) (<-chan PullRequest, error) {
	opts := &gh.PullRequestListOptions{}
	if res, resp, err := g.client.PullRequests.List(g.ctx, g.owner, g.repo, opts); err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Status == %d", resp.StatusCode))
	} else {
		ch := make(chan PullRequest)
		go func() {
			for len(res) > 0 && err == nil {
				for _, pr := range res {
					ch <- GitHubPullRequest{pr}
				}
				opts.Page += 1
				res, _, err = g.client.PullRequests.List(g.ctx, g.owner, g.repo, opts)
			}
			close(ch)
		}()

		return ch, nil
	}
}

type GitHubPullRequest struct {
	*gh.PullRequest
}

func (g GitHubPullRequest) GetCommentsByLine() ([]Comment, map[string]map[int64][]Comment, error) {
	//TODO implement me
	panic("implement me")
}

func (g GitHubPullRequest) GetDiff() ([]*gitdiff.File, error) {
	//TODO implement me
	panic("implement me")
}

func (g GitHubPullRequest) GetBase() Branch {
	//TODO implement me
	panic("implement me")
}

func (g GitHubPullRequest) GetBranch() Branch {
	return GitHubBranch{g.PullRequest.Head}
}

type GitHubBranch struct {
	*gh.PullRequestBranch
}

func (g GitHubBranch) GetName() string {
	return *g.Ref
}

func (g GitHubPullRequest) GetId() interface{} {
	return *g.Number
}

type GitHubAuthor struct {
	*gh.User
}

func (g GitHubAuthor) GetDisplayName() string {
	return g.GetLogin()
}

func (g GitHubPullRequest) GetAuthor() Author {
	return GitHubAuthor{g.User}
}

func (g GitHubPullRequest) GetCreatedOn() time.Time {
	return g.GetCreatedAt()
}
