package status

import (
	"context"

	apiexlabels "github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/release-operator/v2/service/controller/key"
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
	awsClusters, err := r.listPartialObjectMetadata(ctx, metav1.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Version: "v1alpha2",
		Kind:    "AWSCluster",
	}, "default")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []tenantCluster
	for _, cluster := range awsClusters {
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
	azureClusters, err := r.listPartialObjectMetadata(ctx, metav1.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Version: "v1alpha3",
		Kind:    "AzureCluster",
	}, "default")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []tenantCluster
	for _, cluster := range azureClusters {
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
	configs, err := r.listPartialObjectMetadata(ctx, metav1.GroupVersionKind{
		Group:   "provider.giantswarm.io",
		Version: "v1alpha1",
		Kind:    "AWSConfig",
	}, "default")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []tenantCluster
	for _, cluster := range configs {
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
	configs, err := r.listPartialObjectMetadata(ctx, metav1.GroupVersionKind{
		Group:   "provider.giantswarm.io",
		Version: "v1alpha1",
		Kind:    "AzureConfig",
	}, "default")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []tenantCluster
	for _, cluster := range configs {
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
	configs, err := r.listPartialObjectMetadata(ctx, metav1.GroupVersionKind{
		Group:   "provider.giantswarm.io",
		Version: "v1alpha1",
		Kind:    "KVMConfig",
	}, "default")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []tenantCluster
	for _, cluster := range configs {
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

func (r *Resource) listPartialObjectMetadata(ctx context.Context, gvk metav1.GroupVersionKind, namespace string) ([]metav1.PartialObjectMetadata, error) {
	results := metav1.PartialObjectMetadataList{
		TypeMeta: metav1.TypeMeta{
			Kind: gvk.Kind,
			APIVersion: metav1.GroupVersion{
				Group:   gvk.Group,
				Version: gvk.Version,
			}.String(),
		},
	}
	err := r.k8sClient.CtrlClient().List(ctx, &results, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return results.Items, nil
}
