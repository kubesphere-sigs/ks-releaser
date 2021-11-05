package controllers

import (
	"fmt"
	"github.com/blang/semver"
	devopsv1alpha1 "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
	"sigs.k8s.io/yaml"
	"strings"
)

func bumpVersion(versionStr string) (nextVersion string, err error) {
	nextVersion = versionStr // keep using the old version if there's any problem happened

	var version semver.Version
	if version, err = semver.ParseTolerant(versionStr); err != nil {
		err = fmt.Errorf("cannot bump an invalid version: %s, error: %v", versionStr, err)
		return
	}

	if preVersionCount := len(version.Pre); preVersionCount > 0 {
		for i := preVersionCount -1; i >= 0; i-- {
			preVersion := &version.Pre[i]
			if preVersion.IsNumeric() {
				preVersion.VersionNum+=1
				break
			}
		}
	} else {
		version.Patch += 1
	}

	nextVersion = version.String()
	if strings.HasPrefix(versionStr, "v") {
		nextVersion = "v" + nextVersion
	}
	return
}

func bumpReleaser(releaser *devopsv1alpha1.Releaser) {
	currentVersion := releaser.Spec.Version
	nextVersion, _ := bumpVersion(currentVersion)
	if strings.HasSuffix(releaser.Name, currentVersion) {
		nameWithoutVersion := strings.ReplaceAll(releaser.Name, currentVersion, "")
		releaser.Name = nameWithoutVersion + nextVersion
	}

	releaser.Spec.Phase = devopsv1alpha1.PhaseDraft
	releaser.Spec.Version = nextVersion

	for i, _ := range releaser.Spec.Repositories {
		repo := &releaser.Spec.Repositories[i]
		repo.Version, _ = bumpVersion(repo.Version)
	}
	return
}

func bumpReleaserAsData(data []byte) (result []byte, filename string, err error) {
	targetReleaser := &devopsv1alpha1.Releaser{}
	if err = yaml.Unmarshal(data, targetReleaser); err == nil {
		bumpReleaser(targetReleaser)
		filename = targetReleaser.Name + ".yaml"
		result, err = yaml.Marshal(targetReleaser)
	}
	return
}
