package collector

import (
	"context"

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
	logger micrologger.Logger

	installationName string
}

type ReleaseCollectorConfig struct {
	Logger micrologger.Logger

	InstallationName string
}

func NewReleaseCollector(config ReleaseCollectorConfig) (*ReleaseCollector, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	rc := &ReleaseCollector{
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return rc, nil
}

func (r *ReleaseCollector) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	r.logger.LogCtx(ctx, "level", "debug", "message", "collecting metrics")

	err := r.collectReleaseStatus(ctx, ch)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finished collecting metrics")
	return nil
}

func (r *ReleaseCollector) Describe(ch chan<- *prometheus.Desc) error {
	ch <- ReleaseDesc
	return nil
}

func (r *ReleaseCollector) collectReleaseStatus(ctx context.Context, ch chan<- prometheus.Metric) error {
	return nil
}
