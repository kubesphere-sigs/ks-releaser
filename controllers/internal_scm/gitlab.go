package internal_scm

import (
	"github.com/jenkins-x/go-scm/scm/driver/gitlab"
	"github.com/jenkins-x/go-scm/scm/transport"
	"net/http"
)

type Gitlab struct {
	repo  string
	token string
}

// NewGitlab creates a new instance
func NewGitlab(repo, token string) *Gitlab {
	return &Gitlab{
		repo:  repo,
		token: token,
	}
}

func (r *Gitlab) Release(version, commitish string, draft, prerelease bool) (err error) {
	client := gitlab.NewDefault()
	client.Client = &http.Client{
		Transport: &transport.BearerToken{
			Token: r.token,
		},
	}
	err = release(client, r.repo, version, commitish, draft, prerelease)
	return
}
