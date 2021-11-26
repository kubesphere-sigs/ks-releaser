package internal_scm

import (
	"context"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
)

type GitReleaser interface {
	Release(version, commitish string, draft, prerelease bool) (err error)
	CreateIssue(title, body string) (err error)
}

func release(client *scm.Client, repo, version, commitish string, draft, prerelease bool) (err error) {
	releaseInput := &scm.ReleaseInput{
		Title:      version,
		Tag:        version,
		Commitish:  commitish,
		Draft:      draft,
		Prerelease: prerelease,
	}

	// just publish the draft release if it is existing
	release := findRelease(client, repo, version)
	if release == nil {
		_, _, err = client.Releases.Create(context.TODO(), repo, releaseInput)
	} else if release.Draft {
		releaseInput.Description = release.Description
		releaseInput.Title = release.Title
		_, _, err = client.Releases.Update(context.TODO(), repo, release.ID, releaseInput)
	}
	// ignore the existing release
	return
}

func findRelease(client *scm.Client, repo, version string) (release *scm.Release) {
	var err error
	cxt := context.TODO()
	if release, _, err = client.Releases.FindByTag(cxt, repo, version); err == scm.ErrNotFound {
		release = nil

		var list []*scm.Release
		if list, _, err = client.Releases.List(cxt, repo, scm.ReleaseListOptions{Page: 1, Size: 200}); err != nil {
			release = nil
		} else {
			for i, _ := range list {
				if list[i].Tag == version {
					release = list[i]
					return
				}
			}
		}
	} else if err != nil {
		release = nil
	}
	return
}

// GetGitProvider returns the GitReleaser implement by kind
func GetGitProvider(kind, server, repo, token string) GitReleaser {
	switch v1alpha1.Provider(kind) {
	case v1alpha1.ProviderGitHub:
		return NewGitHub(repo, token)
	case v1alpha1.ProviderGitlab:
		return NewGitlab(repo, token)
	case v1alpha1.ProviderGitea:
		return NewGitea(server, repo, token)
	}
	return nil
}
