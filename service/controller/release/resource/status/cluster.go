package status

import (
	"context"

	apiexlabels "github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/release-operator/service/controller/key"
)

type tenantCluster struct {
	ID               string
	OperatorVersion  string
	ProviderOperator string
	ReleaseVersion   string
}

// Takes a list of tenant clusters and returns two maps containing the versions of their release and operator versions.
func consolidateClusterVersions(clusters []tenantCluster) (releaseVersions map[string]bool, operatorVersions map[string]map[string]bool) {
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

// Returns a list of tenant clusters currently running on the installation.
func (r *Resource) getCurrentTenantClusters(ctx context.Context) ([]tenantCluster, error) {
	tcGetters := []func(context.Context) ([]tenantCluster, error){
		r.getCurrentAWSClusters,
		r.getCurrentAzureClusters,
		r.getLegacyAWSClusters,
		r.getLegacyAzureClusters,
		r.getLegacyKVMClusters,
	}

	var tenantClusters []tenantCluster
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
func (r *Resource) getCurrentAWSClusters(ctx context.Context) ([]tenantCluster, error) {
	awsClusters := apiv1alpha2.ClusterList{}
	err := r.k8sClient.CtrlClient().List(ctx, &awsClusters)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []tenantCluster
	for _, cluster := range awsClusters.Items {
		c := tenantCluster{
			ID:               cluster.Name,
			OperatorVersion:  cluster.Labels[apiexlabels.AWSOperatorVersion],
			ProviderOperator: key.ProviderOperatorAWS,
			ReleaseVersion:   cluster.Labels[apiexlabels.ReleaseVersion],
		}
		clusters = append(clusters, c)
	}

	return clusters, nil
}

// Returns a list of Azure clusters according to the azurecluster resource.
func (r *Resource) getCurrentAzureClusters(ctx context.Context) ([]tenantCluster, error) {
	azureClusters := apiv1alpha2.ClusterList{}
	err := r.k8sClient.CtrlClient().List(ctx, &azureClusters)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []tenantCluster
	for _, cluster := range azureClusters.Items {
		c := tenantCluster{
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
func (r *Resource) getLegacyAWSClusters(ctx context.Context) ([]tenantCluster, error) {
	awsconfigs, err := r.k8sClient.G8sClient().ProviderV1alpha1().AWSConfigs("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []tenantCluster
	for _, cluster := range awsconfigs.Items {
		c := tenantCluster{
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
func (r *Resource) getLegacyAzureClusters(ctx context.Context) ([]tenantCluster, error) {
	azureconfigs, err := r.k8sClient.G8sClient().ProviderV1alpha1().AzureConfigs("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []tenantCluster
	for _, cluster := range azureconfigs.Items {
		c := tenantCluster{
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
func (r *Resource) getLegacyKVMClusters(ctx context.Context) ([]tenantCluster, error) {
	kvmconfigs, err := r.k8sClient.G8sClient().ProviderV1alpha1().KVMConfigs("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []tenantCluster
	for _, cluster := range kvmconfigs.Items {
		c := tenantCluster{
			ID:               cluster.Name,
			OperatorVersion:  cluster.Labels[apiexlabels.KVMOperatorVersion],
			ProviderOperator: key.ProviderOperatorKVM,
			ReleaseVersion:   cluster.Labels[apiexlabels.ReleaseVersion],
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}
