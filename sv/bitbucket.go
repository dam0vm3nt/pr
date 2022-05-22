package sv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/antihax/optional"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/vballestra/sv/bitbucket"
	"net/http"
	"strconv"
	"time"
)

type BitBucketSv struct {
	ctx       context.Context
	client    *bitbucket.APIClient
	repoSlug  string
	workspace string
	localRepo string
}

func (b *BitBucketSv) GetPullRequest(id string) (PullRequest, error) {
	n, _ := strconv.ParseInt(id, 10, 32)

	pr, resp, err2 := b.client.PullrequestsApi.RepositoriesWorkspaceRepoSlugPullrequestsPullRequestIdGet(b.ctx, int32(n), b.repoSlug, b.workspace)

	if err2 != nil || resp.StatusCode != 200 {
		return nil, err2
	}
	return BitbucketPullRequestWrapper{&pr, b}, nil
}

func NewBitBucketSv(username string, password string, repoSlug string, workspace string, repo string) Sv {
	cfg := bitbucket.NewConfiguration()
	cfg.HTTPClient = &http.Client{}
	auth := bitbucket.BasicAuth{UserName: username, Password: password}
	ctx := context.WithValue(context.Background(), bitbucket.ContextBasicAuth, auth)
	return &BitBucketSv{ctx: ctx, client: bitbucket.NewAPIClient(cfg), repoSlug: repoSlug, workspace: workspace,
		localRepo: repo,
	}
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
			res <- BitbucketPullRequestWrapper{&pr, b}
		}
		close(res)
	}()

	return res, nil
}

type BitbucketPullRequestWrapper struct {
	*bitbucket.Pullrequest
	client *BitBucketSv
}

func (b BitbucketPullRequestWrapper) GetReviews() ([]Review, error) {
	//TODO implement me
	panic("implement me")
}

func (b BitbucketPullRequestWrapper) GetChecks() ([]Check, error) {
	//TODO implement me
	panic("implement me")
}

func (b BitbucketPullRequestWrapper) GetBase() Branch {
	data := b.Destination.Branch.(map[string]interface{})
	return BitBucketBranchWrapper{&data}
}

func (b BitbucketPullRequestWrapper) GetCommentsByLine() ([]Comment, map[string]map[int64][]Comment, error) {
	sv := b.client
	cl := sv.client
	ctx := sv.ctx

	comments, _, err5 := cl.PullrequestsApi.RepositoriesWorkspaceRepoSlugPullrequestsPullRequestIdCommentsGet(ctx, b.Id, sv.repoSlug, sv.workspace)
	if err5 != nil {
		return nil, nil, err5
	}
	paginatedComments := PaginatedPullRequestComments{&comments}
	commentsChan := Paginate[bitbucket.PullrequestComment, bitbucket.PaginatedPullrequestComments](ctx, paginatedComments)

	commentMap := make(map[string]map[int64][]Comment)
	prComments := make([]Comment, 0)
	for comment := range commentsChan {
		if comment.Deleted {
			continue
		}
		cmt := BitbucketComment{&bitbucket.Comment{
			Id:        comment.Id,
			CreatedOn: comment.CreatedOn,
			Content:   comment.Content,
			User:      comment.User,
			Parent:    comment.Parent,
		}}
		if inline, ok := comment.Inline.(map[string]interface{}); ok {
			to := int64(inline["to"].(float64))
			path := inline["path"].(string)

			commentsByPath, hasPath := commentMap[path]
			if !hasPath {
				commentsByPath = make(map[int64][]Comment)
				commentMap[path] = commentsByPath
			}
			commentsByLine, hasLine := commentsByPath[to]
			if !hasLine {
				commentsByLine = make([]Comment, 0)
			}
			commentsByLine = append(commentsByLine, cmt)
			commentsByPath[to] = commentsByLine
		} else {
			prComments = append(prComments, cmt)
		}
	}

	return prComments, commentMap, nil
}

type BitbucketComment struct {
	*bitbucket.Comment
}

func (b BitbucketComment) GetParentId() interface{} {
	if b.Parent != nil {
		return b.Parent.Id
	}
	return nil
}

func (b BitbucketComment) GetId() interface{} {
	return b.Id
}

type BitbucketUserWrapper struct {
	*bitbucket.User
}

func (b BitbucketUserWrapper) GetDisplayName() string {
	return b.DisplayName
}

func (b BitbucketComment) GetUser() Author {
	return BitbucketUserWrapper{b.User}
}

func (b BitbucketComment) GetCreatedOn() time.Time {
	return b.CreatedOn
}

func (b BitbucketComment) GetContent() CommentContent {
	content := b.Content.(map[string]interface{})

	return BitbucketCommentContent{content}
}

type BitbucketCommentContent struct {
	content map[string]interface{}
}

func (b BitbucketCommentContent) GetRaw() string {
	return b.content["raw"].(string)
}

func (b BitbucketPullRequestWrapper) GetDiff() ([]*gitdiff.File, error) {
	sv := b.client
	cl := sv.client
	ctx := sv.ctx

	diff, _, err4 := cl.PullrequestsApi.RepositoriesWorkspaceRepoSlugPullrequestsPullRequestIdDiffGet(ctx, b.Id, sv.repoSlug, sv.workspace)
	if err4 != nil {
		return nil, err4
	}
	files, _, err5 := gitdiff.Parse(bytes.NewBuffer(diff))
	if err5 != nil {
		return nil, err5
	}

	return files, nil
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
