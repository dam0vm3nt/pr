//go:generate go run github.com/Khan/genqlient
package sv

import "C"
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/antihax/optional"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/briandowns/spinner"
	"github.com/cli/cli/v2/api"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	gh "github.com/google/go-github/v43/github"
	"github.com/pterm/pterm"
	"github.com/shurcooL/githubv4"
	"github.com/vballestra/sv/sv/gh_utils"
	sshagent "github.com/xanzy/ssh-agent"
	ssh2 "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/oauth2"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

type idOrError struct {
	string
	error
}

func toUserIdsChan(repo *GitHubSv, logins []string) <-chan idOrError {
	ch := make(chan idOrError)

	go func() {
		for _, login := range logins {
			if resp, err := getUserIdByLogin(repo.ctx, login); err != nil {
				ch <- idOrError{error: err}
				break
			} else {
				ch <- idOrError{string: resp.User.Id}
			}
		}
		close(ch)
	}()

	return ch
}

func toLabelIdsChan(repo *GitHubSv, labels []string) <-chan idOrError {
	ch := make(chan idOrError)

	go func() {

		// Get the repo id first
		if repoIdResp, err := repositoryId(repo.ctx, repo.owner, repo.repo); err != nil {
			ch <- idOrError{error: err}
		} else {
			repoId := repoIdResp.Repository.Id
			for _, label := range labels {
				if resp, err := getLabelByName(repo.ctx, label, repo.owner, repo.repo); err != nil {
					ch <- idOrError{error: err}
					break
				} else if resp.Repository.Label != nil {
					ch <- idOrError{string: resp.Repository.Label.Id}
				} else if resp2, err := createLabel(repo.ctx, label, nil, "a0a0a0", repoId); err != nil {
					ch <- idOrError{error: err}
					break
				} else {
					ch <- idOrError{string: resp2.CreateLabel.Label.Id}
				}
			}
		}
		close(ch)
	}()

	return ch
}

func toUserIds(repo *GitHubSv, logins []string) ([]string, error) {
	return toIds(func() <-chan idOrError {
		return toUserIdsChan(repo, logins)
	})
}

func toLabelIds(repo *GitHubSv, labels []string) ([]string, error) {
	return toIds(func() <-chan idOrError {
		return toLabelIdsChan(repo, labels)
	})
}

func toIds(iter func() <-chan idOrError) ([]string, error) {
	userIds := make([]string, 0)

	for res := range iter() {
		if res.error != nil {
			return nil, res.error
		}
		userIds = append(userIds, res.string)
	}

	return userIds, nil
}

func (g *GitHubSv) resolveHeadBranch(headBranch optional.String) (string, error) {
	if headBranch.IsSet() {
		return headBranch.Value(), nil
	} else {
		return g.GetCurrentBranch()
	}
}

func (g *GitHubSv) GetDefaultBranch() (string, error) {
	if resp, err := defaultBranch(g.ctx, g.repo, g.owner); err != nil {
		return "", err
	} else {
		return resp.Repository.DefaultBranchRef.Name, nil
	}
}

func (g *GitHubSv) resolveBaseBranch(baseBranch optional.String) (string, error) {
	if baseBranch.IsSet() {
		return baseBranch.Value(), nil
	} else {
		return g.GetDefaultBranch()
	}
}

