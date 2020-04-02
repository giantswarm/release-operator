package label

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "label"
)

type Config struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

// Resource sets "release-operator.giantswarm.io/release-cycle-phase" on the
// observed CR. The value is taken from .Status.Cycle.Phase. This is to allow
// selecting non-EOL Release CRs efficiently using label selector.
//
// Label resource and status resource are separated to be able to cancel
// reconciliation independently. There are two requests one to update status
// and second to update labels which generate separate events.
type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
