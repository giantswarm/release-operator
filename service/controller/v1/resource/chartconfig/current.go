package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/v1/controllercontext"
	"github.com/giantswarm/release-operator/service/controller/v1/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	c, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	err = c.Validate()
	if controllercontext.IsInvalidContext(err) {
		// In case the controller context is not valid we miss certain information
		// necessary to compute the chart config CRs. In this case we stop here.
		r.logger.LogCtx(ctx, "level", "debug", "message", "invalid controller context")
		r.logger.LogCtx(ctx, "level", "debug", "message", "cannot compute chart config CR")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	customResource, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var chartConfigCR *v1alpha1.ChartConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding current state")

		m, err := r.g8sClient.CoreV1alpha1().ChartConfigs(r.namespace).Get(key.OperatorChartName(customResource), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find current state")
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found current state")
			chartConfigCR = m
		}
	}

	return chartConfigCR, nil
}
