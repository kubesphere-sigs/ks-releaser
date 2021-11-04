package controllers

import (
	devopsv1alpha1 "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
	"reflect"
	"testing"
)

func Test_updateReleaserAsYAML(t *testing.T) {
	type args struct {
		data     []byte
		callback func(*devopsv1alpha1.Releaser)
	}
	tests := []struct {
		name       string
		args       args
		wantResult []byte
		wantErr    bool
	}{{
		name: "normal case",
		args: args{
			data: []byte(`apiVersion: devops.kubesphere.io/v1alpha1
kind: Releaser
metadata:
  name: releaser-sample-1
spec:
  phase: ready`),
			callback: func(releaser *devopsv1alpha1.Releaser) {
				releaser.Spec.Phase = devopsv1alpha1.PhaseDone
			},
		},
		wantErr: false,
		wantResult: []byte(`apiVersion: devops.kubesphere.io/v1alpha1
kind: Releaser
metadata:
  creationTimestamp: null
  name: releaser-sample-1
spec:
  phase: done
  secret: {}
status: {}
`),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := updateReleaserAsYAML(tt.args.data, tt.args.callback)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateReleaserAsYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("updateReleaserAsYAML() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