func (g *GitHubSv) resolveTitleAndDescription(titleOpt optional.String, descriptionOpt optional.String, headBranch string, baseBranch string) (string, string, error) {
	if titleOpt.IsSet() && descriptionOpt.IsSet() {
		return titleOpt.Value(), descriptionOpt.Value(), nil
	}

	// Check if we only have one commit
	rep, err := git.PlainOpen(g.localRepo)
	if err != nil {
		return "", "", err
	}

	// Get Head Commit
	br, err := rep.Branch(headBranch)
	if err != nil {
		return "", "", err
	}
	hdRef, err := rep.Reference(br.Merge, true)
	if err != nil {
		return "", "", err
	}
	hdCommit, err := rep.CommitObject(hdRef.Hash())
	if err != nil {
		return "", "", err
	}

	// Get Base Commit
	ba, err := rep.Branch(baseBranch)
	if err != nil {
		return "", "", err
	}
	baRef, err := rep.Reference(ba.Merge, true)
	if err != nil {
		return "", "", err
	}
	baCommit, err := rep.CommitObject(baRef.Hash())
	if err != nil {
		return "", "", err
	}

	mbs, err := hdCommit.MergeBase(baCommit)
	if err != nil {
		return "", "", err
	}
	if len(mbs) != 1 {
		return "", "", fmt.Errorf("cannot find unique common base between %v and %v", headBranch, baseBranch)
	}
	mb := mbs[0]
	isValid := object.CommitFilter(func(commit *object.Commit) bool {
		if commit.ID() == mb.ID() {
			return false
		} else if isAncestor, err := mb.IsAncestor(commit); err == nil {
			return isAncestor
		} else {
			return false
		}
	})
	isLimit := object.CommitFilter(func(c *object.Commit) bool {
		return !isValid(c)
	})

	iter := object.NewFilterCommitIter(hdCommit, &isValid, &isLimit)
	logs := make([]string, 0)
	_ = iter.ForEach(func(commit *object.Commit) error {
		logs = append(logs, strings.Split(commit.Message, "\n")[0])
		return nil
	})

	lastIdx := len(logs) - 1
	if lastIdx > 0 {
		return titleOpt.Default(logs[lastIdx]), descriptionOpt.Default(strings.Join(logs[0:lastIdx], "\n")), nil
	} else if lastIdx == 0 {
		return titleOpt.Default(logs[lastIdx]), "", nil
	} else {
		return "", "", fmt.Errorf("no commits from %v to %v, cannot infer pr title", baseBranch, headBranch)
	}

}

func (g *GitHubSv) CreatePullRequest(args CreatePullRequestArgs) (PullRequestStatus, error) {
	// First get the repository id
	if labelIds, err := toLabelIds(g, args.Labels); err != nil {
		return nil, err
	} else if reviewerIds, err := toUserIds(g, args.Reviewers); err != nil {
		return nil, err
	} else if resp, err := repositoryId(g.ctx, g.owner, g.repo); err != nil {
		return nil, err
	} else if headBranch, err := g.resolveHeadBranch(args.HeadBranch); err != nil {
		return nil, err
	} else if baseBranch, err := g.resolveBaseBranch(args.BaseBranch); err != nil {
		return nil, err
	} else if title, description, err := g.resolveTitleAndDescription(args.Title, args.Description, headBranch, baseBranch); err != nil {
		return nil, err
	} else if resp2, err := createPullRequest(g.ctx, resp.Repository.Id, headBranch, baseBranch, title, &description); err != nil {
		return nil, err
	} else {
		pr := resp2.CreatePullRequest.PullRequest.singleStatusPullRequest
		// If we have labels add them
		if len(labelIds) > 0 {
			if resp3, err := editPullRequest(g.ctx, pr.Id, labelIds); err != nil {
				pterm.Debug.Printfln("cannot add labels: %v", err)
			} else {
				pr = resp3.UpdatePullRequest.PullRequest.singleStatusPullRequest
			}
		}
		// If we have reviewers add them
		if len(reviewerIds) > 0 {
			if resp3, err := editPullRequestReviewers(g.ctx, pr.Id, reviewerIds); err != nil {
				pterm.Debug.Printfln("cannot add reviewers: %v", err)
			} else {
				pr = resp3.RequestReviews.PullRequest.singleStatusPullRequest
			}
		}

		return &GitHubPullRequestStatusWrapper{
			singleStatusPullRequest: &pr,
			isMine:                  true,
		}, nil
	}
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
	*singleStatusPullRequest
	isMine bool
}

func (g GitHubPullRequestStatusWrapper) IsMine() bool {
	return g.isMine
}

func (g GitHubPullRequestStatusWrapper) GetAuthor() string {
	return (*g.Author).GetLogin()
}

