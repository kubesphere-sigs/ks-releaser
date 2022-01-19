//go:build integration
// +build integration

package controllers

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestRemoteTagExists(t *testing.T) {
	repo, err := clone("https://gitee.com/linuxsuren/test", "master", nil, "bin/tmp")
	assert.Nil(t, err)

	result := remoteTagExists("v0.0.7", repo)
	assert.True(t, result)

	result = remoteTagExists("v0.0.8", repo)
	assert.True(t, result)

	def := rand.New(rand.NewSource(time.Now().UnixNano()))

	// not existing
	fakeTag := strconv.Itoa(def.Int())
	result = remoteTagExists(fakeTag, repo)
	assert.False(t, result)

	result, err = setTag(repo, fakeTag, "message", "user")
	assert.True(t, result)
	assert.Nil(t, err)

	result = remoteTagExists(fakeTag, repo)
	assert.True(t, result)

	err = pushTags(repo, "xx", nil)
	assert.NotNil(t, err)

	_, err = clone("xx://wrong url format", "xxx", nil, "bin/tmp")
	assert.NotNil(t, err)
}
