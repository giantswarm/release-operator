package service

import "github.com/giantswarm/release-operator/flag/service/kubernetes"

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Kubernetes kubernetes.Kubernetes
}
