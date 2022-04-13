package controllers

import (
	"fmt"
	"github.com/blang/semver"
	devopsv1alpha1 "github.com/kubesphere-sigs/ks-releaser/api/v1alpha1"
	"sigs.k8s.io/yaml"
	"strings"
)

func isPreRelease(versionStr string) bool {
	if version, err := semver.ParseTolerant(versionStr); err == nil {
		return len(version.Pre) > 0
	}
	return false
}

func bumpVersionTo(versionStr string, remainPre bool) (nextVersion string, isPre bool, err error) {
	nextVersion = versionStr // keep using the old version if there's any problem happened

	var version semver.Version
	if version, err = semver.ParseTolerant(versionStr); err != nil {
		err = fmt.Errorf("cannot bump an invalid version: %s, error: %v", versionStr, err)
		return
	}

	if preVersionCount := len(version.Pre); preVersionCount > 0 {
		isPre = true
		if remainPre {
			for i := preVersionCount - 1; i >= 0; i-- {
				preVersion := &version.Pre[i]
				if preVersion.IsNumeric() {
					preVersion.VersionNum += 1
					break
				}
			}
		} else {
			version.Pre = nil
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

func bumpVersion(versionStr string) (nextVersion string, isPre bool, err error) {
	nextVersion, isPre, err = bumpVersionTo(versionStr, true)
	return
}

func bumpReleaser(releaser *devopsv1alpha1.Releaser, remainPre bool) (isPre bool) {
	var nextVersion string

	currentVersion := releaser.Spec.Version
	nextVersion, isPre, _ = bumpVersionTo(currentVersion, remainPre)
	if strings.HasSuffix(releaser.Name, currentVersion) {
		nameWithoutVersion := strings.ReplaceAll(releaser.Name, currentVersion, "")
		releaser.Name = nameWithoutVersion + nextVersion
	}

	// remove the metadata
	releaser.ObjectMeta.Generation = 0
	releaser.ObjectMeta.SelfLink = ""
	releaser.ObjectMeta.Annotations = nil
	releaser.ObjectMeta.ResourceVersion = ""

	releaser.Spec.Phase = devopsv1alpha1.PhaseDraft
	releaser.Spec.Version = nextVersion

	for i, _ := range releaser.Spec.Repositories {
		repo := &releaser.Spec.Repositories[i]
		repo.Version, _, _ = bumpVersionTo(repo.Version, remainPre)
	}

	// remove status
	releaser.Status = devopsv1alpha1.ReleaserStatus{}
	return
}

func bumpReleaserAsData(data []byte, remainPre bool) (result []byte, filename string, isPre bool, err error) {
	targetReleaser := &devopsv1alpha1.Releaser{}
	if err = yaml.Unmarshal(data, targetReleaser); err == nil {
		isPre = bumpReleaser(targetReleaser, remainPre)
		filename = targetReleaser.Name + ".yaml"
		result, err = yaml.Marshal(targetReleaser)
	}
	return
}
