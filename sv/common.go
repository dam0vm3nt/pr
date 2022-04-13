package sv

import "time"

type PullRequest interface {
	GetBranch() Branch
	GetId() interface{}
	GetTitle() string
	GetAuthor() Author
	GetState() string
	GetCreatedOn() time.Time
}

type Author interface {
	GetDisplayName() string
}

type Branch interface {
	GetName() string
}

type Sv interface {
	ListPullRequests(query string) (<-chan PullRequest, error)
	getPullRequest(id string) PullRequest
}
