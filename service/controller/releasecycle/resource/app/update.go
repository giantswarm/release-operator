package app

import (
	"context"
	"fmt"
	"reflect"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/resource/crud"

	"github.com/giantswarm/release-operator/service/controller/key"
)

// ApplyUpdateChange ensures updateChange App CR is updated in k8s api.
func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	appCR, err := key.ToAppCR(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if appCR != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring update of release App CR %#q", appCR.GetName()))

		_, err = r.g8sClient.ApplicationV1alpha1().Apps(key.Namespace).Update(appCR)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured update of release App CR %#q", appCR.GetName()))
	}

	return nil
}

// NewUpdatePatch computes the create and update changes to be applied.
func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

// newUpdateChange computes whether the App CR has to be updated.
//
// nil, nil is returned when no App CR has to be updated.
func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentAppCR, err := key.ToAppCR(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredAppCR, err := key.ToAppCR(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var updateAppCR *applicationv1alpha1.App
	if desiredAppCR != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computing update state %#q", desiredAppCR.GetName()))

		if currentAppCR != nil && currentAppCR.GetName() != "" && isAppCRModified(desiredAppCR, currentAppCR) {
			updateAppCR = desiredAppCR
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed update state %#q", desiredAppCR.GetName()))
	}

	return updateAppCR, nil
}

func isAppCRModified(a, b *applicationv1alpha1.App) bool {
	if !reflect.DeepEqual(a.Spec, b.Spec) {
		return true
	}
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return true
	}

	return false
}
