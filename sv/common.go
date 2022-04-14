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
}

type Comment interface {
	GetContent() CommentContent
	GetParentId() interface{}
	GetId() interface{}
	GetUser() Author
	GetCreatedOn() time.Time
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
}
