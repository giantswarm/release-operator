package app

import (
	"fmt"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "app"
)

// Config represents the configuration used to create a new app resource.
type Config struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// Resource implements the app resource.
//
// It ensures each release cycle has its corresponding release installed.
// It does so by creating an App CR for the release, which will then be
// installed by app-operator.
// Note: releases are never removed, so removing a release cycle CR has no effect
// 	 on the previously installed release.
type Resource struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

// New creates a new configured app resource.
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

	r := &Resource{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func getAppCR(list []*applicationv1alpha1.App, namespace, name string) (*applicationv1alpha1.App, bool) {
	for _, l := range list {
		b := true
		b = b && l.Namespace == namespace
		b = b && l.Name == name
		if b {
			return l, true
		}
	}

	return nil, false
}

func toAppCRs(v interface{}) ([]*applicationv1alpha1.App, error) {
	appCRs, ok := v.([]*applicationv1alpha1.App)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*applicationv1alpha1.App{}, v)
	}

	return appCRs, nil
}

// releaseAppCRName returns the name of the release App CR for the given release cycle.
func releaseAppCRName(releaseCycleCR *releasev1alpha1.ReleaseCycle) string {
	return releasePrefix(releaseCycleCR.GetName())
}

// releasePrefix adds release- prefix to name.
func releasePrefix(name string) string {
	return fmt.Sprintf("release-%s", name)
}
