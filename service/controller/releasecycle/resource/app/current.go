package app

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/key"
)

// GetCurrentState collects the current App CR for the release referenced in ReleaseCycle CR from k8s api.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseCycleCR, err := key.ToReleaseCycleCR(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	appName := releaseAppCRName(releaseCycleCR)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding current state %#q", appName))

	appCR, err := r.g8sClient.ApplicationV1alpha1().Apps(key.Namespace).Get(appName, v1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// fallthrough
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found current state %#q", appName))

	return appCR, nil
}
