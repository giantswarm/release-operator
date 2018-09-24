package chartconfig

import (
	"context"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	chartConfigCRsToUpdate, err := toChartConfigCRs(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigCRsToUpdate) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating ChartConfig CRs in the Kubernetes API")

		for _, c := range chartConfigCRsToUpdate {
			_, err = r.g8sClient.CoreV1alpha1().ChartConfigs(r.namespace).Update(c)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated ChartConfig CRs in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "ChartConfig CRs does not have to be updated in the Kubernetes API")
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
	currentChartConfigCRs, err := toChartConfigCRs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigCRs, err := toChartConfigCRs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var chartConfigCRsToUpdate []*v1alpha1.ChartConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing update state")

		for _, c := range currentChartConfigCRs {
			d, err := getChartConfigCRByName(desiredChartConfigCRs, c.Name)
			if IsNotFound(err) {
				continue
			} else if err != nil {
				return nil, microerror.Mask(err)
			}

			if isChartConfigCRModified(d, c) {
				u := d.DeepCopy()
				u.ObjectMeta.ResourceVersion = c.ObjectMeta.ResourceVersion

				chartConfigCRsToUpdate = append(chartConfigCRsToUpdate, u)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed update state")
	}

	return chartConfigCRsToUpdate, nil
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
