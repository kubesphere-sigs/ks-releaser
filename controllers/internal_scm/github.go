package internal_scm

import (
	"context"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/github"
	"github.com/jenkins-x/go-scm/scm/transport"
	"net/http"
)

type GitHub struct {
	repo  string
	token string
}

// NewGitHub creates a new instance
func NewGitHub(repo, token string) *GitHub {
	return &GitHub{
		repo:  repo,
		token: token,
	}
}

func (r *GitHub) Release(version, commitish string, draft, prerelease bool) (err error) {
	client := r.getClient()

	releaseInput := &scm.ReleaseInput{
		Title:      version,
		Tag:        version,
		Commitish:  commitish,
		Draft:      draft,
		Prerelease: prerelease,
	}

	// just publish the draft release if it is existing
	release := r.findDraftRelease(version)
	if release != nil {
		releaseInput.Description = release.Description
		releaseInput.Title = release.Title
		_, _, err = client.Releases.Update(context.TODO(), r.repo, release.ID, releaseInput)
	} else {
		_, _, err = client.Releases.Create(context.TODO(), r.repo, releaseInput)
	}
	return
}

func (r *GitHub) getClient() (client *scm.Client) {
	client = github.NewDefault()
	client.Client = &http.Client{
		Transport: &transport.BearerToken{
			Token: r.token,
		},
	}
	return
}

func (r *GitHub) findDraftRelease(version string) *scm.Release {
	client := r.getClient()

	if releaseList, _, err := client.Releases.List(context.TODO(), r.repo, scm.ReleaseListOptions{
		Page: 1,
		Size: 50,
	}); err == nil {
		for i, _ := range releaseList {
			release := releaseList[i]
			if release.Draft && release.Tag == version {
				return release
			}
		}
	}
	return nil
}
