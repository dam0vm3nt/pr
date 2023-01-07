package sv

import (
	"github.com/antihax/optional"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"os/exec"
	"time"
)

type PullRequest interface {
	GetBranch() Branch
	GetId() interface{}
	GetTitle() string
	GetAuthor() Author
	GetState() string
	GetCreatedOn() time.Time
	GetCommentsByLine() ([]Comment, map[string]map[int64][]Comment, error)
	GetDiff() ([]*gitdiff.File, error)
	GetBase() Branch
	GetChecks() ([]Check, error)
	GetReviews() ([]Review, error)
	ReplyToComment(comment Comment, replyText string) (Comment, error)
	CreateComment(path string, commitId string, line int, isNew bool, body string) (Comment, error)
	GetLastCommitId() string
	GetPendingReview() (Review, error)
	StartReview() (Review, error)
	Merge() error
}

type Comment interface {
	GetContent() CommentContent
	GetParentId() interface{}
	GetId() interface{}
	GetUser() Author
	GetCreatedOn() time.Time
	GetReactions() Reactions
}

type Reactions map[string][]Reaction

type Reaction interface {
	GetAuthor() Author
	GetCreatedOn() time.Time
}

type Check interface {
	GetName() string
	GetStatus() string
	GetUrl() string
}

type Review interface {
	GetId() string
	GetState() string
	GetAuthor() string
	GetSubmitedAt() time.Time
	Dismiss() error
	Close(comment *string) error
	Approve(comment *string) error
	RequestChanges(comment *string) error
	Cancel() error
}

type CommentContent interface {
	GetRaw() string
}

type Author interface {
	GetDisplayName() string
}

type Branch interface {
	GetName() string
}

type Sv interface {
	ListPullRequests(query string) (<-chan PullRequest, error)
	GetPullRequest(id string) (PullRequest, error)
	PullRequestStatus() (<-chan PullRequestStatus, error)
	Fetch() error
	GetRepositoryFullName() string
	CreatePullRequest(args CreatePullRequestArgs) (PullRequestStatus, error)
	GetCurrentBranch() (string, error)
}

type CreatePullRequestArgs struct {
	BaseBranch          optional.String
	HeadBranch          optional.String
	Title               optional.String
	Description         optional.String
	Labels              []string
	Reviewers           []string
	CreateMissingLabels bool
}

type PullRequestStatus interface {
	GetId() interface{}
	GetTitle() string
	GetStatus() string
	GetBranchName() string
	GetBaseName() string
	GetReviews() []Review
	GetChecksByStatus() map[string]int
	GetContextByStatus() map[string]int
	GetAuthor() string
	GetRepository() string
	IsMine() bool
}

func execGitFetch() error {
	if path, err := exec.LookPath("git"); err != nil {
		return err
	} else {
		cmd := exec.Command(path, "fetch")
		// cmd.Stdin = os.Stdin
		// cmd.Stdout = os.Stdout
		if err = cmd.Start(); err != nil {
			return err
		}
		if err = cmd.Wait(); err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				if err.ExitCode() != 1 {
					return err
				} else {
					// We can ignore exit code 1 from fetch.
					return nil
				}
			}
			return err
		}

		return nil
	}

}

func ForceFetch(repo Sv) error {
	if err := repo.Fetch(); err == nil {
		return nil
	} else if err := execGitFetch(); err == nil {
		return nil
	} else {
		return err
	}
}
