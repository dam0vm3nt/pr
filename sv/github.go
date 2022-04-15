package sv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	gh "github.com/google/go-github/v43/github"
	"github.com/pterm/pterm"
	"golang.org/x/oauth2"
	"net/http"
	"strconv"
	"time"
)

type GitHubSv struct {
	ctx       context.Context
	client    *gh.Client
	owner     string
	repo      string
	tc        *http.Client
	localRepo string
}

func (g *GitHubSv) GetPullRequest(id string) (PullRequest, error) {
	num, _ := strconv.ParseInt(id, 10, 32)
	pr, _, err := g.client.PullRequests.Get(g.ctx, g.owner, g.repo, int(num))
	if err != nil {
		return nil, err
	}

	return GitHubPullRequest{pr, g}, nil
}

func NewGitHubSv(token string, repo string) Sv {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	cl := gh.NewClient(tc)
	return &GitHubSv{
		ctx:       ctx,
		client:    cl,
		owner:     "Latch",
		repo:      "latch-cortex",
		tc:        tc,
		localRepo: repo,
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
					ch <- GitHubPullRequest{pr, g}
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
	sv *GitHubSv
}

func (g GitHubPullRequest) GetCommentsByLine() ([]Comment, map[string]map[int64][]Comment, error) {
	cl := g.sv.client

	opts := &gh.PullRequestListCommentsOptions{}
	opts.Page = 1

	commentsChan := make(chan *gh.PullRequestComment)

	// Get all the comments
	go func() {
		hasMore := true
		for hasMore {
			comments, _, err := cl.PullRequests.ListComments(g.sv.ctx, g.sv.owner, g.sv.repo, g.GetNumber(), opts)
			if err != nil {
				break
			}

			for _, ghC := range comments {
				commentsChan <- ghC
			}

			hasMore = len(comments) > 0
			opts.Page += 1
		}

		close(commentsChan)
	}()

	prComments := make([]Comment, 0)
	commentMap := make(map[string]map[int64][]Comment)

	for ghC := range commentsChan {
		cmt := GitHubCommentWrapper{ghC}
		if ghC.Path != nil {
			path := *ghC.Path
			byLine, ok := commentMap[path]
			if !ok {
				byLine = make(map[int64][]Comment)
				commentMap[path] = byLine
			}

			line := int64(*ghC.Line)
			lineComments, hasComments := byLine[line]
			if !hasComments {
				lineComments = make([]Comment, 0)
			}
			lineComments = append(lineComments, cmt)
			byLine[line] = lineComments
		} else {
			prComments = append(prComments, cmt)
		}
	}

	return prComments, commentMap, nil
}

type GitHubCommentWrapper struct {
	*gh.PullRequestComment
}

func (g GitHubCommentWrapper) GetContent() CommentContent {
	return GitHubCommentContentWrapper{g.Body}
}

type GitHubCommentContentWrapper struct {
	*string
}

func (g GitHubCommentContentWrapper) GetRaw() string {
	return *g.string
}

func (g GitHubCommentWrapper) GetParentId() interface{} {
	return g.InReplyTo
}

func (g GitHubCommentWrapper) GetId() interface{} {
	return *g.ID
}

func (g GitHubCommentWrapper) GetUser() Author {
	return GitHubAuthor{g.User}
}

func (g GitHubCommentWrapper) GetCreatedOn() time.Time {
	return *g.CreatedAt
}

func (g GitHubPullRequest) GetDiff() ([]*gitdiff.File, error) {

	agent, _ := ssh.NewSSHAgentAuth("vittorioballestra@vittorioballestraMacBookPro11,4")

	rep, giterr := git.PlainOpen(g.sv.localRepo)
	if giterr != nil {
		return nil, giterr
	}
	err1 := rep.Fetch(&git.FetchOptions{RemoteName: "origin", Auth: agent})
	if err1 != nil {
		pterm.Warning.Println(err1)
		// return nil, err1
	}

	base, _ := rep.Branch(*g.Base.Ref)
	br, _ := rep.Branch(*g.Head.Ref)

	refBase, _ := rep.Reference(base.Merge, true)
	refBr, _ := rep.Reference(br.Merge, true)

	refBaseHash := refBase.Hash()
	refBrHash := refBr.Hash()

	cBr, _ := rep.CommitObject(refBrHash)
	cBase, _ := rep.CommitObject(refBaseHash)

	baseTree, err2 := cBase.Tree()
	if err2 != nil {
		pterm.Error.Println(err2)
	}
	brTree, err3 := cBr.Tree()
	if err3 != nil {
		pterm.Error.Println(err3)
	}

	changes, _ := baseTree.Patch(brTree)
	buf := &bytes.Buffer{}
	changes.Encode(buf)
	str := changes.String()
	fmt.Println(str)

	files, _, err5 := gitdiff.Parse(buf)
	if err5 != nil {
		return nil, err5
	}

	return files, nil

}

func (g GitHubPullRequest) GetBase() Branch {
	return GitHubBranch{g.Base}
}

func (g GitHubPullRequest) GetBranch() Branch {
	return GitHubBranch{g.Head}
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