func (g GitHubPullRequestStatusWrapper) GetRepository() string {
	return fmt.Sprintf("%s/%s", g.Repository.Owner.GetLogin(), g.Repository.Name)
}

func (g GitHubPullRequestStatusWrapper) GetId() interface{} {
	return g.Number
}

func (g GitHubPullRequestStatusWrapper) GetTitle() string {
	return g.Title
}

func (g GitHubPullRequestStatusWrapper) GetStatus() string {
	return string(g.State)
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
		var author PullRequestsListRepositoryPullRequestReviewsPullRequestReviewConnectionNodesPullRequestReviewAuthorActor = &PullRequestsListRepositoryPullRequestReviewsPullRequestReviewConnectionNodesPullRequestReviewAuthorUser{
			Login: (*r.Author).GetDisplayName(),
		}
		rev = append(rev, GitHubReview{&PullRequestsListRepositoryPullRequestReviewsPullRequestReviewConnectionNodesPullRequestReview{
			Author:      &author,
			Body:        "",
			SubmittedAt: nil,
			State:       r.State,
		}})
	}

	return rev
}

func (g GitHubPullRequestStatusWrapper) GetChecksByStatus() map[string]int {
	states := make(map[string]int)
	if rollups := g.Commits.Nodes[0].Commit.StatusCheckRollup; rollups != nil {
		for _, s := range rollups.Contexts.CheckRunCountsByState {
			states[string(s.State)] = s.Count
		}
	}
	return states
}

func (g GitHubPullRequestStatusWrapper) GetContextByStatus() map[string]int {
	states := make(map[string]int)
	if rollups := g.Commits.Nodes[0].Commit.StatusCheckRollup; rollups != nil {
		for _, s := range rollups.Contexts.StatusContextCountsByState {
			states[string(s.State)] = s.Count
		}
	}
	return states
}

func (g *GitHubSv) GetCurrentBranch() (string, error) {
	if rep, err := git.PlainOpen(g.localRepo); err != nil {
		return "", err
	} else if hd, err := rep.Head(); err != nil {
		return "", err
	} else if hd.Name().IsBranch() {
		return hd.Name().Short(), nil
	} else {
		return "", fmt.Errorf("current reference '%v' is not a branch", hd.Name())
	}
}

func (g *GitHubSv) Fetch() error {
	rep, giterr := git.PlainOpen(g.localRepo)
	if giterr != nil {
		return giterr
	}
	if sshagent.Available() {

		if a, _, err := sshagent.New(); err != nil {
			return fmt.Errorf("error creating SSH agent: %w", err)
		} else if sigs, err := a.Signers(); err != nil {
			return fmt.Errorf("while getting signers : %w", err)
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
				return fmt.Errorf("couldn't find any suitable keys, won't try to fetch the repo '%s'", g.localRepo)
			}
		}
	} else {
		return errors.New("please use ssh agent")
	}
}

const githubDefaultHost = "github.com"
const githubDefaultGraphQLUrl = "https://api.github.com/graphql"

type GitHubExtendedHttpClient struct {
	*http.Client
}

func (g GitHubExtendedHttpClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("accept", "application/vnd.github.bane-preview+json")
	return g.Client.Do(req)
}

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

		cl2 := graphql.NewClient(githubDefaultGraphQLUrl, GitHubExtendedHttpClient{tc})
		ctx = gh_utils.InitContext(ctx, cl2)

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

func (g GitHubPullRequest) Merge() error {
	_, err := mergePullRequest(g.sv.ctx, g.GetNodeID())
	return err
}

func (g GitHubPullRequest) StartReview() (Review, error) {

	if _, err := newReview(g.sv.ctx, g.GetNodeID()); err != nil {
		return nil, err
	} else {
		return g.GetPendingReview()
	}
}

