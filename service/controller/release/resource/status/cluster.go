package status

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	apiexlabels "github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	azurecapi "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	"github.com/giantswarm/release-operator/service/controller/key"
)

type TenantCluster struct {
	ID               string
	OperatorVersion  string
	ProviderOperator string
	ReleaseVersion   string
}

// Takes a list of tenant clusters and returns two maps containing the versions of their release and operator versions.
func consolidateClusterVersions(clusters []TenantCluster) (releaseVersions map[string]bool, operatorVersions map[string]map[string]bool) {
	releaseVersions = make(map[string]bool)
	operatorVersions = make(map[string]map[string]bool)

	// operatorVersions is a nested map including the operator name and version
	// e.g. operatorVersions["aws-operator"]["8.7.6"]:true

	for _, c := range clusters {
		releaseVersions[c.ReleaseVersion] = true

		if operatorVersions[c.ProviderOperator] == nil {
			operatorVersions[c.ProviderOperator] = make(map[string]bool)
		}
		operatorVersions[c.ProviderOperator][c.OperatorVersion] = true
	}

	return
}

// E 09/16 09:58:15 /apis/release.giantswarm.io/v1alpha1/releases/v11.5.1 status error finding tenant clusters: no matches for kind "AzureCluster" in version "infrastructure.cluster.x-k8s.io/v1alpha3" | release-operator/service/controller/release/resource/status/create.go:54 | controller=release-operator-release | event=update | loop=73 | version=46222030
// Returns a list of tenant clusters currently running on the installation.
func (r *Resource) getCurrentTenantClusters(ctx context.Context) ([]TenantCluster, error) {
	tcGetters := []func(context.Context) ([]TenantCluster, error){
		r.getCurrentAWSClusters,
		r.getCurrentAzureClusters,
		r.getLegacyAWSClusters,
		r.getLegacyAzureClusters,
		r.getLegacyKVMClusters,
	}

	var tenantClusters []TenantCluster
	{
		for _, f := range tcGetters {
			clusters, err := f(ctx)
			if IsResourceNotFound(err) || IsNoMatchesForKind(err) {
				// Fall through
			} else if err != nil {
				return nil, microerror.Mask(err)
			}
			tenantClusters = append(tenantClusters, clusters...)
		}
	}

	return tenantClusters, nil
}

// Returns a list of AWS clusters according to the awscluster resource.
func (r *Resource) getCurrentAWSClusters(ctx context.Context) ([]TenantCluster, error) {
	awsClusters := infrastructurev1alpha2.AWSClusterList{}
	err := r.k8sClient.CtrlClient().List(ctx, &awsClusters)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []TenantCluster
	for _, cluster := range awsClusters.Items {
		c := TenantCluster{
			ID:               cluster.Name,
			OperatorVersion:  cluster.Labels[apiexlabels.AWSOperatorVersion],
			ReleaseVersion:   cluster.Labels[apiexlabels.ReleaseVersion],
			ProviderOperator: key.ProviderOperatorAWS,
		}
		clusters = append(clusters, c)
	}

	return clusters, nil
}

// Returns a list of Azure clusters according to the azurecluster resource.
func (r *Resource) getCurrentAzureClusters(ctx context.Context) ([]TenantCluster, error) {
	azureClusters := azurecapi.AzureClusterList{}
	err := r.k8sClient.CtrlClient().List(ctx, &azureClusters)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []TenantCluster
	for _, cluster := range azureClusters.Items {
		c := TenantCluster{
			ID:               cluster.Name,
			OperatorVersion:  cluster.Labels[apiexlabels.AzureOperatorVersion],
			ProviderOperator: key.ProviderOperatorAzure,
			ReleaseVersion:   cluster.Labels[apiexlabels.ReleaseVersion],
		}
		clusters = append(clusters, c)
	}

	return clusters, nil
}

// Returns a list of running AWS legacy clusters based on awsconfig resources.
func (r *Resource) getLegacyAWSClusters(ctx context.Context) ([]TenantCluster, error) {
	awsconfigs, err := r.k8sClient.G8sClient().ProviderV1alpha1().AWSConfigs("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []TenantCluster
	for _, cluster := range awsconfigs.Items {
		c := TenantCluster{
			ID:               cluster.Name,
			OperatorVersion:  cluster.Labels[apiexlabels.AWSOperatorVersion],
			ProviderOperator: key.ProviderOperatorAWS,
			ReleaseVersion:   cluster.Labels[apiexlabels.ReleaseVersion],
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}

// Returns a list of running Azure legacy clusters based on azureconfig resources.
func (r *Resource) getLegacyAzureClusters(ctx context.Context) ([]TenantCluster, error) {
	azureconfigs, err := r.k8sClient.G8sClient().ProviderV1alpha1().AzureConfigs("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []TenantCluster
	for _, cluster := range azureconfigs.Items {
		c := TenantCluster{
			ID:               cluster.Name,
			OperatorVersion:  cluster.Labels[apiexlabels.AzureOperatorVersion],
			ProviderOperator: key.ProviderOperatorAzure,
			ReleaseVersion:   cluster.Labels[apiexlabels.ReleaseVersion],
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}

// Returns a list of running KVM legacy clusters based on kvmconfig resources.
func (r *Resource) getLegacyKVMClusters(ctx context.Context) ([]TenantCluster, error) {
	kvmconfigs, err := r.k8sClient.G8sClient().ProviderV1alpha1().KVMConfigs("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []TenantCluster
	for _, cluster := range kvmconfigs.Items {
		c := TenantCluster{
			ID:               cluster.Name,
			OperatorVersion:  cluster.Labels[apiexlabels.KVMOperatorVersion],
			ProviderOperator: key.ProviderOperatorKVM,
			ReleaseVersion:   cluster.Labels[apiexlabels.ReleaseVersion],
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}
