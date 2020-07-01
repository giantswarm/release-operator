package collector

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "release_operator"
	subsystem = "release"
)

var (
	ReleaseDesc *prometheus.Desc = prometheus.NewDesc(
		// TODO: What is the most descriptive name for this? Maybe change to "states" or "statuses"?
		prometheus.BuildFQName(namespace, subsystem, "info"),
		"Metrics for Release statuses.",
		[]string{
			labelInstallation,
		},
		nil,
	)
)

type ReleaseCollector struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

type ReleaseCollectorConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

func NewReleaseCollector(config ReleaseCollectorConfig) (*ReleaseCollector, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	rc := &ReleaseCollector{
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return rc, nil
}

func (r *ReleaseCollector) Collect(ch chan<- prometheus.Metric) error {
	return nil
}

func (r *ReleaseCollector) Describe(ch chan<- *prometheus.Desc) error {
	ch <- ReleaseDesc
	return nil
}
