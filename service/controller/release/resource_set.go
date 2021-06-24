package release

import (
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v5/pkg/resource"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/wrapper/retryresource"

	"github.com/giantswarm/release-operator/v2/service/controller/release/resource/apps"
	"github.com/giantswarm/release-operator/v2/service/controller/release/resource/configs"
	"github.com/giantswarm/release-operator/v2/service/controller/release/resource/status"
)

type ResourceSetConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

func NewResourceSet(config ResourceSetConfig) ([]resource.Interface, error) {
	var err error

	var appsResource resource.Interface
	{
		c := apps.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		appsResource, err = apps.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configsResource resource.Interface
	{
		c := configs.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		configsResource, err = configs.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var statusResource resource.Interface
	{
		c := status.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		statusResource, err = status.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		statusResource,
		configsResource,
		appsResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}
