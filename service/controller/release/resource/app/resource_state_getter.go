package app

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/micrologger"
)

type resourceStateGetter struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger

	appCatalog string
}
