package app

import (
	"context"
	"fmt"
	"reflect"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	appCRsToUpdate, err := toAppCRs(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(appCRsToUpdate) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating %d App CRs in the Kubernetes API", len(appCRs)))

		for _, c := range appCRsToUpdate {
			_, err = r.g8sClient.ApplicationV1alpha1().Apps(r.namespace).Update(c)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated %d App CRs in the Kubernetes API", len(appCRs)))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "App CRs do not have to be updated in the Kubernetes API")
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

	var appCRsToUpdate []*applicationv1alpha1.App
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing update state")

		for _, c := range currentAppCRs {
			d, err := getAppCRByName(desiredAppCRs, c.Name)
			if IsNotFound(err) {
				continue
			} else if err != nil {
				return nil, microerror.Mask(err)
			}

			if isAppCRModified(d, c) {
				u := d.DeepCopy()
				u.ObjectMeta.ResourceVersion = c.ObjectMeta.ResourceVersion

				appCRsToUpdate = append(appCRsToUpdate, u)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed update state")
	}

	return appCRsToUpdate, nil
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
