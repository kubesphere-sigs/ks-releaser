package controllers

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestFindReleaserFile(t *testing.T) {
	result := findReleaserFile("path.go", ".")
	assert.True(t, strings.HasSuffix(result, "path.go"))
}
