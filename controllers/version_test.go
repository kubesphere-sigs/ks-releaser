package controllers

import (
	"fmt"
	devopsv1alpha1 "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strings"
	"testing"
)

func TestVersionBump(t *testing.T) {
	type testCase struct {
		name string
		arg  struct {
			version string
		}
		wantErr     bool
		wantVersion string
	}

	testCases := []testCase{{
		name: "invalid version string",
		arg: struct{ version string }{
			version: "abc",
		},
		wantErr:     true,
		wantVersion: "abc",
	}, {
		name: "valid version string",
		arg: struct{ version string }{
			version: "v1.0.0",
		},
		wantErr:     false,
		wantVersion: "v1.0.1",
	}, {
		name: "valid version string, without patch number",
		arg: struct{ version string }{
			version: "v1.0",
		},
		wantErr:     false,
		wantVersion: "v1.0.1",
	}}

	for i, _ := range testCases {
		caseItem := testCases[i]

		nextVersion, err := bumpVersion(caseItem.arg.version)
		if caseItem.wantErr {
			assert.NotNil(t, err, fmt.Sprintf("test failed with case[%d]", i))
		}

		assert.Equal(t, caseItem.wantVersion, nextVersion, fmt.Sprintf("test failed with case[%d]", i))
	}
}

func Test_bumpReleaser(t *testing.T) {
	type args struct {
		releaser *devopsv1alpha1.Releaser
	}
	tests := []struct {
		name       string
		args       args
		wantResult *devopsv1alpha1.Releaser
		wantErr    bool
	}{{
		name: "only have main version",
		args: args{
			releaser: &devopsv1alpha1.Releaser{
				Spec: devopsv1alpha1.ReleaserSpec{Version: "v1.0.1"},
			},
		},
		wantErr: false,
		wantResult: &devopsv1alpha1.Releaser{
			Spec: devopsv1alpha1.ReleaserSpec{
				Phase:   devopsv1alpha1.PhaseDraft,
				Version: "v1.0.2"},
		},
	}, {
		name: "have repositories",
		args: args{
			releaser: &devopsv1alpha1.Releaser{
				Spec: devopsv1alpha1.ReleaserSpec{
					Version: "v1.0.1",
					Repositories: []devopsv1alpha1.Repository{{
						Version: "v1.2.3",
					}},
				},
			}},
		wantErr: false,
		wantResult: &devopsv1alpha1.Releaser{
			Spec: devopsv1alpha1.ReleaserSpec{
				Phase:   devopsv1alpha1.PhaseDraft,
				Version: "v1.0.2",
				Repositories: []devopsv1alpha1.Repository{{
					Version: "v1.2.4",
				}},
			},
		},
	}, {
		name: "bump cr name",
		args: args{
			releaser: &devopsv1alpha1.Releaser{
				ObjectMeta: metav1.ObjectMeta{Name: "test-v1.0.1"},
				Spec:       devopsv1alpha1.ReleaserSpec{Version: "v1.0.1"},
			},
		},
		wantErr: false,
		wantResult: &devopsv1alpha1.Releaser{
			ObjectMeta: metav1.ObjectMeta{Name: "test-v1.0.2"},
			Spec: devopsv1alpha1.ReleaserSpec{
				Phase:   devopsv1alpha1.PhaseDraft,
				Version: "v1.0.2"},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bumpReleaser(tt.args.releaser)
			if !reflect.DeepEqual(tt.args.releaser, tt.wantResult) {
				t.Errorf("bumpReleaser() gotResult = %v, want %v", tt.args.releaser, tt.wantResult)
			}
		})
	}
}

func Test_bumpReleaserAsData(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
		wantErr    bool
	}{{
		name: "normal case",
		args: args{
			data: `apiVersion: devops.kubesphere.io/v1alpha1
kind: Releaser
metadata:
  creationTimestamp: null
  name: ks-releaser-v0.0.5
spec:
  version: v0.0.5`,
		},
		wantResult: `apiVersion: devops.kubesphere.io/v1alpha1
kind: Releaser
metadata:
  creationTimestamp: null
  name: ks-releaser-v0.0.6
spec:
  phase: draft
  secret: {}
  version: v0.0.6
status: {}
`,
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, _, err := bumpReleaserAsData([]byte(tt.args.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("bumpReleaserAsData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if strings.TrimSpace((string(gotResult))) != strings.TrimSpace(tt.wantResult) {
				t.Errorf("bumpReleaserAsData() gotResult = %s, want %s", string(gotResult), tt.wantResult)
			}
		})
	}
}