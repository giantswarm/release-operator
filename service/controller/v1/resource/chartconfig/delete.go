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
	chartConfigCRToDelete, err := toChartConfigCR(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if chartConfigCRToDelete != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting ChartConfig CR in the Kubernetes API")

		err = r.g8sClient.CoreV1alpha1().ChartConfigs(r.namespace).Delete(chartConfigCRToDelete.Name, &metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted ChartConfig CR in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "ChartConfig CR does not have to be deleted in the Kubernetes API")
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
	currentChartConfigCR, err := toChartConfigCR(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the ChartConfig CR has to be deleted")

	var chartConfigCRToDelete *v1alpha1.ChartConfig
	if currentChartConfigCR != nil {
		chartConfigCRToDelete = currentChartConfigCR
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found out if the ChartConfig CR has to be deleted")

	return chartConfigCRToDelete, nil
}
