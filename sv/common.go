package sv

import (
	"github.com/bluekeyes/go-gitdiff/gitdiff"
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
}

type Comment interface {
	GetContent() CommentContent
	GetParentId() interface{}
	GetId() interface{}
	GetUser() Author
	GetCreatedOn() time.Time
}

type Check interface {
	GetName() string
	GetStatus() string
	GetUrl() string
}

type Review interface {
	GetState() string
	GetAuthor() string
	GetSubmitedAt() time.Time
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
	Fetch() error
}
