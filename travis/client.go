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

	// loop over all pages in response
	ghBranches := []*github.Branch{}
	opt := &github.ListOptions{}
	for {
		ghb, resp, err := ghc.Repositories.ListBranches(context.Background(), t.Org, t.Repo, opt)
		if err != nil {
			return nil, err
		}
		ghBranches = append(ghBranches, ghb...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	branches := make([]Branch, len(ghBranches))
	for i, branch := range ghBranches {
		cs, _, err := ghc.Repositories.GetCombinedStatus(context.Background(), t.Org, t.Repo, *branch.Name, nil)
		if err != nil {
			return nil, err
		}

		// determine most recent status change
		var mostRecent time.Time
		statusChecks := []StatusCheck{}
		for _, status := range cs.Statuses {
			if status.UpdatedAt.After(mostRecent) {
				mostRecent = *status.UpdatedAt
			}

			statusCheck := StatusCheck{
				State:       *status.State,
				Description: *status.Description,
				StatusURL:   status.GetTargetURL(),
			}
			statusChecks = append(statusChecks, statusCheck)
		}

		// if there aren't any status checks, get the time of the most recent commit
		if len(statusChecks) == 0 {
			ghCommit, _, err := ghc.Repositories.GetCommit(context.Background(), t.Org, t.Repo, *branch.Commit.SHA)
			if err != nil {
				return nil, err
			}
			mostRecent = *ghCommit.Commit.Author.Date
		}

		commitsURL := "https://github.com/" + t.Org + "/" + t.Repo + "/commits/" + *branch.Name
		commitURL := "https://github.com/" + t.Org + "/" + t.Repo + "/commit/" + *branch.Commit.SHA
		branch := Branch{
			Name:         *branch.Name,
			State:        *cs.State, // failure, pending, or success (maybe error?)
			LastUpdated:  mostRecent,
			CommitsURL:   commitsURL,
			CommitURL:    commitURL,
			StatusChecks: statusChecks,
			SHA:          *branch.Commit.SHA,
		}

		branches[i] = branch
	}

	sort.Sort(ByTime(branches))

	return branches, nil
}

func (t *Client) Repository() (Repo, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: t.GitHubAccessToken})
	tc := oauth2.NewClient(context.Background(), ts)
	ghc := github.NewClient(tc)
	ghr, _, err := ghc.Repositories.Get(context.Background(), t.Org, t.Repo)
	if err != nil {
		return Repo{}, err
	}

	repo := Repo{
		Name:        ghr.GetName(),
		Description: ghr.GetDescription(),
		Website:     ghr.GetHomepage(),
		GitHubURL:   ghr.GetHTMLURL(),
	}

	branches, err := t.Branches()
	if err != nil {
		return repo, err
	}
	repo.Branches = branches

	return repo, nil
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
