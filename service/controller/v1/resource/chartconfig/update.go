package chartconfig

import (
	"context"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	chartConfigCRToUpdate, err := toChartConfigCR(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if chartConfigCRToUpdate != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating ChartConfig CR in the Kubernetes API")

		_, err = r.g8sClient.CoreV1alpha1().ChartConfigs(r.namespace).Update(chartConfigCRToUpdate)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated ChartConfig CR in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "ChartConfig CR does not need to be updated in the Kubernetes API")
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
	currentChartConfigCR, err := toChartConfigCR(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigCR, err := toChartConfigCR(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if ChartConfig CR has to be updated")

	var chartConfigCRToUpdate *v1alpha1.ChartConfig
	if isChartConfigCRModified(currentChartConfigCR, desiredChartConfigCR) {
		chartConfigCRToUpdate = desiredChartConfigCR.DeepCopy()
		chartConfigCRToUpdate.ObjectMeta.ResourceVersion = currentChartConfigCR.ObjectMeta.ResourceVersion
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found out if ChartConfig CR has to be updated")

	return chartConfigCRToUpdate, nil
}

func isChartConfigCRModified(a, b *v1alpha1.ChartConfig) bool {
	if !reflect.DeepEqual(a.Spec, b.Spec) {
		return true
	}
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return true
	}

	return false
}
