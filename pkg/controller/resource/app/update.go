package app

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"

	"github.com/giantswarm/release-operator/service/controller/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	appCR, err := key.ToAppCR(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if appCR != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating App CR %#q in namespace %#q", appCR.Name, appCR.Namespace))

		_, err = r.g8sClient.ApplicationV1alpha1().Apps(appCR.Namespace).Update(appCR)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated App CR %#q in namespace %#q", appCR.Name, appCR.Namespace))
	}

	return nil
}

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

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentAppCRs, err := toAppCRs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredAppCRs, err := toAppCRs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var appCRsToUpdate []*v1alpha1.App
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computing App CRs to update"))

		for _, c := range currentAppCRs {
			for _, d := range desiredAppCRs {
				m := newAppCRToUpdate(c, d)
				if m != nil {
					appCRsToUpdate = append(appCRsToUpdate, m)
				}
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed %d App CRs to update", appCRsToUpdate))
	}

	return appCRsToUpdate, nil
}

func newAppCRToUpdate(current, desired *v1alpha1.App) *v1alpha1.App {
	if current.Namespace != desired.Namespace {
		return nil
	}
	if current.Name != desired.Name {
		return nil
	}

	merged := current.DeepCopy()

	merged.Annotations = desired.Annotations
	merged.Labels = desired.Labels

	merged.Spec = desired.Spec

	if reflect.DeepEqual(current, merged) {
		return nil
	}

	return merged
}
