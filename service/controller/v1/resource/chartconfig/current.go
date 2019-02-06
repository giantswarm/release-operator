package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
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

	releaseCycleCR, err := key.ToReleaseCycleCR(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseCR, err := r.g8sClient.ReleaseV1alpha1().Releases(releaseCycleCR.GetNamespace()).Get(releaseCycleCR.Spec.Release.Name, metav1.GetOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var chartConfigCRs []*v1alpha1.ChartConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding current state")

		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", key.LabelReleaseVersion, key.ReleaseVersion(*releaseCR)),
		}
		list, err := r.g8sClient.CoreV1alpha1().ChartConfigs(r.namespace).List(o)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, c := range list.Items {
			chartConfigCRs = append(chartConfigCRs, &c)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found current state")
	}

	return chartConfigCRs, nil
}
