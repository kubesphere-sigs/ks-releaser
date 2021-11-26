package controllers

import (
	"github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetFailedConditions(t *testing.T) {
	assert.Nil(t, getFailedConditions(nil))
	assert.Nil(t, getFailedConditions([]v1alpha1.Condition{}))

	assert.Equal(t, 1, len(getFailedConditions([]v1alpha1.Condition{{
		Status: v1alpha1.ConditionStatusFailed,
	}})))
	assert.Equal(t, 1, len(getFailedConditions([]v1alpha1.Condition{{
		Status: v1alpha1.ConditionStatusFailed,
	}, {
		Status: v1alpha1.ConditionStatusSuccess,
	}})))
}

func TestErrorReportRender(t *testing.T) {
	message, err := errorReportRender(&v1alpha1.Releaser{Status: v1alpha1.ReleaserStatus{Conditions: []v1alpha1.Condition{{
		ConditionType: v1alpha1.ConditionTypeOther,
		Status:        v1alpha1.ConditionStatusFailed,
		Message:       "message",
	}}}})
	assert.Nil(t, err)
	assert.Equal(t, `Errors found with releaser: 
|Type|Status|Message|
|---|---|---|
|other|failed|message|
`, message)

	// with whitespace in the message
	message, err = errorReportRender(&v1alpha1.Releaser{Status: v1alpha1.ReleaserStatus{Conditions: []v1alpha1.Condition{{
		ConditionType: v1alpha1.ConditionTypeOther,
		Status:        v1alpha1.ConditionStatusFailed,
		Message:       `
message
`,
	}}}})
	assert.Nil(t, err)
	assert.Equal(t, `Errors found with releaser: 
|Type|Status|Message|
|---|---|---|
|other|failed|message|
`, message)
}
