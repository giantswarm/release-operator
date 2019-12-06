package controller

import (
	"github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/release-operator/pkg/project"
	"github.com/giantswarm/release-operator/service/controller/releasecycle"
)

var (
	releaseCycleControllerName = project.Name() + "-releasecycle"
)

type ReleaseCycleConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type ReleaseCycle struct {
	*controller.Controller
}

func NewReleaseCycle(config ReleaseCycleConfig) (*ReleaseCycle, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	var err error

	var resourceSet *controller.ResourceSet
	{
		c := releasecycle.ResourceSetConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		resourceSet, err = releasecycle.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var releaseController *controller.Controller
	{
		c := controller.Config{
			CRD:       v1alpha1.NewReleaseCycleCRD(),
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			Name:      releaseCycleControllerName,
			ResourceSets: []*controller.ResourceSet{
				resourceSet,
			},
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.ReleaseCycle)
			},
		}

		releaseController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &ReleaseCycle{
		Controller: releaseController,
	}

	return c, nil
}
