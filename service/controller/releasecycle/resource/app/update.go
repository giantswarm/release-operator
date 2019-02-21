package app

import (
	"context"
	"reflect"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/release-operator/service/controller/key"
)

// ApplyUpdateChange ensures updateChange App CR is updated in k8s api.
func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	appCR, err := key.ToAppCR(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if appCR != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring update of release App CR", "app", appCR.GetName())

		_, err = r.g8sClient.ApplicationV1alpha1().Apps(r.namespace).Update(appCR)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured update of release App CR", "app", appCR.GetName())
	}

	return nil
}

// NewUpdatePatch computes the create and update changes to be applied.
func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

// newUpdateChange computes the App CR to be updated.
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

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing update state", "app", desiredAppCR.GetName())

	var updateAppCR *applicationv1alpha1.App
	{
		if currentAppCR != nil && currentAppCR.GetName() != "" && isAppCRModified(desiredAppCR, currentAppCR) {
			updateAppCR = desiredAppCR
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computed update state", "app", desiredAppCR.GetName())

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
