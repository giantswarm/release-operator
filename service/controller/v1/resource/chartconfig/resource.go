package chartconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "chartconfigv1"
)

// Config represents the configuration used to create a new chartconfig resource.
type Config struct {
	// Dependencies.
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	ChartOperatorVersion string
	Namespace            string
}

// Resource implements the chartconfig resource.
type Resource struct {
	// Dependencies.
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	chartOperatorVersion string
	namespace            string
}

// New creates a new configured chartconfig resource.
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

	if config.ChartOperatorVersion == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ChartOperatorVersion must not be empty", config)
	}
	if config.Namespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Namespace must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		chartOperatorVersion: config.ChartOperatorVersion,
		namespace:            config.Namespace,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func containsChartConfigCRs(list []*v1alpha1.ChartConfig, item *v1alpha1.ChartConfig) bool {
	_, err := getChartConfigCRByName(list, item.Name)
	if err != nil {
		return false
	}

	return true
}

func getChartConfigCRByName(list []*v1alpha1.ChartConfig, name string) (*v1alpha1.ChartConfig, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func toChartConfigCR(v interface{}) (*v1alpha1.ChartConfig, error) {
	chartConfigCRPointer, ok := v.(*v1alpha1.ChartConfig)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.ChartConfig{}, v)
	}

	return chartConfigCRPointer, nil
}

func toChartConfigCRs(v interface{}) ([]*v1alpha1.ChartConfig, error) {
	chartConfigCRsPointer, ok := v.([]*v1alpha1.ChartConfig)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*v1alpha1.ChartConfig{}, v)
	}

	return chartConfigCRsPointer, nil
}
