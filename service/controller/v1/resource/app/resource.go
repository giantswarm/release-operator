package app

import (
	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "appv1"
)

// Config represents the configuration used to create a new app resource.
type Config struct {
	// Dependencies.
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	Namespace string
}

// Resource implements the app resource.
type Resource struct {
	// Dependencies.
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	namespace string
}

// New creates a new configured app resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Namespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Namespace must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		namespace: config.Namespace,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func containsAppCRs(list []*applicationv1alpha1.App, item *applicationv1alpha1.App) bool {
	_, err := getAppCRByName(list, item.Name)
	if err != nil {
		return false
	}

	return true
}

func getAppCRByName(list []*applicationv1alpha1.App, name string) (*applicationv1alpha1.App, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func toAppCRs(v interface{}) ([]*applicationv1alpha1.App, error) {
	appCRs, ok := v.([]*applicationv1alpha1.App)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*applicationv1alpha1.App{}, v)
	}

	return appCRs, nil
}
