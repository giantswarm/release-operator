package key

import (
	"fmt"

	applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	corev1alpha1 "github.com/giantswarm/config-controller/api/v1alpha1"
	apiexlabels "github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	releasev1alpha1 "github.com/giantswarm/release-operator/v2/api/v1alpha1"
	"github.com/giantswarm/release-operator/v2/pkg/project"
)

const (
	AppStatusDeployed = "deployed"

	// ReconcileDeprecatedReleaseAnnotation makes a Release to never be skipped, even though is deprecated or not used.
	ReconcileDeprecatedReleaseAnnotation = "release-operator.giantswarm.io/reconcile-deprecated"

	// Namespace is the namespace where App CRs are created.
	Namespace = "giantswarm"

	LabelAppOperatorVersion = "app-operator.giantswarm.io/version"
	LabelManagedBy          = "giantswarm.io/managed-by"
	LabelServiceType        = "giantswarm.io/service-type"

	ValueServiceTypeManaged = "managed"
)

const (
	ProviderOperatorAWS   = "aws-operator"
	ProviderOperatorAzure = "azure-operator"
	ProviderOperatorKVM   = "kvm-operator"
)

func AppReferenced(app applicationv1alpha1.App, components map[string]releasev1alpha1.ReleaseSpecComponent) bool {
	component, ok := components[app.Name]
	if ok && IsSameApp(component, app) {
		return true
	}

	return false
}

func ConfigReferenced(config corev1alpha1.Config, components map[string]releasev1alpha1.ReleaseSpecComponent) bool {
	component, ok := components[config.Name]
	if ok && IsSameConfig(component, config) {
		return true
	}

	return false
}

func BuildAppName(component releasev1alpha1.ReleaseSpecComponent) string {
	return fmt.Sprintf("%s-%s", component.Name, component.Version)
}

func BuildConfigName(component releasev1alpha1.ReleaseSpecComponent) string {
	return fmt.Sprintf("%s-%s", component.Name, component.Version)
}

func ConstructApp(component releasev1alpha1.ReleaseSpecComponent) applicationv1alpha1.App {
	return applicationv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BuildAppName(component),
			Namespace: Namespace,
			Labels: map[string]string{
				LabelAppOperatorVersion: "0.0.0",
				LabelManagedBy:          project.Name(),
			},
		},
		Spec: applicationv1alpha1.AppSpec{
			Catalog: component.Catalog,
			KubeConfig: applicationv1alpha1.AppSpecKubeConfig{
				InCluster: true,
			},
			Name:      component.Name,
			Namespace: Namespace,
			Version:   GetComponentRef(component),
		},
	}
}

func ConstructConfig(component releasev1alpha1.ReleaseSpecComponent) corev1alpha1.Config {
	return corev1alpha1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BuildConfigName(component),
			Namespace: Namespace,
			Labels: map[string]string{
				apiexlabels.ConfigControllerVersion: "0.0.0",
				LabelManagedBy:                      project.Name(),
			},
		},
		Spec: corev1alpha1.ConfigSpec{
			App: corev1alpha1.ConfigSpecApp{
				Catalog: component.Catalog,
				Name:    component.Name,
				Version: GetComponentRef(component),
			},
		},
	}
}

func ExcludeDeletedRelease(releases releasev1alpha1.ReleaseList) releasev1alpha1.ReleaseList {
	var active releasev1alpha1.ReleaseList
	for _, release := range releases.Items {
		if release.DeletionTimestamp == nil {
			active.Items = append(active.Items, release)
		}
	}
	return active
}

func ExcludeUnusedDeprecatedReleases(releases releasev1alpha1.ReleaseList) releasev1alpha1.ReleaseList {
	var active releasev1alpha1.ReleaseList

	for _, release := range releases.Items {

		if release.Spec.State == releasev1alpha1.StateDeprecated && !release.Status.InUse && release.Annotations[ReconcileDeprecatedReleaseAnnotation] == "" {
			// skip
		} else {
			active.Items = append(active.Items, release)
		}
	}

	return active
}

