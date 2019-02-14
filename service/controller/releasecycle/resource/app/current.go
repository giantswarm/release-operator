package app

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/release-operator/service/controller/key"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetCurrentState collects the current App CR for the release referenced in obj ReleaseCycle CR from k8s api.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseCycleCR, err := key.ToReleaseCycleCR(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	appName := key.ReleaseAppCRName(releaseCycleCR)
	r.logger.LogCtx(ctx, "level", "debug", "message", "finding current state", "app", appName)

	appCR, err := r.g8sClient.ApplicationV1alpha1().Apps(r.namespace).Get(appName, v1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// fallthrough
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found current state", "app", appName)

	return appCR, nil
}
