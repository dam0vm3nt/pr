package sv

import "C"
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/briandowns/spinner"
	"github.com/cli/cli/v2/api"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	gh "github.com/google/go-github/v43/github"
	"github.com/pterm/pterm"
	sshagent "github.com/xanzy/ssh-agent"
	ssh2 "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/oauth2"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type GitHubSv struct {
	ctx            context.Context
	client         *gh.Client
	gqlClient      *api.Client
	owner          string
	repo           string
	tc             *http.Client
	localRepo      string
	host           string
	sshKeySelector *regexp.Regexp
}

func (g *GitHubSv) Fetch() error {
	rep, giterr := git.PlainOpen(g.localRepo)
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
					if g.sshKeySelector.MatchString(k.Comment) {
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
				return fmt.Errorf("Couldn't find any suitable keys, won't try to fetch the repo '%s'", g.localRepo)
			}
		}
	} else {
		return errors.New("Please use ssh agent")
	}
}

const githubDefaultHost = "github.com"

func (g *GitHubSv) GetPullRequest(id string) (PullRequest, error) {
	num, _ := strconv.ParseInt(id, 10, 32)
	pr, _, err := g.client.PullRequests.Get(g.ctx, g.owner, g.repo, int(num))
	if err != nil {
		return nil, err
	}

	return GitHubPullRequest{pr, g}, nil
}

func NewGitHubSv(token string, repo string, sshKeyComment string, owner string, name string) Sv {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	cl := gh.NewClient(tc)

	// Reading owner and repo from local workspace remote

	if re, err := regexp.Compile(sshKeyComment); err == nil {

		return &GitHubSv{
			ctx:            ctx,
			client:         cl,
			owner:          owner,
			repo:           name,
			tc:             tc,
			localRepo:      repo,
			gqlClient:      api.NewClientFromHTTP(tc),
			host:           githubDefaultHost,
			sshKeySelector: re,
		}
	} else {
		pterm.Fatal.Println("Error while compiling selector re '", sshKeyComment, "'", err)
		panic(err)
	}
}

