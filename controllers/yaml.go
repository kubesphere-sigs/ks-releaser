package controllers

import (
	devopsv1alpha1 "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
	"sigs.k8s.io/yaml"
)

func updateReleaserAsYAML(data []byte, callback func(*devopsv1alpha1.Releaser)) (result []byte, err error) {
	targetReleaser := &devopsv1alpha1.Releaser{}
	if err = yaml.Unmarshal(data, targetReleaser); err == nil {
		callback(targetReleaser)

		result, err = yaml.Marshal(targetReleaser)
	}
	return
}
