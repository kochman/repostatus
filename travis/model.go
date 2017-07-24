package travis

import "time"

type BuildsResp struct {
	Builds  []Build
	Commits []TravisCommit
}

type Build struct {
	Id                int
	Number            string `json:"number"`
	EventType         string
	PullRequest       bool   `json:"pull_request"`
	PullRequestTitle  string `json:"pull_request_title"`
	PullRequestNumber int
	State             string `json:"state"`
	Duration          int
	StartedAt         time.Time `json:"started_at"`
	FinishedAt        time.Time `json:"finished_at"`
}

type TravisCommit struct {
	Id     int
	Sha    string
	Branch string
}

type TravisBranchResp struct {
	Branch TravisBranch
	Commit TravisCommit
}

type TravisBranch struct {
	Id                int
	Number            string
	EventType         string
	PullRequest       bool
	PullRequestTitle  string
	PullRequestNumber int
	State             string `json:"state"`
	Duration          int
	StartedAt         time.Time `json:"started_at"`
	FinishedAt        time.Time `json:"finished_at"`
}

type Repo struct {
}

type Branch struct {
	Name        string    `json:"name"`
	State       string    `json:"state"`
	LastUpdated time.Time `json:"last_updated"`
	CommitsURL  string    `json:"commits_url"`
}
