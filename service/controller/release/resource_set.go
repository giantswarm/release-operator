package release

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/release-operator/service/controller/release/controllercontext"
	"github.com/giantswarm/release-operator/service/controller/release/resource/app"
	"github.com/giantswarm/release-operator/service/controller/release/resource/label"
	"github.com/giantswarm/release-operator/service/controller/release/resource/status"
)

type ResourceSetConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

func NewResourceSet(config ResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var statusResource resource.Interface
	{
		c := status.Config{
			G8sClient: config.G8sClient,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		statusResource, err = status.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var labelResource resource.Interface
	{
		c := label.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		labelResource, err = label.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appResource resource.Interface
	{
		c := app.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		appResource, err = app.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		statusResource,
		labelResource,
		appResource,
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

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		c := controllercontext.Context{}

		ctx = controllercontext.NewContext(ctx, c)

		return ctx, nil
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}

func toCRUDResource(logger micrologger.Logger, ops controller.CRUDResourceOps) (*controller.CRUDResource, error) {
	c := controller.CRUDResourceConfig{
		Logger: logger,
		Ops:    ops,
	}

	r, err := controller.NewCRUDResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