func (g *GitHubSv) ListPullRequests(query string) (<-chan PullRequest, error) {
	opts := &gh.PullRequestListOptions{}
	opts.Page = 1
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

func (g GitHubPullRequest) GetLastCommitId() string {
	return g.Head.GetSHA()
}

func (g GitHubPullRequest) CreateComment(path string, commitId string, line int, isNew bool, body string) (Comment, error) {
	side := "LEFT"
	if isNew {
		side = "RIGHT"
	}
	if comment, _, err := g.sv.client.PullRequests.CreateComment(g.sv.ctx,
		g.sv.owner,
		g.sv.repo,
		g.GetNumber(),
		&gh.PullRequestComment{
			Path:     &path,
			CommitID: &commitId,
			Side:     &side,
			Position: &line,
			Body:     &body,
		}); err == nil {
		return GitHubCommentWrapper{comment}, nil
	} else {
		return nil, err
	}
}

func (g GitHubPullRequest) ReplyToComment(comment Comment, replyText string) (Comment, error) {
	if c, ok := comment.(GitHubCommentWrapper); ok {
		newDiscussionMutation := `
mutation newReview($prId: ID!) {
  addPullRequestReview(input: {pullRequestId: $prId}) { 
    pullRequestReview {
      id
    }
  }
}
`
		var newDiscussionMutationResponse struct {
			AddPullRequestReview struct {
				PullRequestReview struct {
					Id string
				}
			}
		}

		if err := g.sv.gqlClient.GraphQL(g.sv.host, newDiscussionMutation, map[string]interface{}{"prId": g.GetNodeID()}, &newDiscussionMutationResponse); err != nil {
			return nil, err
		}

		mutation := `
mutation replyTo($revId: ID!, $commentId: ID!, $body: String!) {
  addPullRequestReviewComment(
    input: {pullRequestReviewId: $revId, inReplyTo: $commentId, body: $body}
  ) { 
    comment {
      id
    }
  }
}`
		var resp1 struct {
			AddPullRequestReviewComment struct {
				Comment struct {
					Id string
				}
			}
		}

		if err := g.sv.gqlClient.GraphQL(g.sv.host, mutation, map[string]interface{}{
			"commentId": c.GetNodeID(),
			"revId":     newDiscussionMutationResponse.AddPullRequestReview.PullRequestReview.Id,
			"body":      replyText}, &resp1); err != nil {
			return nil, err
		}

		// Finally close the review and return

		closeReviewMutation := `
mutation closeReview($revId: ID!) {
  submitPullRequestReview(input: {pullRequestReviewId: $revId, event: COMMENT}) {
    clientMutationId
  }
}`
		var closeReviewMutationResponse struct {
			SubmitPullRequestReview struct {
				ClientMutationId *string
			}
		}

		if err := g.sv.gqlClient.GraphQL(g.sv.host, closeReviewMutation, map[string]interface{}{
			"revId": newDiscussionMutationResponse.AddPullRequestReview.PullRequestReview.Id,
		}, &closeReviewMutationResponse); err != nil {
			return nil, err
		}

		return nil, nil
	}

	if c, ok := comment.(GitHubCommentWrapper); ok {
		if cmt, _, err := g.sv.client.PullRequests.CreateComment(g.sv.ctx, g.sv.owner, g.sv.repo, g.GetNumber(), &gh.PullRequestComment{
			InReplyTo: c.ID,
			Position:  c.Position,
			// Path:      c.Path,
			Body: &replyText,
			// CommitID:  c.CommitID,
		}); err == nil {
			return GitHubCommentWrapper{cmt}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Illegal argument: not a github comment")
	}
}

func (g GitHubPullRequest) GetChecks() ([]Check, error) {
	type response struct {
		Repository *struct {
			PullRequest *struct {
				StatusCheckRollup *struct {
					Nodes []*struct {
						Commit *struct {
							StatusCheckRollup *struct {
								Contexts *struct {
									Nodes []*api.CheckContext
								}
							}
						}
					}
				}
			}
		}
	}

	var resp1 response

	err := g.sv.gqlClient.GraphQL(g.sv.host, `
query GetRepos($name: String!, $owner: String!, $number: Int!) {
	repository(name: $name, owner: $owner) {
		pullRequest(number: $number) {
			statusCheckRollup: commits(last: 1) {
				nodes {
					commit {
						statusCheckRollup {
							contexts(first:100) {
								nodes {
									__typename
									...on StatusContext {
										context,
										state,
										targetUrl
									},
									...on CheckRun {
										name,
										status,
										conclusion,
										startedAt,
										completedAt,
										detailsUrl
									}
								},
								pageInfo{hasNextPage,endCursor}
							}
						}
					}
				}
			}
		}
	}
}`, map[string]interface{}{"name": g.sv.repo, "owner": g.sv.owner, "number": g.GetNumber()}, &resp1)
	if err != nil {
		pterm.Fatal.Println(err)
	}
	result := make([]Check, 0)

	for _, rollups := range resp1.Repository.PullRequest.StatusCheckRollup.Nodes {

		if rollups != nil && rollups.Commit != nil && rollups.Commit.StatusCheckRollup != nil && rollups.Commit.StatusCheckRollup.Contexts != nil {
			for _, checks := range rollups.Commit.StatusCheckRollup.Contexts.Nodes {
				if checks.State != "" && checks.Name != "" {
					result = append(result, GitHubCheck{checks})
				}
			}
		}

	}

	return result, nil
}

type GitHubCheck struct {
	*api.CheckContext
}

func (g GitHubCheck) GetUrl() string {
	return g.TargetURL
}

func (g GitHubCheck) GetName() string {
	return g.Context
}

func (g GitHubCheck) GetStatus() string {
	return g.State
}

func (g GitHubPullRequest) GetReviews() ([]Review, error) {
	type response struct {
		Repository *struct {
			PullRequest *struct {
				Reviews *api.PullRequestReviews
			}
		}
	}

	var resp1 response

	err := g.sv.gqlClient.GraphQL(g.sv.host, `
query GetRepos($name: String!, $owner: String!, $number: Int!) {
	repository(name: $name, owner: $owner) {
		pullRequest(number: $number) {
			reviews(first: 100) {
				nodes {
					author { login }
					body
					bodyHTML
					bodyText
					state
					submittedAt	
				}
			}
		}
	}
}`, map[string]interface{}{"name": g.sv.repo, "owner": g.sv.owner, "number": g.GetNumber()}, &resp1)
	if err != nil {
		pterm.Fatal.Println(err)
	}
	result := make([]Review, 0)

	for _, rev := range resp1.Repository.PullRequest.Reviews.Nodes {

		//for _, rev1 := range rev.Nodes {
		result = append(result, GitHubReview{rev})
		//}

	}

	return result, nil
}

type GitHubReview struct {
	api.PullRequestReview
}

func (g GitHubReview) GetState() string {
	return g.State
}

func (g GitHubReview) GetAuthor() string {
	return g.Author.Login
}

func (g GitHubReview) GetSubmitedAt() time.Time {
	if ok := g.SubmittedAt; ok != nil {
		return *ok
	} else {
		return time.Now()
	}
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

	commentsById := make(map[int64]*gh.PullRequestComment)

	for ghC := range commentsChan {
		commentsById[*ghC.ID] = ghC
		cmt := GitHubCommentWrapper{ghC}
		if ghC.Path != nil {
			path := *ghC.Path
			byLine, ok := commentMap[path]
			if !ok {
				byLine = make(map[int64][]Comment)
				commentMap[path] = byLine
			}

			var line int64
			if ghC.Line != nil {
				line = int64(*ghC.Line)
			} else if ghC.OriginalLine != nil {
				line = -int64(*ghC.OriginalLine)
			} else {
				pterm.Fatal.Println("comment without a line")
			}

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
	if g.InReplyTo == nil {
		return nil
	}
	return *g.InReplyTo
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
	rep, giterr := git.PlainOpen(g.sv.localRepo)
	if giterr != nil {
		return nil, giterr
	}

	refBaseHash := plumbing.NewHash(*g.Base.SHA)
	refBrHash := plumbing.NewHash(*g.Head.SHA)

	cBr, err := rep.CommitObject(refBrHash)
	if err != nil {
		pterm.Fatal.Println("Cannot get the pr branch head commit, do you need to update your local repo ?", err)
	}
	cBase, err := rep.CommitObject(refBaseHash)
	if err != nil {
		pterm.Fatal.Println("Cannot get the base branch commit, do you need to update your local repo ?", err)
	}

	merge, err := cBr.MergeBase(cBase)
	if err != nil {
		pterm.Fatal.Println("Cannot find a merge base?!?")
	}
	if len(merge) != 1 {
		pterm.Fatal.Printfln("More than one merge base ?!? %s", merge)
	}

	baseTree, err2 := merge[0].Tree()
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
