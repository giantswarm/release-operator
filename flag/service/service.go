package service

import (
	"github.com/giantswarm/operatorkit/v2/pkg/flag/service/kubernetes"

	"github.com/giantswarm/release-operator/v2/flag/service/app"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	App        app.App
	Kubernetes kubernetes.Kubernetes
}
