package configmap

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "configmap"
)

// Config represents the configuration used to create a new configmap resource.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	Name      string
	Namespace string
}

// Resource implements the configmap resource.
type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	name      string
	namespace string
}

// New creates a new configured configmap resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Name must not be empty", config)
	}
	if config.Namespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Namespace must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		name:      config.Name,
		namespace: config.Namespace,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
