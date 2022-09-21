package sv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/antihax/optional"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/briandowns/spinner"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/vballestra/sv/bitbucket"
	sshagent "github.com/xanzy/ssh-agent"
	ssh2 "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"net"
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

func (b *BitBucketSv) PullRequestStatus() (<-chan PullRequestStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BitBucketSv) Fetch() error {
	rep, giterr := git.PlainOpen(b.localRepo)
	if giterr != nil {
		return giterr
	}
	if sshagent.Available() {

		if a, _, err := sshagent.New(); err != nil {
			return fmt.Errorf("error creating SSH agent: %q", err)
		} else if sigs, err := a.Signers(); err != nil {
			return fmt.Errorf("While getting signers", err)
		} else {
			newSigs := make([]ssh2.Signer, 0)
			for _, s := range sigs {
				if k, ok := s.PublicKey().(*agent.Key); ok {
					// We don't use selector for now in BB (just because I'm lazy and don't want to spend
					// time on it)
					if k != nil /*b.sshKeySelector.MatchString(k.Comment)*/ {
						newSigs = append(newSigs, s)
					}
				}
			}

			if len(newSigs) > 0 {
				ag := &ssh.PublicKeysCallback{
					User: "git",
					Callback: func() ([]ssh2.Signer, error) {
						return newSigs, nil
					},
				}

				ag.HostKeyCallback = func(hostname string, remote net.Addr, key ssh2.PublicKey) error {
					return nil
				}

				sp := spinner.New(spinner.CharSets[55], time.Millisecond*50, spinner.WithSuffix(fmt.Sprintf(" Updating repository")))
				sp.Start()
				err = rep.Fetch(&git.FetchOptions{RemoteName: "origin", Auth: ag})
				sp.Stop()

				return err

			} else {
				return fmt.Errorf("Couldn't find any suitable keys, won't try to fetch the repo '%s'", b.localRepo)
			}
		}
	} else {
		return errors.New("Please use ssh agent")
	}
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

func (b BitbucketPullRequestWrapper) CreateComment(path string, commitId string, line int, isNew bool, body string) (Comment, error) {
	//TODO implement me
	panic("implement me")
}

func (b BitbucketPullRequestWrapper) GetLastCommitId() string {
	//TODO implement me
	panic("implement me")
}

func (b BitbucketPullRequestWrapper) ReplyToComment(comment Comment, replyText string) (Comment, error) {
	//TODO implement me
	panic("implement me")
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
