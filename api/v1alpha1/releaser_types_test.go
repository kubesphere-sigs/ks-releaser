package v1alpha1

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPhase(t *testing.T) {
	assert.False(t, Phase("unknown").IsValid())
	assert.True(t, Phase("draft").IsValid())
	assert.True(t, Phase("ready").IsValid())
	assert.True(t, Phase("done").IsValid())
}
