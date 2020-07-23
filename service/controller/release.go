package controller

import (
	"github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/release-operator/pkg/project"
	"github.com/giantswarm/release-operator/service/controller/release"
)

var (
	releaseControllerName = project.Name() + "-release"
)

type ReleaseConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type Release struct {
	*controller.Controller
}

func NewRelease(config ReleaseConfig) (*Release, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	var resourceSet []resource.Interface
	{
		c := release.ResourceSetConfig{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		resourceSet, err = release.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var releaseController *controller.Controller
	{
		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			Name:      releaseControllerName,
			Resources: resourceSet,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.Release)
			},
		}

		releaseController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Release{
		Controller: releaseController,
	}

	return c, nil
}
