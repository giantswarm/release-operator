package status

import (
	"context"
	"fmt"

	appv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/release-operator/pkg/project"
	"github.com/giantswarm/release-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	release, err := key.ToReleaseCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if release.DeletionTimestamp != nil {
		return nil
	}

	operators := key.ExtractOperators(release.Spec.Components)

	var apps appv1alpha1.AppList
	{
		err := r.k8sClient.CtrlClient().List(
			ctx,
			&apps,
			&client.ListOptions{
				LabelSelector: labels.SelectorFromSet(labels.Set{
					key.LabelManagedBy: project.Name(),
				}),
			},
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var releaseDeployed bool
	{
		releaseDeployed = true
		for _, operator := range operators {
			if !key.OperatorDeployed(apps.Items, operator) {
				releaseDeployed = false
			}
		}
	}

	{
		// TODO: Actually check if we actually need to update
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting status for release %#q in namespace %#q", release.Name, release.Namespace))

		release.Status.Ready = releaseDeployed
		err := r.k8sClient.CtrlClient().Status().Update(
			ctx,
			release,
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status set for release %#q in namespace %#q", release.Name, release.Namespace))
	}

	return nil
}
