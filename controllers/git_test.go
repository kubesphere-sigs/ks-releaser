package controllers

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
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
