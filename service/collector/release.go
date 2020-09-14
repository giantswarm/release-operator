package collector

import (
	"context"
	"strconv"

	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gaugeValue float64 = 1
)

const (
	namespace = "release_operator"
	subsystem = "release"
)

var (
	ReleaseDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "status"),
		"Metric about the status of a Release.",
		[]string{
			labelName,
			labelState,
			labelReady,
			labelInUse,
		},
		nil,
	)
)

type ReleaseCollector struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

type ReleaseCollectorConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

func NewReleaseCollector(config ReleaseCollectorConfig) (*ReleaseCollector, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	rc := &ReleaseCollector{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
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
	releases, err := r.k8sClient.G8sClient().ReleaseV1alpha1().Releases().List(ctx, metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, release := range releases.Items {
		ch <- prometheus.MustNewConstMetric(
			ReleaseDesc,
			prometheus.GaugeValue,
			gaugeValue,
			release.Name,
			release.Spec.State.String(),
			strconv.FormatBool(release.Status.Ready),
			strconv.FormatBool(release.Status.InUse),
		)
	}

	return nil
}
