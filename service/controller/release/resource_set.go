package release

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/release-operator/service/controller/release/resource/apps"
)

type ResourceSetConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

func NewResourceSet(config ResourceSetConfig) ([]resource.Interface, error) {
	var err error

	var appsResource resource.Interface
	{
		c := apps.Config{
			G8sClient: config.G8sClient,
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
