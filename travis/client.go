package travis

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"sort"
	"strings"
	"time"
)

type Client struct {
	RepoSlug          string
	GitHubAccessToken string
}

func (t *Client) Org() string {
	s := strings.Split(t.RepoSlug, "/")
	if len(s) != 2 {
		return ""
	}
	return s[0]
}

func (t *Client) Repo() string {
	s := strings.Split(t.RepoSlug, "/")
	if len(s) != 2 {
		return ""
	}
	return s[1]
}

func (t *Client) Branches() ([]Branch, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: t.GitHubAccessToken})
	tc := oauth2.NewClient(context.Background(), ts)
	ghc := github.NewClient(tc)
	ghb, _, err := ghc.Repositories.ListBranches(context.Background(), t.Org(), t.Repo(), nil)
	if err != nil {
		return nil, err
	}

	branches := make([]Branch, len(ghb))

	for i, branch := range ghb {
		cs, _, err := ghc.Repositories.GetCombinedStatus(context.Background(), t.Org(), t.Repo(), *branch.Name, nil)
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

		branch := Branch{
			Name:        *branch.Name,
			State:       *cs.State, // failure, pending, or success (maybe error?)
			LastUpdated: mostRecent,
		}

		branches[i] = branch
		log.Println(branch)
	}

	sort.Sort(ByTime(branches))

	log.Println(branches)
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
