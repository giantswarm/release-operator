package collector

import (
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type helperConfig struct {
	Clients k8sclient.Interface
	Logger  micrologger.Logger
}

type helper struct {
	clients k8sclient.Interface
	logger  micrologger.Logger
}

func newHelper(config helperConfig) (*helper, error) {
	if config.Clients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	h := &helper{
		clients: config.Clients,
		logger:  config.Logger,
	}

	return h, nil
}
