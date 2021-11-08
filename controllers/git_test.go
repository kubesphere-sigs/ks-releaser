package controllers

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestGetAuth(t *testing.T) {
	var secret *v1.Secret
	var auth transport.AuthMethod

	// secret is nil
	auth = getAuth(secret)
	assert.Nil(t, auth)

	// empty secret
	secret = &v1.Secret{}
	auth = getAuth(secret)
	assert.Nil(t, auth)

	// basic auth
	secret = &v1.Secret{
		Type: v1.SecretTypeBasicAuth,
		Data: map[string][]byte{
			v1.BasicAuthUsernameKey: []byte("username"),
			v1.BasicAuthPasswordKey: []byte("password"),
		},
	}
	auth = getAuth(secret)
	assert.NotNil(t, auth)
	assert.Contains(t, auth.String(), "username")
	assert.Equal(t, "http-basic-auth", auth.Name())

	// ssh auth
	secret = &v1.Secret{
		Type: v1.SecretTypeSSHAuth,
	}
	auth = getAuth(secret)
	assert.NotNil(t, auth)
	assert.Equal(t, ssh.PublicKeysName, auth.Name())
}

func Test_getAction(t *testing.T) {
	type args struct {
		repo v1alpha1.Repository
	}
	tests := []struct {
		name       string
		args       args
		wantAction v1alpha1.Action
	}{{
		name: "auto action",
		args: args{
			repo: v1alpha1.Repository{
				Action:  v1alpha1.ActionAuto,
				Version: "v1.2.3-alpha.1",
			},
		},
		wantAction: v1alpha1.ActionPreRelease,
	}, {
		name: "release action",
		args: args{
			repo: v1alpha1.Repository{
				Action:  v1alpha1.ActionRelease,
				Version: "v1.2.3-alpha.1",
			},
		},
		wantAction: v1alpha1.ActionRelease,
	}, {
		name: "tag action",
		args: args{
			repo: v1alpha1.Repository{
				Action:  v1alpha1.ActionTag,
				Version: "v1.2.3-alpha.1",
			},
		},
		wantAction: v1alpha1.ActionTag,
	}, {
		name: "pre-release action",
		args: args{
			repo: v1alpha1.Repository{
				Action:  v1alpha1.ActionPreRelease,
				Version: "v1.2.3-alpha.1",
			},
		},
		wantAction: v1alpha1.ActionPreRelease,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAction := getAction(tt.args.repo); gotAction != tt.wantAction {
				t.Errorf("getAction() = %v, want %v", gotAction, tt.wantAction)
			}
		})
	}
}

func Test_getOrgAndRepo(t *testing.T) {
	type args struct {
		repo   v1alpha1.Repository
		server string
	}
	tests := []struct {
		name           string
		args           args
		wantOrgAndRepo string
	}{{
		name: "github with http protocol",
		args: args{
			repo: v1alpha1.Repository{
				Address: "https://github.com/x/b",
			},
		},
		wantOrgAndRepo: "x/b",
	}, {
		name: "github with git protocol",
		args: args{
			repo: v1alpha1.Repository{
				Address: "https://github.com/x/b.git",
			},
		},
		wantOrgAndRepo: "x/b",
	}, {
		name: "gitlab with git protocol",
		args: args{
			repo: v1alpha1.Repository{
				Address: "https://gitlab.com/x/b.git",
			},
		},
		wantOrgAndRepo: "x/b",
	}, {
		name: "gitee with git protocol",
		args: args{
			repo: v1alpha1.Repository{
				Address: "https://gitee.com/x/b.git",
			},
		},
		wantOrgAndRepo: "x/b",
	}, {
		name: "bitbucket with git protocol",
		args: args{
			repo: v1alpha1.Repository{
				Address: "https://bitbucket.org/x/b.git",
			},
		},
		wantOrgAndRepo: "x/b",
	}, {
		name: "gitea with git protocol",
		args: args{
			repo: v1alpha1.Repository{
				Provider: v1alpha1.ProviderGitea,
				Address:  "https://localhost/x/b.git",
			},
			server: "https://localhost/",
		},
		wantOrgAndRepo: "x/b",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOrgAndRepo := getOrgAndRepo(tt.args.repo, tt.args.server); gotOrgAndRepo != tt.wantOrgAndRepo {
				t.Errorf("getOrgAndRepo() = %v, want %v", gotOrgAndRepo, tt.wantOrgAndRepo)
			}
		})
	}
}