// ExtractComponents extracts the components that this operator is responsible for.
func ExtractComponents(releases releasev1alpha1.ReleaseList) map[string]releasev1alpha1.ReleaseSpecComponent {
	var components = make(map[string]releasev1alpha1.ReleaseSpecComponent)

	for _, release := range releases.Items {
		for _, component := range release.Spec.Components {
			if component.ReleaseOperatorDeploy && (components[BuildAppName(component)] == releasev1alpha1.ReleaseSpecComponent{}) {
				components[BuildAppName(component)] = component
			}
		}
	}
	return components
}

// FilterComponents filters the components that this operator is responsible for.
func FilterComponents(comps []releasev1alpha1.ReleaseSpecComponent) []releasev1alpha1.ReleaseSpecComponent {
	var filteredComponents []releasev1alpha1.ReleaseSpecComponent
	for _, c := range comps {
		if c.ReleaseOperatorDeploy {
			filteredComponents = append(filteredComponents, c)
		}
	}
	return filteredComponents
}

func GetComponentRef(comp releasev1alpha1.ReleaseSpecComponent) string {
	if comp.Reference != "" {
		return comp.Reference
	}
	return comp.Version
}

func GetProviderOperators() []string {
	return []string{ProviderOperatorAWS, ProviderOperatorAzure, ProviderOperatorKVM}
}

func GetAppConfig(app applicationv1alpha1.App, configs corev1alpha1.ConfigList) (
	appConfig corev1alpha1.ConfigStatusConfig) {

	for _, config := range configs.Items {
		configManagedByLabel, configIsManagedByReleaseOperator := config.Labels[LabelManagedBy]

		matches := true
		matches = matches && app.Name == config.Name
		matches = matches && app.Spec.Name == config.Status.App.Name
		matches = matches && app.Spec.Version == config.Status.App.Version
		matches = matches && app.Spec.Catalog == config.Status.App.Catalog
		matches = matches && configIsManagedByReleaseOperator
		matches = matches && configManagedByLabel == project.Name()

		if matches {

			appConfig.ConfigMapRef = config.Status.Config.ConfigMapRef
			appConfig.SecretRef = config.Status.Config.SecretRef
			break
		}
	}

	return appConfig
}

func IsSameApp(component releasev1alpha1.ReleaseSpecComponent, app applicationv1alpha1.App) bool {
	return BuildAppName(component) == app.Name &&
		component.Catalog == app.Spec.Catalog &&
		GetComponentRef(component) == app.Spec.Version
}

func IsSameConfig(component releasev1alpha1.ReleaseSpecComponent, config corev1alpha1.Config) bool {
	configManagedByLabel, configIsManagedByReleaseOperator := config.Labels[LabelManagedBy]
	return component.Name == config.Spec.App.Name &&
		component.Catalog == config.Spec.App.Catalog &&
		GetComponentRef(component) == config.Spec.App.Version &&
		configIsManagedByReleaseOperator &&
		configManagedByLabel == project.Name()
}

func ComponentAppCreated(component releasev1alpha1.ReleaseSpecComponent, apps []applicationv1alpha1.App) bool {
	for _, a := range apps {
		if IsSameApp(component, a) {
			return true
		}
	}

	return false
}

func ComponentAppDeployed(component releasev1alpha1.ReleaseSpecComponent, apps []applicationv1alpha1.App) bool {
	for _, a := range apps {
		if IsSameApp(component, a) && a.Status.Release.Status == AppStatusDeployed {
			return true
		}
	}

	return false
}

func ComponentConfigCreated(component releasev1alpha1.ReleaseSpecComponent, configs []corev1alpha1.Config) bool {
	for _, c := range configs {
		if IsSameConfig(component, c) {
			return true
		}
	}

	return false
}

// ToReleaseCR converts v into a Release CR.
func ToReleaseCR(v interface{}) (*releasev1alpha1.Release, error) {
	x, ok := v.(*releasev1alpha1.Release)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", x, v)
	}

	return x, nil
}
