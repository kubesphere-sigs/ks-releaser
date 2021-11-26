package internal_scm

import (
	"context"
	"fmt"
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
	client := github.NewDefault()
	client.Client = &http.Client{
		Transport: &transport.BearerToken{
			Token: r.token,
		},
	}
	err = release(client, r.repo, version, commitish, draft, prerelease)
	return
}

func (r *GitHub) CreateIssue(title, body string) (err error) {
	ctx := context.TODO()
	client := github.NewDefault()
	client.Client = &http.Client{
		Transport: &transport.BearerToken{
			Token: r.token,
		},
	}

	var list []*scm.SearchIssue
	if list, _, err = client.Issues.Search(ctx, scm.SearchOptions{
		Query: fmt.Sprintf("%s in:title repo:%s", title, r.repo),
	}); err == nil && len(list) > 0 {
		issue := list[0]

		_, _, err = client.Issues.CreateComment(ctx, r.repo, issue.Number, &scm.CommentInput{
			Body: body,
		})
	} else {
		_, _, err = client.Issues.Create(ctx, r.repo, &scm.IssueInput{
			Title: title,
			Body:  body,
		})
	}
	return
}
