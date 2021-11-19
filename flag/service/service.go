package service

import "github.com/giantswarm/operatorkit/v6/pkg/flag/service/kubernetes"

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Kubernetes kubernetes.Kubernetes
}
