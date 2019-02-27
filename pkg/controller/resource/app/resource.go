package app

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "app"
)

type Config struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	GetCurrentStateFunc func(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error)
	GetDesiredStateFunc func(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error)
}

type Resource struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	getCurrentStateFunc func(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error)
	getDesiredStateFunc func(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error)
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.GetCurrentStateFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetCurrentStateFunc must not be empty", config)
	}
	if config.GetDesiredStateFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetDesiredStateFunc must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		getCurrentStateFunc: config.GetCurrentStateFunc,
		getDesiredStateFunc: config.GetDesiredStateFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func containsAppCR(cr *v1alpha1.App, crs []*v1alpha1.App) bool {
	for _, a := range crs {
		if cr.Name == a.Name && cr.Namespace == a.Namespace {
			return true
		}
	}

	return false
}

func toState(v interface{}) ([]*v1alpha1.App, error) {
	x, ok := v.([]*v1alpha1.App)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", x, v)
	}

	return x, nil
}