func (g GitHubPullRequest) GetPendingReview() (Review, error) {
	if login, err := g.sv.currentLogin(); err != nil {
		return nil, err
	} else if resp, err := currentPendingReview(g.sv.ctx, g.GetNodeID(), &login); err != nil {
		return nil, err
	} else if rev, ok := (*resp.Node).(*currentPendingReviewNodePullRequest); !ok {
		return nil, errors.New(fmt.Sprintf("bad response, we didn't get a pull request with node id '%s'", g.GetNodeID()))
	} else if l := len(rev.Reviews.Nodes); l != 1 {
		if l > 1 {
			return nil, errors.New(fmt.Sprintf("bad response, more than one pending review? %d > 1", l))
		}
		// No error but no reviews
		return nil, nil
	} else {
		return &GitHubReviewInfoWrapper{rev.Reviews.Nodes[0].ReviewInfo, g.sv}, nil
	}
}

func (g GitHubPullRequest) createOrGetPendingReview() (Review, bool, error) {
	if rev, err := g.GetPendingReview(); err != nil {
		return nil, false, err
	} else if rev != nil {
		return rev, false, nil
	} else if resp, err := newReview(g.sv.ctx, g.GetNodeID()); err != nil {
		return nil, false, err
	} else {
		return &GitHubReviewInfoWrapper{resp.AddPullRequestReview.PullRequestReview.ReviewInfo, g.sv}, true, nil
	}
}

type GitHubReviewInfoWrapper struct {
	ReviewInfo
	sv *GitHubSv
}

func (g *GitHubReviewInfoWrapper) Cancel() error {
	_, err := cancelReview(g.sv.ctx, g.Id)
	return err
}

func (g *GitHubReviewInfoWrapper) Dismiss() error {
	if _, err := closeReviewWithEvent(g.sv.ctx, g.Id, PullRequestReviewEventDismiss, nil); err != nil {
		return err
	} else {
		return nil
	}
}

func (g *GitHubReviewInfoWrapper) Close(comment *string) error {
	if _, err := closeReviewWithEvent(g.sv.ctx, g.Id, PullRequestReviewEventComment, comment); err != nil {
		return err
	} else {
		return nil
	}
}

func (g *GitHubReviewInfoWrapper) Approve(comment *string) error {
	if _, err := closeReviewWithEvent(g.sv.ctx, g.Id, PullRequestReviewEventApprove, comment); err != nil {
		return err
	} else {
		return nil
	}
}

func (g *GitHubReviewInfoWrapper) RequestChanges(comment *string) error {
	if _, err := closeReviewWithEvent(g.sv.ctx, g.Id, PullRequestReviewEventRequestChanges, comment); err != nil {
		return err
	} else {
		return nil
	}
}

func (g *GitHubReviewInfoWrapper) GetState() string {
	return string(g.State)
}

func (g *GitHubReviewInfoWrapper) GetAuthor() string {
	return (*g.Author).GetDisplayName()
}

