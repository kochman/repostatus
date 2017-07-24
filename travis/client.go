package travis

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"sort"
	"time"
)

type Client struct {
	Org               string
	Repo              string
	GitHubAccessToken string
}

func (t *Client) Branches() ([]Branch, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: t.GitHubAccessToken})
	tc := oauth2.NewClient(context.Background(), ts)
	ghc := github.NewClient(tc)
	ghb, _, err := ghc.Repositories.ListBranches(context.Background(), t.Org, t.Repo, nil)
	if err != nil {
		return nil, err
	}

	branches := make([]Branch, len(ghb))

	for i, branch := range ghb {
		cs, _, err := ghc.Repositories.GetCombinedStatus(context.Background(), t.Org, t.Repo, *branch.Name, nil)
		if err != nil {
			return nil, err
		}

		// determine most recent status change
		var mostRecent time.Time
		for _, status := range cs.Statuses {
			if status.UpdatedAt.After(mostRecent) {
				mostRecent = *status.UpdatedAt
			}
		}

		commitsURL := "https://github.com/" + t.Org + "/" + t.Repo + "/commits/" + *branch.Name
		branch := Branch{
			Name:        *branch.Name,
			State:       *cs.State, // failure, pending, or success (maybe error?)
			LastUpdated: mostRecent,
			CommitsURL:  commitsURL,
		}

		branches[i] = branch
	}

	sort.Sort(ByTime(branches))

	return branches, nil
}

type ByTime []Branch

func (b ByTime) Len() int {
	return len(b)
}

func (b ByTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByTime) Less(i, j int) bool {
	return b[i].LastUpdated.After(b[j].LastUpdated)
}
