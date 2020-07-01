package collector

import (
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
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
