package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	chartConfigCRToCreate, err := toChartConfigCR(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if chartConfigCRToCreate != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating ChartConfig CR in the Kubernetes API")

		_, err = r.g8sClient.CoreV1alpha1().ChartConfigs(r.namespace).Create(chartConfigCRToCreate)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created ChartConfig CR in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "ChartConfig CR does not have to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentChartConfigCR, err := toChartConfigCR(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigCR, err := toChartConfigCR(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out if the ChartConfig CR has to be created")

	var chartConfigCRToCreate *v1alpha1.ChartConfig
	if currentChartConfigCR == nil {
		chartConfigCRToCreate = desiredChartConfigCR
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found out if the ChartConfig CR has to be created")

	return chartConfigCRToCreate, nil
}
