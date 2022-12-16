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
	"github.com/shurcooL/githubv4"
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

func (g *GitHubSv) GetRepositoryFullName() string {
	return fmt.Sprintf("%s/%s", g.owner, g.repo)
}

type PullRequestStatusResponse struct {
	Number      int64
	Title       string
	State       string
	BaseRefName string
	HeadRefName string
	Repository  struct {
		Name  string
		Owner struct {
			Login string
		}
	}
	Author struct {
		Login string
	}
	Reviews struct {
		Nodes []struct {
			Author struct {
				Login string
			}
			State string
		}
	}
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup struct {
					Contexts struct {
						CheckRunCountsByState []struct {
							State string
							Count int
						}
						StatusContextCountsByState []struct {
							State string
							Count int
						}
					}
				}
			}
		}
	}
}

func (g *GitHubSv) PullRequestStatus() (<-chan PullRequestStatus, error) {
	ch := make(chan PullRequestStatus)
	go func() {
		// Get my login
		if login, err := g.currentLogin(); err == nil {

			// Get My PRs
			if ch2, err := g.searchForPrStatus(fmt.Sprintf("type:pr state:open author:%s", login), true); err == nil {
				for r := range ch2 {
					ch <- r
				}
			}

			// Now get reviews as well
			if ch2, err := g.searchForPrStatus(fmt.Sprintf("type:pr state:open review-requested:%s", login), false); err == nil {
				for r := range ch2 {
					ch <- r
				}
			}
		}

		close(ch)
	}()

	return ch, nil
}

type GitHubPullRequestStatusWrapper struct {
	PullRequestStatusResponse
	isMine bool
}

func (g GitHubPullRequestStatusWrapper) IsMine() bool {
	return g.isMine
}

func (g GitHubPullRequestStatusWrapper) GetAuthor() string {
	return g.Author.Login
}

func (g GitHubPullRequestStatusWrapper) GetRepository() string {
	return fmt.Sprintf("%s/%s", g.Repository.Owner.Login, g.Repository.Name)
}

func (g GitHubPullRequestStatusWrapper) GetId() interface{} {
	return g.Number
}

func (g GitHubPullRequestStatusWrapper) GetTitle() string {
	return g.Title
}

func (g GitHubPullRequestStatusWrapper) GetStatus() string {
	return g.State
}

func (g GitHubPullRequestStatusWrapper) GetBranchName() string {
	return g.HeadRefName
}

func (g GitHubPullRequestStatusWrapper) GetBaseName() string {
	return g.BaseRefName
}

func (g GitHubPullRequestStatusWrapper) GetReviews() []Review {
	rev := make([]Review, 0)
	for _, r := range g.Reviews.Nodes {
		rev = append(rev, GitHubReview{api.PullRequestReview{
			Author: api.Author{
				Login: r.Author.Login,
			},
			AuthorAssociation:   "",
			Body:                "",
			SubmittedAt:         nil,
			IncludesCreatedEdit: false,
			ReactionGroups:      nil,
			State:               r.State,
			URL:                 "",
		}})
	}

	return rev
}

func (g GitHubPullRequestStatusWrapper) GetChecksByStatus() map[string]int {
	states := make(map[string]int)
	for _, s := range g.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.CheckRunCountsByState {
		states[s.State] = s.Count
	}
	return states
}

func (g GitHubPullRequestStatusWrapper) GetContextByStatus() map[string]int {
	states := make(map[string]int)
	for _, s := range g.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.StatusContextCountsByState {
		states[s.State] = s.Count
	}
	return states
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

func (g *GitHubSv) ListPullRequests(_ string) (<-chan PullRequest, error) {
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
			Line:     &line,
			Body:     &body,
		}); err == nil {
		return GitHubCommentWrapper{comment}, nil
	} else {
		return nil, err
	}
}

func (g *GitHubSv) graphQL(query string, variables map[string]interface{}, data interface{}) error {
	return g.gqlClient.GraphQL(g.host, query, variables, data)
}

func (g *GitHubSv) currentLogin() (string, error) {
	loginQuery := `query myLogin {
    viewer {
        login
    }
}`
	var resp struct {
		Viewer struct {
			Login string
		}
	}

	if err := g.graphQL(loginQuery, nil, &resp); err != nil {
		return "", err
	}

	return resp.Viewer.Login, nil
}

type IssueResNode struct {
	Id         string
	Repository struct {
		NameWithOwner string
	}
	Number int
}

