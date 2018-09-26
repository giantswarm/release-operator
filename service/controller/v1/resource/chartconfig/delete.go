package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	chartConfigCRsToDelete, err := toChartConfigCRs(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigCRsToDelete) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting ChartConfig CRs in the Kubernetes API")

		for _, c := range chartConfigCRsToDelete {
			err := r.g8sClient.CoreV1alpha1().ChartConfigs(r.namespace).Delete(c.GetName(), &metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted ChartConfig CRs in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "ChartConfig CRs do not have to be deleted in the Kubernetes API")
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
	currentChartConfigCRs, err := toChartConfigCRs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigCRs, err := toChartConfigCRs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var chartConfigCRsToDelete []*v1alpha1.ChartConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing delete state")

		for _, c := range currentChartConfigCRs {
			if containsChartConfigCRs(desiredChartConfigCRs, c) {
				chartConfigCRsToDelete = append(chartConfigCRsToDelete, c)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed delete state")
	}

	return chartConfigCRsToDelete, nil
}
