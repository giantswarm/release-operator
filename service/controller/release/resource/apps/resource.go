package apps

import (
	"context"
	"fmt"

	appv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	apiexlabels "github.com/giantswarm/apiextensions/pkg/label"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	var tenantClusters []TenantCluster
	{
		var err error
		tenantClusters, err = r.getCurrentTenantClusters()
		if err != nil {
			return microerror.Mask(err) // Might be better to proceed here instead of aborting
		}
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d running tenant clusters", len(tenantClusters)))
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

type TenantCluster struct {
	ID               string
	ReleaseVersion   string
	PossibleVersions []string
	Labels           []string
	OperatorVersion  string
	Provider         string
}

func (r *Resource) excludeUnusedDeprecatedReleases(releases releasev1alpha1.ReleaseList, clusters []TenantCluster) (releasev1alpha1.ReleaseList, error) {
	// Go over releases
	// If deprecated, AND no existing cluster with this version, exclude it
	// Cluster with this version:
	//    - has release label with this version or
	//    - if no release label, check operator label
	//        - with operator label, find possible releases

	releaseVersions, operatorVersions := consolidateClusterVersions(clusters)
	var active releasev1alpha1.ReleaseList
	for _, release := range releases.Items {
		if release.Spec.State == "deprecated" { // TODO: Should make this constant public in apiextensions
			// check set of release versions -- if present ,keep
			if releaseVersions[release.Name] {
				active.Items = append(active.Items, release)
			} else {
				operatorVersion := getOperatorVersionInRelease("aws-operator", release)
				// check set of operator versions -- if present ,keep
				if operatorVersion != "" && operatorVersions[operatorVersion] {
					active.Items = append(active.Items, release)
				}
			}
		}
	}

	fmt.Println(apiexlabels.AWSOperatorVersion)
	return active, nil
}

func getOperatorVersionInRelease(operator string, release releasev1alpha1.Release) string {
	for _, component := range release.Spec.Components {
		if component.Name == operator { // TODO: parameterize the operator version or check all
			return component.Version
		}
	}
	return ""
}

func consolidateClusterVersions(clusters []TenantCluster) (map[string]bool, map[string]bool) {
	releaseVersions := make(map[string]bool)
	operatorVersions := make(map[string]bool)

	for _, c := range clusters {
		releaseVersions[c.ReleaseVersion] = true
		operatorVersions[c.OperatorVersion] = true
	}

	return releaseVersions, operatorVersions
}

func (r *Resource) getCurrentTenantClusters() ([]TenantCluster, error) {

	var tenantClusters []TenantCluster
	{
		awsClusters, err := r.getCurrentAWSClusters()
		if err != nil {
			return nil, microerror.Mask(err)
		}
		tenantClusters = append(tenantClusters, awsClusters...)
		r.logger.Log("level", "debug", "message", fmt.Sprintf("found %d aws tenant clusters", len(awsClusters)))

		legacyClusters, err := r.getLegacyClusters()
		if err != nil {
			return nil, microerror.Mask(err)
		}
		r.logger.Log("level", "debug", "message", fmt.Sprintf("found %d legacy tenant clusters", len(legacyClusters)))
		tenantClusters = append(tenantClusters, legacyClusters...)
	}

	return tenantClusters, nil
}

func (r *Resource) getCurrentAWSClusters() ([]TenantCluster, error) {
	awsclusters, err := r.k8sClient.G8sClient().InfrastructureV1alpha2().AWSClusters("default").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []TenantCluster
	for _, cluster := range awsclusters.Items {
		c := TenantCluster{
			ID:              cluster.Name,
			OperatorVersion: cluster.Labels[apiexlabels.AWSOperatorVersion],
			ReleaseVersion:  cluster.Labels[apiexlabels.ReleaseVersion],
			Provider:        "aws", // TODO: Parameterize or detect
		}
		clusters = append(clusters, c)
	}

	return clusters, nil
}

func (r *Resource) getLegacyClusters() ([]TenantCluster, error) {
	var legacyClusters []TenantCluster
	aws, err := r.getLegacyAWSClusters()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	legacyClusters = append(legacyClusters, aws...)

	// Same for Azure and KVM

	return legacyClusters, nil
}

func (r *Resource) getLegacyAWSClusters() ([]TenantCluster, error) {
	awsconfigs, err := r.k8sClient.G8sClient().ProviderV1alpha1().AWSConfigs("default").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []TenantCluster
	for _, cluster := range awsconfigs.Items {
		c := TenantCluster{
			ID:              cluster.Name,
			OperatorVersion: cluster.Labels[apiexlabels.AWSOperatorVersion],
			Provider:        "aws",
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}

// func getCurrentAzureClusters() {

// }

// func getCurrentKVMClusters() {

// }

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
