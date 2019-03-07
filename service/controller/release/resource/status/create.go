package status

import (
	"context"
	"fmt"
	"reflect"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToReleaseCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var releaseCycleCR *releasev1alpha1.ReleaseCycle
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding ReleaseCycle CR %#q in namespace %#q", cr.Name, cr.Namespace))

		// Release CR and corresponding ReleaseCycle CR have the same name and namespace.
		releaseCycleCR, err = r.g8sClient.ReleaseV1alpha1().ReleaseCycles(cr.GetNamespace()).Get(cr.GetName(), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			releaseCycleCR = nil
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find ReleaseCycle CR %#q in namespace %#q", cr.Name, cr.Namespace))
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found ReleaseCycle CR %#q in namespace %#q", cr.Name, cr.Namespace))
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding if custom resource status needs to be updated")

		if releaseCycleCR == nil && cr.Status.Cycle.Phase == releasev1alpha1.CyclePhaseUpcoming {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found custom resource status does not need to be updated")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		if reflect.DeepEqual(cr.Status.Cycle, releaseCycleCR.Spec) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found custom resource status does not need to be updated")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found custom resource status needs to be updated")
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating custom resource status")

		if releaseCycleCR == nil {
			cr.Status.Cycle.Phase = releasev1alpha1.CyclePhaseUpcoming
		} else {
			cr.Status.Cycle = releaseCycleCR.Spec
		}

		_, err = r.g8sClient.ReleaseV1alpha1().Releases(cr.GetNamespace()).UpdateStatus(cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated custom resource status")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
