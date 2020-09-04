package apps

import (
	"context"
	"fmt"

	appv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/release-operator/pkg/project"
	"github.com/giantswarm/release-operator/service/controller/key"
)

const (
	Name = "apps"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensureState(ctx context.Context) error {
	var releases releasev1alpha1.ReleaseList
	{
		err := r.k8sClient.CtrlClient().List(
			ctx,
			&releases,
		)
		if err != nil {
			return microerror.Mask(err)
		}
		releases = excludeDeletedRelease(releases)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "searching for running tenant clusters")
	var tenantClusters []TenantCluster
	{
		var err error
		tenantClusters, err = r.getCurrentTenantClusters(ctx)
		if err != nil {
			r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("error finding tenant clusters: %s", err))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d running tenant clusters", len(tenantClusters)))
		}
	}

	releases, err := r.excludeUnusedDeprecatedReleases(releases, tenantClusters)
	if err != nil {
		return microerror.Mask(err)
	}

	var components map[string]releasev1alpha1.ReleaseSpecComponent
	{
		components = key.ExtractComponents(releases)
	}

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

	appsToDelete := calculateObsoleteApps(components, apps)
	for _, app := range appsToDelete.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting app %#q in namespace %#q", app.Name, app.Namespace))

		err := r.k8sClient.CtrlClient().Delete(
			ctx,
			&app,
		)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted app %#q in namespace %#q", app.Name, app.Namespace))
	}

	appsToCreate := calculateMissingApps(components, apps)
	for _, app := range appsToCreate.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating app %#q in namespace %#q", app.Name, app.Namespace))

		err := r.k8sClient.CtrlClient().Create(
			ctx,
			&app,
		)
		if apierrors.IsAlreadyExists(err) {
			// fall through.
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created app %#q in namespace %#q", app.Name, app.Namespace))
	}

	return nil
}

func (r *Resource) excludeUnusedDeprecatedReleases(releases releasev1alpha1.ReleaseList, clusters []TenantCluster) (releasev1alpha1.ReleaseList, error) {
	// Go over releases
	// If deprecated, AND no existing cluster with this version, exclude it
	// A cluster with this version either:
	//    - has release label with this release version, or
	//    - uses the provider operator version specified in this release

	// Get two sets of just deduplicated versions
	releaseVersions, operatorVersions := consolidateClusterVersions(clusters)

	var active releasev1alpha1.ReleaseList
	for _, release := range releases.Items {
		if release.Spec.State == "deprecated" { // TODO: Should make this constant public in apiextensions
			// Check the set of release versions and keep this release if it is used.
			if releaseVersions[release.Name] {
				active.Items = append(active.Items, release)
				r.logger.Log("level", "debug", "message", fmt.Sprintf("keeping release %s because it is explicitly used", release.Name))
			} else {
				for _, o := range key.GetProviderOperators() {
					operatorVersion := getOperatorVersionInRelease(o, release)
					// Check the set of operator versions and keep this release if its operator version is used.
					if operatorVersion != "" && operatorVersions[o][operatorVersion] {
						active.Items = append(active.Items, release)
						r.logger.Log("level", "debug", "message", fmt.Sprintf("keeping release %s because a cluster using its operator version (%s) is present", release.Name, operatorVersion))
					}
				}
			}
		} else {
			active.Items = append(active.Items, release)
		}
	}

	r.logger.Log("level", "debug", "message", fmt.Sprintf("excluded %d unused deprecated releases", len(releases.Items)-len(active.Items)))

	return active, nil
}

// Searches the components in a release for the given operator and returns the version.
func getOperatorVersionInRelease(operator string, release releasev1alpha1.Release) string {
	for _, component := range release.Spec.Components {
		if component.Name == operator {
			return component.Version
		}
	}
	return ""
}

func calculateMissingApps(components map[string]releasev1alpha1.ReleaseSpecComponent, apps appv1alpha1.AppList) appv1alpha1.AppList {
	var missingApps appv1alpha1.AppList

	for _, component := range components {
		if !key.ComponentCreated(component, apps.Items) {
			missingApp := key.ConstructApp(component)
			missingApps.Items = append(missingApps.Items, missingApp)
		}
	}

	return missingApps
}

func calculateObsoleteApps(components map[string]releasev1alpha1.ReleaseSpecComponent, apps appv1alpha1.AppList) appv1alpha1.AppList {
	var obsoleteApps appv1alpha1.AppList

	for _, app := range apps.Items {
		if !key.AppReferenced(app, components) {
			obsoleteApps.Items = append(obsoleteApps.Items, app)
		}
	}

	return obsoleteApps
}

func excludeDeletedRelease(releases releasev1alpha1.ReleaseList) releasev1alpha1.ReleaseList {
	var active releasev1alpha1.ReleaseList
	for _, release := range releases.Items {
		if release.DeletionTimestamp == nil {
			active.Items = append(active.Items, release)
		}
	}
	return active
}
