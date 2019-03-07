package status

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToReleaseCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding if custom resource needs to be updated")

		currentLabels := cr.Labels
		if currentLabels == nil {
			currentLabels = map[string]string{}
		}

		v := cr.Labels[key.LabelReleaseCyclePhase]
		if v == string(cr.Status.Cycle.Phase) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found custom resource does not need to be updated")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found custom resource needs to be updated")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the latest version of custom resource")

		cr, err = r.g8sClient.ReleaseV1alpha1().Releases(cr.GetNamespace()).Get(cr.GetName(), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the latest version of custom resource")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating custom resource")

		desiredLabels := cr.Labels
		if desiredLabels == nil {
			desiredLabels = map[string]string{}
		}
		desiredLabels[key.LabelReleaseCyclePhase] = string(cr.Status.Cycle.Phase)

		cr.Labels = desiredLabels

		_, err = r.g8sClient.ReleaseV1alpha1().Releases(cr.GetNamespace()).Update(cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated custom resource")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
