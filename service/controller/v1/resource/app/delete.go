package app

import (
	"context"
	"fmt"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	appCRs, err := toAppCRs(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(appCRs) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting %d App CRs in the Kubernetes API", len(appCRs)))

		for _, c := range appCRs {
			err := r.g8sClient.ApplicationV1alpha1().Apps(r.namespace).Delete(c.GetName(), &metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted %d App CRs in the Kubernetes API", len(appCRs)))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "App CRs do not have to be deleted in the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentAppCRs, err := toAppCRs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredAppCRs, err := toAppCRs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var appCRsToDelete []*applicationv1alpha1.App
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing delete state")

		for _, c := range currentAppCRs {
			if containsAppCRs(desiredAppCRs, c) {
				appCRsToDelete = append(appCRsToDelete, c)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed delete state")
	}

	return appCRsToDelete, nil
}
