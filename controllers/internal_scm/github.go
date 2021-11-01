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

func (r *GitHub) Release(version string, draft, prerelease bool) (err error) {
	client := github.NewDefault()
	client.Client = &http.Client{
		Transport: &transport.BearerToken{
			Token: r.token,
		},
	}
	_, _, err = client.Releases.Create(context.TODO(), r.repo, &scm.ReleaseInput{
		Title:      version,
		Tag:        version,
		Draft:      draft,
		Prerelease: prerelease,
	})
	return
}
