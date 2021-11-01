package internal_scm

import "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"

type GitReleaser interface {
	Release(version string, draft, prerelease bool) (err error)
}

func GetGitProvider(kind, repo, token string) GitReleaser {
	switch v1alpha1.Provider(kind) {
	case v1alpha1.ProviderGitHub:
		return NewGitHub(repo, token)
	}
	return nil
}
