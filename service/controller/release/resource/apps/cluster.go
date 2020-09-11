package apps

import (
	"context"
	"fmt"

	apiexlabels "github.com/giantswarm/apiextensions/v2/pkg/label"
	// azurecapi "github.com/giantswarm/cluster-api-provider-azure/api/v1alpha3"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	azurecapi "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"

	// "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/release-operator/service/controller/key"
)

type TenantCluster struct {
	ID               string
	OperatorVersion  string
	ProviderOperator string
	ReleaseVersion   string
}

// Takes a list of tenant clusters and returns two maps containing the versions of their release and operator versions.
func consolidateClusterVersions(clusters []TenantCluster) (map[string]bool, map[string]map[string]bool) {
	releaseVersions := make(map[string]bool)
	operatorVersions := make(map[string]map[string]bool)

	// operatorVersions is a nested map including the operator name and version
	// e.g. operatorVersions["aws-operator"]["8.7.6"]:true

	for _, c := range clusters {
		fmt.Printf("Cluster %s (%s) is using operator version %s\n", c.ID, c.ReleaseVersion, c.OperatorVersion)
		releaseVersions[c.ReleaseVersion] = true

		if operatorVersions[c.ProviderOperator] == nil {
			operatorVersions[c.ProviderOperator] = make(map[string]bool)
		}
		operatorVersions[c.ProviderOperator][c.OperatorVersion] = true
	}

	return releaseVersions, operatorVersions
}

// Returns a list of tenant clusters currently running on the installation.
func (r *Resource) getCurrentTenantClusters(ctx context.Context) ([]TenantCluster, error) {

	var tenantClusters []TenantCluster
	{
		awsClusters, err := r.getCurrentAWSClusters(ctx)
		if IsResourceNotFound(err) {
			// Fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
		tenantClusters = append(tenantClusters, awsClusters...)
		r.logger.Log("level", "debug", "message", fmt.Sprintf("found %d aws tenant clusters", len(awsClusters)))

		azureClusters, err := r.getCurrentAzureClusters(ctx)
		if IsResourceNotFound(err) || IsNoMatchesForKind(err) {
			// Fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
		tenantClusters = append(tenantClusters, azureClusters...)
		r.logger.Log("level", "debug", "message", fmt.Sprintf("found %d azure tenant clusters", len(azureClusters)))

		legacyClusters, err := r.getLegacyClusters(ctx)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		tenantClusters = append(tenantClusters, legacyClusters...)
		r.logger.Log("level", "debug", "message", fmt.Sprintf("found %d legacy tenant clusters", len(legacyClusters)))
	}

	return tenantClusters, nil
}

// Returns a list of AWS clusters according to the awscluster resource (non-legacy).
func (r *Resource) getCurrentAWSClusters(ctx context.Context) ([]TenantCluster, error) {
	awsclusters, err := r.k8sClient.G8sClient().InfrastructureV1alpha2().AWSClusters("default").List(ctx, metav1.ListOptions{})
	if IsResourceNotFound(err) {
		// Fall through
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusters []TenantCluster
	for _, cluster := range awsclusters.Items {
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

// Returns a list of legacy clusters based on <provider>config resources.
func (r *Resource) getLegacyClusters(ctx context.Context) ([]TenantCluster, error) {
	var legacyClusters []TenantCluster
	aws, err := r.getLegacyAWSClusters(ctx)
	if IsResourceNotFound(err) {
		// Fall through
	} else if err != nil {
		r.logger.Log("level", "error", "message", fmt.Sprintf("error getting aws legacy clusters: %s", err))
	}
	r.logger.Log("level", "debug", "message", fmt.Sprintf("found %d aws legacy clusters", len(aws)))
	legacyClusters = append(legacyClusters, aws...)

	azure, err := r.getLegacyAzureClusters(ctx)
	if IsResourceNotFound(err) {
		// Fall through
	} else if err != nil {
		r.logger.Log("level", "error", "message", fmt.Sprintf("error getting azure legacy clusters: %s", err))
	}
	r.logger.Log("level", "debug", "message", fmt.Sprintf("found %d azure legacy clusters", len(azure)))
	legacyClusters = append(legacyClusters, azure...)

	kvm, err := r.getLegacyKVMClusters(ctx)
	if IsResourceNotFound(err) {
		// Fall through
	} else if err != nil {
		r.logger.Log("level", "error", "message", fmt.Sprintf("error getting kvm legacy clusters: %s", err))
	}
	r.logger.Log("level", "debug", "message", fmt.Sprintf("found %d kvm legacy clusters", len(kvm)))
	legacyClusters = append(legacyClusters, kvm...)

	return legacyClusters, nil
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
			OperatorVersion:  cluster.Labels[key.LabelKVMOperator],
			ProviderOperator: key.ProviderOperatorKVM,
			ReleaseVersion:   cluster.Labels[apiexlabels.ReleaseVersion],
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}

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
