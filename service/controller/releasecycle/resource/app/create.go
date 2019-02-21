package app

import (
	"context"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/release-operator/service/controller/key"
)

// ApplyCreateChange ensures the App CR is created in the k8s api.
func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	appCR, err := key.ToAppCR(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if appCR != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring creation of release App CR", "app", appCR.GetName())

		_, err = r.g8sClient.ApplicationV1alpha1().Apps(r.namespace).Create(appCR)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured creation of release App CR", "app", appCR.GetName())
	}

	return nil
}

// newCreateChange computes the App CR to be created.
//
// nil, nil is returned when no App CR has to be created.
func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentAppCR, err := key.ToAppCR(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredAppCR, err := key.ToAppCR(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing create state", "app", desiredAppCR.GetName())

	var createAppCR *applicationv1alpha1.App
	{
		if currentAppCR == nil || currentAppCR.GetName() == "" {
			createAppCR = desiredAppCR
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computed create state", "app", desiredAppCR.GetName())

	return createAppCR, nil
}
