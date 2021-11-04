package internal_scm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetGitProvider(t *testing.T) {
	type args struct {
		kind  string
		repo  string
		token string
	}
	tests := []struct {
		name  string
		args  args
		exist bool
	}{{
		name: "github",
		args: args{
			kind: "github",
		},
		exist: true,
	}, {
		name:  "fake",
		args:  args{},
		exist: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetGitProvider(tt.args.kind, tt.args.repo, tt.args.token)
			if tt.exist {
				assert.NotNil(t, got)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}
