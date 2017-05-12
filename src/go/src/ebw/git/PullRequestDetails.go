package git

import (
	"github.com/google/go-github/github"
)

type PullRequestDetails struct {
	ChangedFiles []string
}
func PullRequestChanges(client *Client, pr *github.PullRequest, repoName string) (PullRequestDetails, error) {
	var cf [] string
	prc := PullRequestDetails{}

	listOptions := &github.ListOptions{
		Page: 0, PerPage: 1000,
	}
	c, _, err := client.PullRequests.ListFiles(client.Context, client.Username,
		repoName, pr.GetNumber(), listOptions)
	if nil != err {
		return prc, err
	}

	for _, f := range c {
		cf = append(cf, f.GetFilename())
	}
	prc.ChangedFiles = cf
	return prc, nil
}
