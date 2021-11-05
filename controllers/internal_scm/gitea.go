package internal_scm

import (
	"fmt"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/gitea"
)

type Gitea struct {
	server string
	repo   string
	token  string

	client *scm.Client
}

func NewGitea(server, repo, token string) *Gitea {
	return &Gitea{
		server: server,
		repo:   repo,
		token:  token,
	}
}

func (r *Gitea) Release(version, commitish string, draft, prerelease bool) (err error) {
	var client *scm.Client
	if client, err = gitea.NewWithToken(r.server, r.token); err != nil || client == nil {
		err = fmt.Errorf("failed to create gitea client, error: %v", err)
	} else {
		err = release(client, r.repo, version, commitish, draft, prerelease)
	}
	return
}
