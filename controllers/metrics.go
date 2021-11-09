package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	releaserTagActionCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "releaser_tag_total",
			Help: "Number of tag actions",
		},
	)
	releaserReleaseActionCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "releaser_release_total",
			Help: "Number of release actions",
		},
	)
	releaserPreReleaseActionCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "releaser_prerelease_total",
			Help: "Number of preRelease actions",
		},
	)
	releaserGitOpsCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "releaser_gitops_total",
			Help: "Number of using GitOps way",
		},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(releaserTagActionCount,
		releaserPreReleaseActionCount,
		releaserReleaseActionCount,
		releaserGitOpsCount)
}

func increaseTagActionCount() {
	releaserTagActionCount.Inc()
}

func increasePreReleaseActionCount() {
	releaserPreReleaseActionCount.Inc()
}

func increaseReleaseActionCount() {
	releaserReleaseActionCount.Inc()
}

func increaseGitOpsCount() {
	releaserGitOpsCount.Inc()
}
