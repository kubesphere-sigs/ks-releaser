package controllers

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	devopsv1alpha1 "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
	"github.com/kubesphere-sigs/ks-releaser/controllers/internal_scm"
	"golang.org/x/crypto/ssh"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

/**
TODO make these functions into a struct
For example, we can share parts of the variables, such as git.Repository, secret .etc.
 */

func saveAndPush(gitRepo *git.Repository, user, targetFile string, data []byte, secret *v1.Secret) (err error) {
	if err = ioutil.WriteFile(targetFile, data, 0644); err != nil {
		fmt.Println("failed to write file", targetFile)
	} else {
		if err = addAndCommit(gitRepo, user); err == nil {
			err = pushTags(gitRepo, "", getAuth(secret))
		}
	}
	return
}

func addAndCommit(repo *git.Repository, user string) (err error) {
	var w *git.Worktree
	if w, err = repo.Worktree(); err == nil {
		_, _ = w.Add(".")
		var commit plumbing.Hash
		commit, err = w.Commit("example go-git commit", &git.CommitOptions{
			All: true,
			Author: &object.Signature{
				Name:  user,
				Email: fmt.Sprintf("%s@users.noreply.github.com", user),
				When:  time.Now(),
			},
		})

		if err == nil {
			_, err = repo.CommitObject(commit)
		}
	}
	return
}

func release(repo devopsv1alpha1.Repository, secret *v1.Secret, user string) (err error) {
	auth := getAuth(secret)

	var gitRepo *git.Repository
	if gitRepo, err = clone(repo.Address, repo.Branch, auth, "tmp"); err != nil {
		return
	}

	if repo.Message == "" {
		repo.Message = "released by ks-releaser"
	}
	if _, err = setTag(gitRepo, repo.Version, repo.Message, user); err != nil {
		return
	}

	if err = pushTags(gitRepo, repo.Version, auth); err != nil {
		return
	}

	var orgAndRepo string
	switch repo.Provider {
	case devopsv1alpha1.ProviderGitHub:
		orgAndRepo = strings.ReplaceAll(repo.Address, "https://github.com/", "")
	}

	switch repo.Action {
	case devopsv1alpha1.ActionPreRelease:
		provider := internal_scm.GetGitProvider(string(repo.Provider), orgAndRepo, string(secret.Data[v1.BasicAuthPasswordKey]))
		if provider != nil {
			err = provider.Release(repo.Version, false, true)
		}
	case devopsv1alpha1.ActionRelease:
		provider := internal_scm.GetGitProvider(string(repo.Provider), orgAndRepo, string(secret.Data[v1.BasicAuthPasswordKey]))
		if provider != nil {
			err = provider.Release(repo.Version, false, false)
		}
	}
	return
}

func getAuth(secret *v1.Secret) (auth transport.AuthMethod) {
	switch secret.Type {
	case v1.SecretTypeBasicAuth:
		auth = &githttp.BasicAuth{
			Username: string(secret.Data[v1.BasicAuthUsernameKey]),
			Password: string(secret.Data[v1.BasicAuthPasswordKey]),
		}
	case v1.SecretTypeSSHAuth:
		signer, _ := ssh.ParsePrivateKey(secret.Data[v1.SSHAuthPrivateKey])
		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}
	return
}

func clone(gitRepo, branch string, auth transport.AuthMethod, cacheDir string) (repo *git.Repository, err error) {
	var gitRepoURL *url.URL
	if gitRepoURL, err = url.Parse(gitRepo); err != nil {
		return
	}

	dir := path.Join(cacheDir, gitRepoURL.Path)
	if ok, _ := PathExists(dir); ok {
		if repo, err = git.PlainOpen(dir); err == nil {
			var wd *git.Worktree

			if wd, err = repo.Worktree(); err == nil {
				if err = wd.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName(branch),
					Create: false,
					Force:  true,
				}); err != nil {
					err = fmt.Errorf("unable to checkout git branch: %s", branch)
					return
				}

				if err = wd.Pull(&git.PullOptions{
					Progress:      os.Stdout,
					ReferenceName: plumbing.NewBranchReferenceName(branch),
					Force:         true, // in case of the force pushing
					Auth:          auth,
				}); err != nil && err != git.NoErrAlreadyUpToDate {
					err = fmt.Errorf("failed to pull git repository '%s', error: %v", repo, err)
				} else {
					err = nil
				}
			}
		} else {
			err = fmt.Errorf("failed to open git local repository, error: %v", err)
		}
	} else {
		repo, err = git.PlainClone(dir, false, &git.CloneOptions{
			URL:           gitRepo,
			ReferenceName: plumbing.NewBranchReferenceName(branch),
			Progress:      os.Stdout,
			Auth:          auth,
		})
	}
	return
}

func tagExists(tag string, r *git.Repository) bool {
	var err error
	var tags storer.ReferenceIter
	if tags, err = r.Tags(); err == nil {
		err = tags.ForEach(func(reference *plumbing.Reference) error {
			tagRef := reference.Name()
			if tagRef.IsTag() && !tagRef.IsRemote() {
				_ = r.DeleteTag(tag)
			} else if tagRef.IsTag() && tagRef.IsRemote() && tagRef.String() == tag {
				return nil
			}
			return errors.New("not found tag")
		})
	}
	return err == nil || (err != nil && err.Error() != "not found tag")
}

func setTag(r *git.Repository, tag, message, user string) (bool, error) {
	if tagExists(tag, r) {
		fmt.Printf("tag %s already exists\n", tag)
		return false, nil
	}
	fmt.Printf("Set tag %s\n", tag)
	h, err := r.Head()
	if err != nil {
		fmt.Printf("get HEAD error: %s\n", err)
		return false, err
	}
	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  user,
			Email: fmt.Sprintf("%s@users.noreply.github.com", user),
			When:  time.Time{},
		},
		Message: message,
	})

	if err != nil {
		fmt.Printf("create tag error: %s\n", err)
		return false, err
	}
	return true, nil
}

func pushTags(r *git.Repository, tag string, auth transport.AuthMethod) (err error) {
	var ref []config.RefSpec
	if tag != "" {
		ref = []config.RefSpec{config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tag, tag))}
	}

	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   ref,
		Auth:       auth,
	}
	if err = r.Push(po); err != nil {
		if err == git.NoErrAlreadyUpToDate {
			fmt.Print("origin remote was up to date, no push done\n")
			err = nil
			return
		}
		err = fmt.Errorf("push to remote origin error: %s\n", err)
	}
	return
}

// PathExists checks if the target path exist or not
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
