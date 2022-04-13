package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	devopsv1alpha1 "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestReleaserReconciler_bumpResource(t *testing.T) {
	schema, err := devopsv1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	defaultReleaser := &devopsv1alpha1.Releaser{
		ObjectMeta: v1.ObjectMeta{
			Namespace:       "fake",
			Name:            "fake-v3.3.0-alpha.0",
			ResourceVersion: "123",
		},
		Spec: devopsv1alpha1.ReleaserSpec{
			Version: "v3.3.0-alpha.0",
		},
	}

	type fields struct {
		logger      logr.Logger
		Client      client.Client
		GitCacheDir string
		gitUser     string
	}
	type args struct {
		releaser *devopsv1alpha1.Releaser
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
		verify  func(t *testing.T, Client client.Client)
	}{{
		name: "pre-release version",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, defaultReleaser.DeepCopy()),
		},
		args: args{
			releaser: defaultReleaser.DeepCopy(),
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
		verify: func(t *testing.T, Client client.Client) {
			item := &devopsv1alpha1.Releaser{}
			err := Client.Get(context.Background(), types.NamespacedName{
				Namespace: "fake",
				Name:      "fake-v3.3.0-alpha.0",
			}, item)
			assert.Nil(t, err)
			assert.Equal(t, devopsv1alpha1.PhaseDone, item.Spec.Phase)

			err = Client.Get(context.Background(), types.NamespacedName{
				Namespace: "fake",
				Name:      "fake-v3.3.0-alpha.1",
			}, item)
			assert.Nil(t, err)
			assert.Equal(t, devopsv1alpha1.PhaseDraft, item.Spec.Phase)

			err = Client.Get(context.Background(), types.NamespacedName{
				Namespace: "fake",
				Name:      "fake-v3.3.0",
			}, item)
			assert.Nil(t, err)
			assert.Equal(t, devopsv1alpha1.PhaseDraft, item.Spec.Phase)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReleaserReconciler{
				logger:      tt.fields.logger,
				Client:      tt.fields.Client,
				GitCacheDir: tt.fields.GitCacheDir,
				gitUser:     tt.fields.gitUser,
			}
			tt.wantErr(t, r.bumpResource(tt.args.releaser), fmt.Sprintf("bumpResource(%v)", tt.args.releaser))
			tt.verify(t, tt.fields.Client)
		})
	}
}
