package status

import (
	"context"
	"fmt"
	"strings"

	appv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/release-operator/v2/pkg/project"
	"github.com/giantswarm/release-operator/v2/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	release, err := key.ToReleaseCR(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if release.DeletionTimestamp != nil {
		return nil
	}

	components := key.FilterComponents(release.Spec.Components)

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

	// Doing this per-release isn't ideal, can we pass this list to each status reconcile somehow?
	var tenantClusters []tenantCluster
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "searching for running tenant clusters")

		var err error
		tenantClusters, err = r.getCurrentTenantClusters(ctx)
		if err != nil {
			r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("error finding tenant clusters: %s", err))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d running tenant clusters", len(tenantClusters)))
		}
	}

	var releaseInUse bool
	{
		// Get two sets of just deduplicated versions
		releaseVersions, operatorVersions := consolidateClusterVersions(tenantClusters)
		// Check the set of release versions and keep this release if it is used.
		r.logger.Log("level", "debug", "message", fmt.Sprintf("checking release %s", release.Name))
		if releaseVersions[strings.TrimPrefix(release.Name, "v")] { // The release name has a leading `v`
			r.logger.Log("level", "debug", "message", fmt.Sprintf("keeping release %s because it is explicitly used", release.Name))
			releaseInUse = true
		} else {
			for _, o := range key.GetProviderOperators() {
				operatorVersion := getOperatorVersionInRelease(o, release)
				// Check the set of operator versions and keep this release if its operator version is used.
				if operatorVersion != "" && operatorVersions[o][operatorVersion] {
					r.logger.Log("level", "debug", "message", fmt.Sprintf("keeping release %s because a cluster using its operator version (%s) is present", release.Name, operatorVersion))
					releaseInUse = true
				}
			}
		}

		// We don't want to unschedule a kvm-operator with no releases in use as long as it has pods it needs to drain.
		// We check the kvm-operator version (blank if this is an AWS/Azure installation), find all pods with that
		// version bundle version annotation and consider the release to be in use if any were found.
		if !releaseInUse { // small optimization
			operatorVersion := getOperatorVersionInRelease(key.ProviderOperatorKVM, release)
			if operatorVersion != "" { // only execute on KVM
				kvmPodsExist, err := r.getKVMOperatorVersionPodsExist(ctx, operatorVersion)
				if err != nil {
					return microerror.Mask(err)
				}
				if kvmPodsExist {
					releaseInUse = true
				}
			}
		}
	}

	var releaseDeployed bool
	{
		releaseDeployed = true
		for _, component := range components {
			if !key.ComponentDeployed(component, apps.Items) {
				releaseDeployed = false

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("component %#q in release %#q is not deployed", component.Name, release.Name))
			}
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting status for release %#q", release.Name))

		release.Status.Ready = releaseDeployed
		release.Status.InUse = releaseInUse
		err := r.k8sClient.CtrlClient().Status().Update(
			ctx,
			release,
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status set for release %#q", release.Name))
	}

	return nil
}

// Searches the components in a release for the given operator and returns the version.
func getOperatorVersionInRelease(operator string, release *releasev1alpha1.Release) string {
	for _, component := range release.Spec.Components {
		if component.Name == operator {
			return component.Version
		}
	}
	return ""
}