func (g *GitHubSv) searchIssueNodes(nodeQuery string) <-chan IssueResNode {
	searchQuery := `query requestedReviews($prQuery: String!, $after: String)
{
    search(query: $prQuery, type: ISSUE, first: 5, after: $after) {
        issueCount
        pageInfo {
            endCursor
            hasNextPage
        }
        edges {
            node {
                ... on Node {
                    id
                }
            }
        }
    }
}`
	args := make(map[string]interface{})
	args["prQuery"] = nodeQuery

	res := make(chan IssueResNode)

	go func() {

		var resp struct {
			Search struct {
				IssueCount int
				PageInfo   struct {
					EndCursor   string
					HasNextPage bool
				}
				Edges []struct {
					Node IssueResNode
				}
			}
		}

		cont := true
		for cont {
			if err := g.graphQL(searchQuery, args, &resp); err == nil {

				for _, n := range resp.Search.Edges {
					res <- n.Node
				}

				if resp.Search.PageInfo.HasNextPage {
					args["after"] = resp.Search.PageInfo.EndCursor
				} else {
					cont = false
				}

			} else {
				cont = false
			}
		}

		close(res)
	}()

	return res
}

func (g *GitHubSv) searchForPrStatus(prQuery string, isMine bool) (<-chan PullRequestStatus, error) {
	singlePrQuery := `
query singleStatus($ids: [ID!]!) {
    nodes(ids: $ids) {
        ... on PullRequest {

            number
            title
            state
			repository {
                name
                owner {
                    login
                }
            }
            author {
                login
            }
            baseRefName
            headRefName
            reviews(first: 5) {
                nodes {
                    author {
                        login
                    }
                    state
                }
            }
            commits(last: 1) {
                nodes {
                    commit {
                        statusCheckRollup {
                            contexts {
                                checkRunCountsByState {
                                    state
                                    count
                                }
                                statusContextCountsByState {
                                    state
                                    count
                                }
                            }
                        }
                    }
                }

            }
        }
    }
}
`
	var resp struct {
		Nodes []PullRequestStatusResponse
	}

	ch := make(chan PullRequestStatus)
	go func() {

		ids := make([]string, 0)
		for nd := range g.searchIssueNodes(prQuery) {
			ids = append(ids, nd.Id)
			if len(ids) > 5 {
				if err := g.graphQL(singlePrQuery, map[string]interface{}{"ids": ids}, &resp); err == nil {
					for _, pr := range resp.Nodes {
						ch <- GitHubPullRequestStatusWrapper{pr, isMine}
					}
				}
				ids = make([]string, 0)
			}
		}

		if err := g.graphQL(singlePrQuery, map[string]interface{}{"ids": ids}, &resp); err == nil {
			for _, pr := range resp.Nodes {
				ch <- GitHubPullRequestStatusWrapper{pr, isMine}
			}
		}

		close(ch)
	}()

	return ch, nil
}

