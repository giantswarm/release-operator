package apps

import (
	"fmt"

	apiexlabels "github.com/giantswarm/apiextensions/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TenantCluster struct {
	ID              string
	OperatorVersion string
	Provider        string
	ReleaseVersion  string
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
	if IsResourceNotFound(err) {
		// Fall through
	} else if err != nil {
		r.logger.Log("level", "error", "message", fmt.Sprintf("error getting aws legacy clusters: %s", err))
	}
	r.logger.Log("level", "debug", "message", fmt.Sprintf("found %d aws legacy clusters", len(aws)))
	legacyClusters = append(legacyClusters, aws...)

	azure, err := r.getLegacyAzureClusters()
	if IsResourceNotFound(err) {
		// Fall through
	} else if err != nil {
		r.logger.Log("level", "error", "message", fmt.Sprintf("error getting azure legacy clusters: %s", err))
	}
	r.logger.Log("level", "debug", "message", fmt.Sprintf("found %d azure legacy clusters", len(azure)))
	legacyClusters = append(legacyClusters, azure...)

	kvm, err := r.getLegacyKVMClusters()
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
