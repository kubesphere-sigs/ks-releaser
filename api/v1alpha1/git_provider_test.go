package v1alpha1

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGitProvider(t *testing.T) {
	assert.Equal(t, ProviderUnknown, GetDefaultProvider(&Repository{Address: "https://baidu.com/x/b"}))
	assert.Equal(t, ProviderGitHub, GetDefaultProvider(&Repository{Address: "https://github.com/x/b"}))
	assert.Equal(t, ProviderGitlab, GetDefaultProvider(&Repository{Address: "https://gitlab.com/x/b"}))
	assert.Equal(t, ProviderGitea, GetDefaultProvider(&Repository{Address: "https://gitea.com/x/b"}))
	assert.Equal(t, ProviderGitee, GetDefaultProvider(&Repository{Address: "https://gitee.com/x/b"}))
	assert.Equal(t, ProviderBitbucket, GetDefaultProvider(&Repository{Address: "https://bitbucket.org/x/b"}))
	assert.Equal(t, ProviderGitHub, GetDefaultProvider(&Repository{Provider: ProviderGitHub}))
	assert.Equal(t, Provider(""), GetDefaultProvider(&Repository{Address: ""}))
}