func (g GitHubPullRequest) ReplyToComment(comment Comment, replyText string) (Comment, error) {
	if c, ok := comment.(GitHubQLThreadCommentWrapper); ok {
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
			"commentId": c.comment.Id,
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
	} else if c, ok := comment.(GitHubCommentWrapper); ok {
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
				if (checks.State != "" || checks.Status != "") && checks.Name != "" {
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
	if g.TypeName == "CheckRun" {
		return g.DetailsURL
	}
	return g.TargetURL
}

func (g GitHubCheck) GetName() string {
	if g.TypeName == "CheckRun" {
		return g.Name
	}
	return g.Context
}

func (g GitHubCheck) GetStatus() string {
	if g.TypeName == "CheckRun" {
		if g.Status != "COMPLETED" {
			return g.Status
		}
		return g.Conclusion
	}
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

type UserInfo struct {
	Login string
}

func (u UserInfo) GetDisplayName() string {
	return u.Login
}

type ReactionInfo struct {
	Content   githubv4.ReactionContent
	CreatedAt time.Time
	User      UserInfo
}

func (r ReactionInfo) GetAuthor() Author {
	return r.User
}

func (r ReactionInfo) GetCreatedOn() time.Time {
	return r.CreatedAt
}

type ReactionsInfo struct {
	TotalCount int
	Nodes      []ReactionInfo
}

func (i ReactionsInfo) toReactions() Reactions {
	res := make(Reactions)

	for _, r := range i.Nodes {
		var l []Reaction
		if ll, ok := res[string(r.Content)]; !ok {
			l = ll
		} else {
			l = make([]Reaction, 0)
		}
		res[string(r.Content)] = append(l, r)
	}

	return res
}

type CommentInfo struct {
	Id        string
	Author    UserInfo
	Body      string
	BodyText  string
	BodyHTML  string
	Reactions ReactionsInfo
	CreatedAt time.Time
}

func (c CommentInfo) GetRaw() string {
	return c.Body
}

type PageInfo struct {
	HasNextPage bool
	EndCursor   string
}

type PullRequestCommentResponse struct {
	Repository struct {
		PullRequest struct {
			Comments struct {
				PageInfo   PageInfo
				TotalCount int
				Nodes      []CommentInfo
			}
		}
	}
}

type PullRequestCommentOrError struct {
	Comment CommentInfo
	error   error
}

type PullRequestThreadOrError struct {
	PullRequestThread PullRequestThread
	error             error
}

type PullRequestThreadComment struct {
	CommentInfo
	ReplyTo *struct {
		Id *string
	}
}

type PullRequestThread struct {
	Line              *int
	OriginalLine      *int
	Path              string
	DiffSide          *githubv4.DiffSide
	StartLine         *int
	StartDiffSide     *githubv4.DiffSide
	OriginalStartLine *int
	IsOutdated        bool
	Comments          struct {
		PageInfo   PageInfo
		TotalCount int
		Nodes      []PullRequestThreadComment
	}
}

func (g GitHubPullRequest) getPullRequestThreadComments() <-chan PullRequestThreadOrError {
	query := `query pullRequestThreads($number:Int!, $owner: String!, $name: String!, $commentAfter: String) {
    repository(owner: $owner,name: $name) {
        pullRequest(number: $number) {
            reviewThreads(first:5, after: $commentAfter) {
                pageInfo {
                    endCursor
                    hasNextPage
                }
                totalCount
                nodes {
                    line
                    originalLine
                    path
                    diffSide

                    startLine
                    startDiffSide
                    originalStartLine

                    isOutdated

                    comments(first: 100) {
                        pageInfo {
                            endCursor
                            hasNextPage
                        }
                        totalCount
                        nodes {
                            replyTo {
                                id
                            }
                            ...ReviewInfo
                            ...ReactionsInfo
                        }

                    }
                }
            }
        }
    }

}



fragment ReviewInfo on Comment {
    id
    author {
        login
    }
    body
    bodyText
    bodyHTML
    createdAt
}

fragment ReactionsInfo on Reactable {
    reactions(first: 20) {
        totalCount
        nodes {
            content
            createdAt
            user {
                login
            }
        }
    }
}
`
	args := map[string]interface{}{
		"number": g.Number,
		"owner":  g.sv.owner,
		"name":   g.sv.repo,
	}

	ch := make(chan PullRequestThreadOrError)

	go func() {
		cont := true
		var resp struct {
			Repository struct {
				PullRequest struct {
					ReviewThreads struct {
						PageInfo   PageInfo
						TotalCount int
						Nodes      []PullRequestThread
					}
				}
			}
		}
		for cont {
			if err := g.sv.graphQL(query, args, &resp); err != nil {
				cont = false
				ch <- PullRequestThreadOrError{error: err}
			}

			for _, c := range resp.Repository.PullRequest.ReviewThreads.Nodes {
				ch <- PullRequestThreadOrError{PullRequestThread: c}
			}
			if resp.Repository.PullRequest.ReviewThreads.PageInfo.HasNextPage {
				args["commentAfter"] = resp.Repository.PullRequest.ReviewThreads.PageInfo.EndCursor
			} else {
				cont = false
			}

		}
		close(ch)
	}()

	return ch
}

func (g GitHubPullRequest) getPullRequestComments() <-chan PullRequestCommentOrError {
	query := `
query pullRequestComments($number:Int!, $owner: String!, $name: String!, $commentAfter: String) {
    repository(owner: $owner,name: $name) {
        pullRequest(number: $number) {
            comments(first: 100, after: $commentAfter){
                pageInfo {
                    hasNextPage
                    endCursor
                }

                totalCount
                nodes {
					id
                    ...ReactionsInfo
                    ...ReviewInfo
                }
            }
        }
    }

}

fragment ReviewInfo on Comment {
    id
    author {
        login
    }
    body
    bodyText
    bodyHTML
    createdAt
}

fragment ReactionsInfo on Reactable {
    reactions(first: 20) {
        totalCount
        nodes {
            content
            createdAt
            user {
                login
            }
        }
    }
}
`
	args := map[string]interface{}{
		"number": g.Number,
		"owner":  g.sv.owner,
		"name":   g.sv.repo,
	}

	ch := make(chan PullRequestCommentOrError)

	go func() {
		cont := true
		var resp PullRequestCommentResponse
		for cont {
			if err := g.sv.graphQL(query, args, &resp); err != nil {
				cont = false
				ch <- PullRequestCommentOrError{error: err}
			}

			for _, c := range resp.Repository.PullRequest.Comments.Nodes {
				ch <- PullRequestCommentOrError{c, nil}
			}
			if resp.Repository.PullRequest.Comments.PageInfo.HasNextPage {
				args["commentAfter"] = resp.Repository.PullRequest.Comments.PageInfo.EndCursor
			} else {
				cont = false
			}

		}
		close(ch)
	}()

	return ch

}

type GitHubQLThreadCommentWrapper struct {
	thread  PullRequestThread
	comment PullRequestThreadComment
}

func (g GitHubQLThreadCommentWrapper) GetReactions() Reactions {
	return g.comment.GetReactions()
}

func (g GitHubQLThreadCommentWrapper) GetContent() CommentContent {
	return g.comment.CommentInfo
}

func (g GitHubQLThreadCommentWrapper) GetParentId() interface{} {
	if g.comment.ReplyTo != nil {
		return *(g.comment.ReplyTo.Id)
	} else {
		return nil
	}
}

func (g GitHubQLThreadCommentWrapper) GetId() interface{} {
	return g.comment.Id
}

func (g GitHubQLThreadCommentWrapper) GetUser() Author {
	return g.comment.Author
}

func (g GitHubQLThreadCommentWrapper) GetCreatedOn() time.Time {
	return g.comment.CreatedAt
}

func (g GitHubPullRequest) GetCommentsByLine() ([]Comment, map[string]map[int64][]Comment, error) {
	prComments := make([]Comment, 0)
	commentMap := make(map[string]map[int64][]Comment)

	//commentsById := make(map[string]*CommentInfo)

	for ghC := range g.getPullRequestThreadComments() {
		if ghC.error != nil {
			return nil, nil, ghC.error
		}

		th := ghC.PullRequestThread

		if th.IsOutdated {
			continue
		}

		for _, c := range th.Comments.Nodes {
			//commentsById[c.Id] = &c.CommentInfo
			cmt := GitHubQLThreadCommentWrapper{th, c}

			path := th.Path
			byLine, ok := commentMap[path]
			if !ok {
				byLine = make(map[int64][]Comment)
				commentMap[path] = byLine
			}

			var line int64

			if th.Line != nil {
				if *th.DiffSide == githubv4.DiffSideRight {
					line = -int64(*th.Line)
				} else {
					line = int64(*th.Line)
				}
			} else if th.OriginalLine != nil {
				// This is an obsolete comment, ignored for now
				continue
			} else {
				pterm.Fatal.Println("comment without a line")
			}

			lineComments, hasComments := byLine[line]
			if !hasComments {
				lineComments = make([]Comment, 0)
			}
			lineComments = append(lineComments, cmt)
			byLine[line] = lineComments
		}
	}

	for re := range g.getPullRequestComments() {
		if re.error != nil {
			return nil, nil, re.error
		}
		prComments = append(prComments, GithubQLCommentWrapper{re.Comment})
	}

	return prComments, commentMap, nil
}

type GithubQLCommentWrapper struct {
	CommentInfo
}

func (g CommentInfo) GetReactions() Reactions {
	return g.Reactions.toReactions()
}

func (g GithubQLCommentWrapper) GetRaw() string {
	return g.Body
}

func (g GithubQLCommentWrapper) GetContent() CommentContent {
	return g
}

func (g GithubQLCommentWrapper) GetParentId() interface{} {
	return nil
}

func (g GithubQLCommentWrapper) GetId() interface{} {
	return g.Id
}

func (g GithubQLCommentWrapper) GetUser() Author {
	return g.Author
}

func (g GithubQLCommentWrapper) GetCreatedOn() time.Time {
	return g.CreatedAt
}

type GitHubCommentWrapper struct {
	*gh.PullRequestComment
}

func (g GitHubCommentWrapper) GetReactions() Reactions {
	return make(Reactions)
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
	if err := changes.Encode(buf); err != nil {
		return nil, err
	}

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