func (g *GitHubReviewInfoWrapper) GetSubmitedAt() time.Time {
	return g.CreatedAt
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

func (g *GitHubSv) currentLogin() (string, error) {
	if resp, err := myLogin(g.ctx); err != nil {
		return "", err
	} else {
		return resp.Viewer.Login, nil
	}
}

type IssueResNode struct {
	Id         string
	Repository struct {
		NameWithOwner string
	}
	Number int
}

func (g *GitHubSv) searchIssueNodes(nodeQuery string) <-chan requestedReviewsSearchSearchResultItemConnectionEdgesSearchResultItemEdgeNodeSearchResultItem {

	res := make(chan requestedReviewsSearchSearchResultItemConnectionEdgesSearchResultItemEdgeNodeSearchResultItem)

	go func() {
		cont := true
		var after *string
		for cont {
			if rs, err := requestedReviews(g.ctx, nodeQuery, after); err == nil {
				for _, e := range rs.Search.Edges {
					res <- *e.Node
				}
				if rs.Search.PageInfo.HasNextPage {
					after = rs.Search.PageInfo.EndCursor
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

	ch := make(chan PullRequestStatus)
	go func() {

		ids := make([]string, 0)
		for nd := range g.searchIssueNodes(prQuery) {
			if nx, ok := nd.(*requestedReviewsSearchSearchResultItemConnectionEdgesSearchResultItemEdgeNodePullRequest); ok && nx != nil {
				ids = append(ids, nx.Id)
			}
			if len(ids) > 5 {
				if resp, err := singleStatus(g.ctx, ids); err == nil {
					for _, pr := range resp.Nodes {
						if pr != nil {
							if spr, ok := (*pr).(*singleStatusNodesPullRequest); ok {
								ch <- GitHubPullRequestStatusWrapper{&spr.singleStatusPullRequest, isMine}
							}
						}
					}
				}

				ids = make([]string, 0)
			}
		}

		if resp, err := singleStatus(g.ctx, ids); err == nil {
			for _, pr := range resp.Nodes {
				if pr != nil {
					if spr, ok := (*pr).(*singleStatusNodesPullRequest); ok {
						ch <- GitHubPullRequestStatusWrapper{&spr.singleStatusPullRequest, isMine}
					}
				}
			}
		}

		close(ch)
	}()

	return ch, nil
}

func (g GitHubPullRequest) ReplyToComment(comment Comment, replyText string) (Comment, error) {

	if c, ok := comment.(GitHubQLThreadCommentWrapper); ok {

		if rev, isNew, err := g.createOrGetPendingReview(); err != nil {
			return nil, err
		} else if _, err := replyTo(g.sv.ctx, rev.GetId(), c.comment.Id, replyText); err != nil {
			return nil, err
		} else if isNew {
			if _, err := closeReview(g.sv.ctx, rev.GetId()); err != nil {
				return nil, err
			}
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
		return nil, fmt.Errorf("illegal argument: not a github comment")
	}
}

func (g GitHubPullRequest) GetChecks() ([]Check, error) {

	resp1, err := GetChecksAndStatus(g.sv.ctx, g.sv.repo, g.sv.owner, g.GetNumber())

	if err != nil {
		pterm.Fatal.Println(err)
	}
	result := make([]Check, 0)

	for _, rollups := range resp1.Repository.PullRequest.StatusCheckRollup.Nodes {

		if rollups != nil && rollups.Commit.StatusCheckRollup != nil {
			for _, checks := range rollups.Commit.StatusCheckRollup.Contexts.Nodes {
				if cr, ok := (*checks).(*GetChecksAndStatusRepositoryPullRequestStatusCheckRollupPullRequestCommitConnectionNodesPullRequestCommitCommitStatusCheckRollupContextsStatusCheckRollupContextConnectionNodesCheckRun); ok {
					result = append(result, &GitHubCheckRun{cr.CheckRunCase})
				} else if sc, ok := (*checks).(*GetChecksAndStatusRepositoryPullRequestStatusCheckRollupPullRequestCommitConnectionNodesPullRequestCommitCommitStatusCheckRollupContextsStatusCheckRollupContextConnectionNodesStatusContext); ok {
					result = append(result, &GitHubCheck{sc.StatusContextCase})
				}
			}
		}

	}

	return result, nil
}

type GitHubCheck struct {
	StatusContextCase
}

type GitHubCheckRun struct {
	CheckRunCase
}

func (v *GitHubCheckRun) GetStatus() string {
	return string(v.Status)
}

func (v *GitHubCheckRun) GetUrl() string {
	return *v.DetailsUrl
}

func (g *GitHubCheck) GetUrl() string {
	return *g.TargetUrl

}

func (g GitHubCheck) GetName() string {
	return g.Context
}

func (g GitHubCheck) GetStatus() string {
	return string(g.State)
}

func (g GitHubPullRequest) GetReviews() ([]Review, error) {

	resp1, err := PullRequestsList(g.sv.ctx, g.sv.repo, g.sv.owner, g.GetNumber())
	if err != nil {
		pterm.Debug.Println(err)
		return nil, err
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
	*PullRequestsListRepositoryPullRequestReviewsPullRequestReviewConnectionNodesPullRequestReview
}

func (g GitHubReview) Cancel() error {
	return errors.New("cannot operate on other's reviews")
}

func (g GitHubReview) Dismiss() error {
	return errors.New("cannot operate on other's reviews")
}

func (g GitHubReview) Close(comment *string) error {
	return errors.New("cannot operate on other's reviews")
}

func (g GitHubReview) Approve(comment *string) error {
	return errors.New("cannot operate on other's reviews")
}

func (g GitHubReview) RequestChanges(comment *string) error {
	return errors.New("cannot operate on other's reviews")
}

func (g GitHubReview) GetId() string {
	return ""
}

func (g GitHubReview) GetState() string {
	return string(g.State)
}

func (g GitHubReview) GetAuthor() string {
	if aa := g.PullRequestsListRepositoryPullRequestReviewsPullRequestReviewConnectionNodesPullRequestReview.GetAuthor(); aa != nil {
		if a := *aa; a != nil {
			return a.GetLogin()
		} else {
			return ""
		}
	} else {
		return ""
	}
}

func (g GitHubReview) GetSubmitedAt() time.Time {
	if ok := g.SubmittedAt; ok != nil {
		return *ok
	} else {
		return time.Now()
	}
}

type ReactionInfo struct {
	Content   githubv4.ReactionContent
	CreatedAt time.Time
	User      UserInfo
}

func (r *ReactionsInfoReactionsReactionConnectionNodesReaction) GetAuthor() Author {
	return r.User
}

func (r *ReactionsInfoReactionsReactionConnectionNodesReaction) GetCreatedOn() time.Time {
	return r.CreatedAt
}

type ReactionsInfoOld struct {
	TotalCount int
	Nodes      []ReactionInfo
}

func (i *ReactionsInfoReactionsReactionConnection) toReactions() Reactions {
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
	Comment pullRequestCommentsRepositoryPullRequestCommentsIssueCommentConnectionNodesIssueComment
	error   error
}

type PullRequestThreadOrError struct {
	PullRequestThread pullRequestThreadsRepositoryPullRequestReviewThreadsPullRequestReviewThreadConnectionNodesPullRequestReviewThread
	error             error
}

func (g GitHubPullRequest) getPullRequestThreadComments() <-chan PullRequestThreadOrError {

	ch := make(chan PullRequestThreadOrError)

	go func() {
		cont := true

		var commentAfter *string
		for cont {

			if rs, err := pullRequestThreads(g.sv.ctx, *g.Number, g.sv.owner, g.sv.repo, commentAfter); err != nil {
				cont = false
				ch <- PullRequestThreadOrError{error: err}
			} else {
				for _, c := range rs.Repository.PullRequest.ReviewThreads.Nodes {
					ch <- PullRequestThreadOrError{PullRequestThread: *c}
				}

				commentAfter = rs.Repository.PullRequest.ReviewThreads.PageInfo.EndCursor
				cont = rs.Repository.PullRequest.ReviewThreads.PageInfo.HasNextPage
			}
		}
		close(ch)
	}()

	return ch
}

func (g GitHubPullRequest) getPullRequestComments() <-chan PullRequestCommentOrError {

	ch := make(chan PullRequestCommentOrError)

	go func() {
		cont := true
		var commentAfter *string
		for cont {

			if rs, err := pullRequestComments(g.sv.ctx, *g.Number, g.sv.owner, g.sv.repo, commentAfter); err != nil {
				cont = false
				ch <- PullRequestCommentOrError{error: err}
			} else {
				for _, c := range rs.Repository.PullRequest.Comments.Nodes {
					ch <- PullRequestCommentOrError{Comment: *c}
				}
				commentAfter = rs.Repository.PullRequest.Comments.PageInfo.EndCursor
				cont = rs.Repository.PullRequest.Comments.PageInfo.HasNextPage
			}
		}
		close(ch)
	}()

	return ch

}

type GitHubQLThreadCommentWrapper struct {
	thread  pullRequestThreadsRepositoryPullRequestReviewThreadsPullRequestReviewThreadConnectionNodesPullRequestReviewThread
	comment pullRequestThreadsRepositoryPullRequestReviewThreadsPullRequestReviewThreadConnectionNodesPullRequestReviewThreadCommentsPullRequestReviewCommentConnectionNodesPullRequestReviewComment
}

func (g GitHubQLThreadCommentWrapper) GetReactions() Reactions {
	return g.comment.Reactions.toReactions()
}

func (g GitHubQLThreadCommentWrapper) GetContent() CommentContent {
	return &g.comment
}

func (g GitHubQLThreadCommentWrapper) GetParentId() interface{} {
	if g.comment.ReplyTo != nil {
		return g.comment.ReplyTo.Id
	} else {
		return nil
	}
}

func (g GitHubQLThreadCommentWrapper) GetId() interface{} {
	return g.comment.GetId()
}

func (g GitHubQLThreadCommentWrapper) GetUser() Author {
	if au := g.comment.GetAuthor(); au != nil {
		return *au
	}

	return nil
}

func (g GitHubQLThreadCommentWrapper) GetCreatedOn() time.Time {
	return g.comment.GetCreatedAt()
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
			cmt := GitHubQLThreadCommentWrapper{th, *c}

			path := th.Path
			byLine, ok := commentMap[path]
			if !ok {
				byLine = make(map[int64][]Comment)
				commentMap[path] = byLine
			}

			var line int64

			if th.Line != nil {
				if th.DiffSide == DiffSideRight {
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
		prComments = append(prComments, GithubQLCommentWrapper{&re.Comment})
	}

	return prComments, commentMap, nil
}

type GithubQLCommentWrapper struct {
	CommentInfo
}

func (g GithubQLCommentWrapper) GetId() interface{} {
	return g.CommentInfo.GetId()
}

func (g GithubQLCommentWrapper) GetReactions() Reactions {
	if ci, ok := g.CommentInfo.(*pullRequestCommentsRepositoryPullRequestCommentsIssueCommentConnectionNodesIssueComment); ok {
		return ci.Reactions.toReactions()
	} else {
		return nil
	}
}

func (g GithubQLCommentWrapper) GetRaw() string {
	return g.CommentInfo.GetRaw()
}

func (g GithubQLCommentWrapper) GetContent() CommentContent {
	return g
}

func (g GithubQLCommentWrapper) GetParentId() interface{} {
	return nil
}

func (g GithubQLCommentWrapper) GetUser() Author {
	if au := g.CommentInfo.GetAuthor(); au != nil {
		return *au
	} else {
		return nil
	}
}

func (g GithubQLCommentWrapper) GetCreatedOn() time.Time {
	return g.GetCreatedAt()
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

type MissingCommitError struct {
	missingHash plumbing.Hash
	cause       error
}

func (m *MissingCommitError) Error() string {
	return fmt.Sprintf("Cannot find commit hash: %s. Cause: %v", m.missingHash, m.cause.Error())
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
		pterm.Debug.Println("Cannot get the pr branch head commit, do you need to update your local repo ?", err)
		return nil, &MissingCommitError{refBrHash, err}
	}
	cBase, err := rep.CommitObject(refBaseHash)
	if err != nil {
		pterm.Debug.Println("Cannot get the base branch commit, do you need to update your local repo ?", err)
		return nil, &MissingCommitError{refBaseHash, err}
	}

	merge, err := cBr.MergeBase(cBase)
	if err != nil {
		pterm.Debug.Println("Cannot find a merge base?!?")
		return nil, err
	}
	if len(merge) != 1 {
		pterm.Debug.Printfln("More than one merge base ?!? %s", merge)
		return nil, errors.New(pterm.Sprintfln("More than one merge base : %d", len(merge)))
	}

	baseTree, err2 := merge[0].Tree()
	if err2 != nil {
		pterm.Debug.Println(err2)
		return nil, err2
	}
	brTree, err3 := cBr.Tree()
	if err3 != nil {
		pterm.Debug.Println(err3)
		return nil, err3
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
