// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"os"
	"testing"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/release-operator/integration/env"
	"github.com/giantswarm/release-operator/service/controller/key"
)

func Setup(m *testing.M, config Config) {
	ctx := context.Background()

	v, err := setup(ctx, m, config)
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "error", "message", "failed to setup test environment", "stack", fmt.Sprintf("%#v", err))
		os.Exit(1)
	}

	os.Exit(v)
}

func setup(ctx context.Context, m *testing.M, config Config) (int, error) {
	var err error

	// Create namespace.
	{
		err := config.K8sSetup.EnsureNamespaceCreated(ctx, key.Namespace)
		if err != nil {
			return 0, microerror.Mask(err)
		}
	}

	// Create App CRD
	{
		crd := applicationv1alpha1.NewAppCRD()

		err := config.K8sSetup.EnsureCRDCreated(ctx, crd)
		if err != nil {
			return 0, microerror.Mask(err)
		}
	}

	// Install tiller.
	{
		err = config.HelmClient.EnsureTillerInstalled(ctx)
		if err != nil {
			return 0, microerror.Mask(err)
		}
	}

	// Install release-operator.
	{
		releaseName := "release-operator"
		chartInfo := release.NewChartInfo("release-operator-chart", env.CircleSHA())
		podNamespace := "giantswarm"
		podLabelSelector := "app=release-operator"

		var values string
		{
			c := chartvalues.ReleaseOperatorConfig{
				RegistryPullSecret: env.RegistryPullSecret(),
			}

			values, err = chartvalues.NewReleaseOperator(c)
			if err != nil {
				return 0, microerror.Mask(err)
			}
		}

		installConditions := []release.ConditionFunc{
			config.Release.Condition().PodExists(ctx, podNamespace, podLabelSelector),
			config.Release.Condition().CRDExists(ctx, releasev1alpha1.NewReleaseCRD()),
			config.Release.Condition().CRDExists(ctx, releasev1alpha1.NewReleaseCycleCRD()),
		}

		err = config.Release.EnsureInstalled(ctx, releaseName, chartInfo, values, installConditions...)
		if err != nil {
			return 0, microerror.Mask(err)
		}
	}

	v := m.Run()
	return v, nil
}
