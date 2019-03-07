package app

import (
	"context"
	"fmt"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/key"
)

func (r *resourceStateGetter) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToReleaseCR(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var currentAppCRs []*applicationv1alpha1.App

	for _, c := range cr.Spec.Components {
		name := appCRName(c)
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding App CR %#q in namespace %#q", name, key.Namespace))

		appCR, err := r.g8sClient.ApplicationV1alpha1().Apps(key.Namespace).Get(name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find App CR %#q in namespace %#q", name, key.Namespace))
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found App CR %#q in namespace %#q", name, key.Namespace))
			currentAppCRs = append(currentAppCRs, appCR)
		}
	}

	return currentAppCRs, nil
}
