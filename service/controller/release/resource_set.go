package release

import (
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"

	"github.com/giantswarm/release-operator/service/controller/release/resource/apps"
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

	resources := []resource.Interface{
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
