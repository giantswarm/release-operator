package app

import (
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/release-operator/pkg/controller/resource/app"
)

const (
	Name = "app"
)

type Config struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	AppCatalog string
}

// New returns a resource creating/updating App CRs for components in non-EOL
// releases and deleting App CRs for components in EOL releases. App CRs for
// components existing in both EOL and non-EOL releases are not deleted by the
// returned resource.
func New(config Config) (controller.Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.AppCatalog == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.AppCatalog must not be empty", config)
	}

	var err error

	stateGetter := &resourceStateGetter{
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		appCatalog: config.AppCatalog,
	}

	var appResource *app.Resource
	{
		c := app.Config{
			G8sClient:   config.G8sClient,
			Logger:      config.Logger,
			StateGetter: stateGetter,

			Name: Name,
		}

		appResource, err = app.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var r *controller.CRUDResource
	{
		c := controller.CRUDResourceConfig{
			Logger: config.Logger,
			Ops:    appResource,
		}

		controller.NewCRUDResource(c)
	}

	return r, nil
}

func appCRName(c releasev1alpha1.ReleaseSpecComponent) string {
	return c.Name + "." + c.Version
}
