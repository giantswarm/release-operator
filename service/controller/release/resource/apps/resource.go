package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// "github.com/giantswarm/api/legacycluster"

	appv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	apiexlabels "github.com/giantswarm/apiextensions/pkg/label"

	// gsclient "github.com/giantswarm/gsclientgen/v2/client"
	// "github.com/giantswarm/gsclientgen/v2/client/clusters"
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

	// var tenantClusters
	// {
	//    tenantClusters, err := r.getCurrentTenantClusters()
	//    if err != nil {
	// 	     return microerror.Mask(err) // Might be better to proceed here instead of aborting
	//    }
	// }
	// releases = excludeUnusedDeprecatedReleases(releases, tenantClusters)

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

func getPossibleReleasesForOperator(operator string, version string, releases releasev1alpha1.ReleaseList) releasev1alpha1.ReleaseList {
	var possibleReleases releasev1alpha1.ReleaseList
	for _, release := range releases.Items {
		for _, c := range release.Spec.Components {
			if c.Name == operator && c.Version == version {
				possibleReleases.Items = append(possibleReleases.Items, release)
			}
		}
	}
	return possibleReleases
}

type TenantCluster struct {
	ID               string
	Version          string
	PossibleVersions []string
	Labels           []string
}

func (r *Resource) excludeUnusedDeprecatedReleases(releases releasev1alpha1.ReleaseList, clusters []string) (releasev1alpha1.ReleaseList, error) {
	// Go over releases
	// If deprecated, AND no existing cluster with this version, exclude it
	// Cluster with this version:
	//    - has release label with this version or
	//    - if no release label, check operator label
	//        - with operator label, find possible releases

	for _, release := range releases.Items {
		if release.Spec.State == "deprecated" { // TODO: Should make this constant public in apiextensions

		}
	}

	var active releasev1alpha1.ReleaseList
	fmt.Println(apiexlabels.AWSOperatorVersion)
	return active, nil
}

// func getPotentialClusterVersions(cluster TenantCluster, releases releasev1alpha1.ReleaseList) []string {
// 	if cluster.Version != "" {
// 		return []string{cluster.Version}
// 	}

// 	operatorLabels := []string{apiexlabels.AWSOperatorVersion, apiexlabels.AzureOperatorVersion}

// 	var versions []string
// 	for _, label := range operatorLabels {
// 		if cluster.Labels[label] != nil {
// 			versions = append(versions, getPossibleReleasesForOperator())
// 		}
// 	}
// 	return versions
// }

func (r *Resource) getCurrentTenantClusters(ctx context.Context) {

	// legacycluster.List()

	// params := clusters.NewGetClustersParams()
	// gsclient.New()

	legacyClusters, err := r.getLegacyClusters(ctx)

}

func (r *Resource) getCurrentAWSClusters() ([]Cluster, error) {
	awsclusters, err := r.k8sClient.G8sClient().InfrastructureV1alpha2().AWSClusters("default").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}
	var clusterIDs []string
	for _, cluster := range awsclusters.Items {
		clusterIDs = append(clusterIDs, cluster.Name)
		// releaseVersion := cluster.Labels[v2labels.ReleaseVersion]
	}

	return []Cluster{}, nil
}

func getCurrentAzureClusters() {

}

func getCurrentKVMClusters() {

}

type Cluster struct {
	CreateDate     time.Time `json:"create_date"`
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Owner          string    `json:"owner"`
	ReleaseVersion string    `json:"release_version"`
}

func (r *Resource) getLegacyClusters(ctx context.Context, orgIDs []string) ([]Cluster, error) {
	storageConfig, err := r.k8sClient.G8sClient().CoreV1alpha1().StorageConfigs("giantswarm").Get("cluster-service", metav1.GetOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterIDs := []string{}
	for _, orgID := range orgIDs {
		key := "/owner/organization/" + orgID
		keyLen := len(key)

		for k := range storageConfig.Spec.Storage.Data {
			if len(k) <= keyLen+1 {
				continue
			}
			if !strings.HasPrefix(k, key) {
				continue
			}

			if k[keyLen] != '/' {
				continue
			}

			clusterID := k[keyLen+1:]
			clusterIDs = append(clusterIDs, clusterID)
		}
	}

	clusters := []Cluster{}
	for _, clusterID := range clusterIDs {
		key := "/cluster/" + clusterID
		value, ok := storageConfig.Spec.Storage.Data[key]
		if !ok {
			return nil, microerror.Maskf(notFoundError, "no value for key '%s'", key)
		}

		var cluster Cluster
		err = json.Unmarshal([]byte(value), &cluster)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		clusters = append(clusters, cluster)
	}

	return clusters, nil

}

// func (r *Resource) excludeUnusedDeprecatedReleases(releases releasev1alpha1.ReleaseList) (releasev1alpha1.ReleaseList, error) {

// 	var clusterIDs []string
// 	{

// 		awsclusters, err := r.k8sClient.G8sClient().InfrastructureV1alpha2().AWSClusters("default").List(metav1.ListOptions{})
// 		if err != nil {
// 			return nil, microerror.Mask(err)
// 		}

// 		for _, cluster := range awsclusters.Items {
// 			clusterIDs = append(clusterIDs, cluster.Name)
// 			releaseVersion := cluster.Labels[v2labels.ReleaseVersion]
// 		}

// 		awsconfigs, err := r.k8sClient.G8sClient().ProviderV1alpha1().AWSConfigs("default").List(metav1.ListOptions{})
// 		if err != nil {
// 			return nil, microerror.Mask(err)
// 		}

// 		for _, cluster := range awsconfigs.Items {
// 			clusterIDs = append(clusterIDs, cluster.Name)
// 			operatorVersion := cluster.Labels[v2labels.AWSOperatorVersion]
// 		}

// 	}

// 	var active releasev1alpha1.ReleaseList
// 	for _, release := range releases.Items {
// 		if release.DeletionTimestamp == nil { // TODO: Change to check usage
// 			active.Items = append(active.Items, release)
// 		}
// 	}
// 	return active
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
