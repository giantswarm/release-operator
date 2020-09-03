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

	r.logger.LogCtx(ctx, "level", "debug", "message", "searching for running tenant clusters")
	var tenantClusters []TenantCluster
	{
		var err error
		tenantClusters, err = r.getCurrentTenantClusters()
		if err != nil {
			r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("error finding tenant clusters: %s", err))
			// return microerror.Mask(err) // Might be better to proceed here instead of aborting
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

type TenantCluster struct {
	ID              string
	OperatorVersion string
	Provider        string
	ReleaseVersion  string
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
				operatorVersion := getOperatorVersionInRelease("aws-operator", release) // TODO: parameterize the operator version or check all
				// Check the set of operator versions and keep this release if its operator version is used.
				if operatorVersion != "" && operatorVersions[operatorVersion] {
					active.Items = append(active.Items, release)
					r.logger.Log("level", "debug", "message", fmt.Sprintf("keeping release %s because a cluster using its operator version (%s) is present", release.Name, operatorVersion))
				}
			}
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

// Takes a list of tenant clusters and returns two maps containing the versions of their release and operator versions.
func consolidateClusterVersions(clusters []TenantCluster) (map[string]bool, map[string]bool) {
	releaseVersions := make(map[string]bool)
	operatorVersions := make(map[string]bool)

	for _, c := range clusters {
		fmt.Printf("Cluster %s (%s) is using operator version %s\n", c.ID, c.ReleaseVersion, c.OperatorVersion)
		releaseVersions[c.ReleaseVersion] = true
		operatorVersions[c.OperatorVersion] = true
	}

	return releaseVersions, operatorVersions
}

// Returns a list of tenant clusters currently running on the installation.
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

// Returns a list of AWS clusters according to the awscluster resource (non-legacy).
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

// Returns a list of legacy clusters based on <provider>config resources.
func (r *Resource) getLegacyClusters() ([]TenantCluster, error) {
	var legacyClusters []TenantCluster
	aws, err := r.getLegacyAWSClusters()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	legacyClusters = append(legacyClusters, aws...)

	azure, err := r.getLegacyAzureClusters()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	legacyClusters = append(legacyClusters, azure...)

	// Same KVM

	return legacyClusters, nil
}

// Returns a list of running AWS legacy clusters based on awsconfig resources.
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

// Returns a list of running Azure legacy clusters based on azureconfig resources.
func (r *Resource) getLegacyAzureClusters() ([]TenantCluster, error) {
	azureconfigs, err := r.k8sClient.G8sClient().ProviderV1alpha1().AzureConfigs("default").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []TenantCluster
	for _, cluster := range azureconfigs.Items {
		c := TenantCluster{
			ID:              cluster.Name,
			OperatorVersion: cluster.Labels[apiexlabels.AzureOperatorVersion],
			Provider:        "azure",
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}

// Returns a list of running KVM legacy clusters based on kvmconfig resources.
func (r *Resource) getLegacyKVMClusters() ([]TenantCluster, error) {
	kvmconfigs, err := r.k8sClient.G8sClient().ProviderV1alpha1().KVMConfigs("default").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []TenantCluster
	for _, cluster := range kvmconfigs.Items {
		c := TenantCluster{
			ID:              cluster.Name,
			OperatorVersion: cluster.Labels["kvm-operator.giantswarm.io/version"], // TODO: Why isn't this in apiextensions?
			Provider:        "kvm",
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}

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
