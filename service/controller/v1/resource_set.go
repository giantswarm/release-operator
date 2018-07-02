package v1

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"k8s.io/client-go/kubernetes"
)

const (
	// ResourceRetries presents number of retries for failed Resource
	// operation before giving up.
	ResourceRetries uint64 = 3
)

type ResourceSetConfig struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Settings.
	ProjectName string
}

func NewResourceSet(config ResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Settings.
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	resources := []controller.Resource{
	// TODO: implement configmap resource
	}

	{
		c := retryresource.WrapConfig{
			BackOffFactory: func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) },
			Logger:         config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{
			Name: config.ProjectName,
		}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		// TODO: implement key package

		return true
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
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
