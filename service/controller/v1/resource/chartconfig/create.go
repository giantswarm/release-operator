package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	chartConfigCRsToCreate, err := toChartConfigCRs(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigCRsToCreate) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating ChartConfig CRs in the Kubernetes API")

		for _, c := range chartConfigCRsToCreate {
			_, err = r.g8sClient.CoreV1alpha1().ChartConfigs(r.namespace).Create(c)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created ChartConfig CRs in the Kubernetes API")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "ChartConfig CRs do not have to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentChartConfigCRs, err := toChartConfigCRs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigCRs, err := toChartConfigCRs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing create state")

	var chartConfigCRsToCreate []*v1alpha1.ChartConfig

	for _, d := range desiredChartConfigCRs {
		if !containsChartConfigCRs(currentChartConfigCRs, d) {
			chartConfigCRsToCreate = append(chartConfigCRsToCreate, d)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computed create state")

	return chartConfigCRsToCreate, nil
}
