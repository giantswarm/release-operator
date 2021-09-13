package key

import (
	"fmt"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	argoappv1alpha1 "github.com/giantswarm/argoapp/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/argoapp/pkg/argoapp"
	"github.com/giantswarm/microerror"
)

const (
	AppStatusDeployed = "deployed"

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

func BuildAppName(component releasev1alpha1.ReleaseSpecComponent) string {
	return fmt.Sprintf("%s-%s", component.Name, component.Version)
}

func ComponentToArgoApplication(component releasev1alpha1.ReleaseSpecComponent) (app *argoappv1alpha1.Application, err error) {
	configRef := "main"
	if component.ConfigReference != "" {
		configRef = component.ConfigReference
	}

	ac := argoapp.ApplicationConfig{
		Name:                    BuildAppName(component),
		AppName:                 component.Name,
		AppVersion:              GetComponentRef(component),
		AppCatalog:              component.Catalog,
		AppDestinationNamespace: Namespace,
		ConfigRef:               configRef,
	}

	app, err = argoapp.NewApplication(ac)
	if err != nil {
		return app, microerror.Mask(err)
	}

	return app, nil
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
		if release.Spec.State == releasev1alpha1.StateDeprecated && !release.Status.InUse {
			// skip
		} else {
			active.Items = append(active.Items, release)
		}
	}

	return active
}

// ExtractComponents extracts the components that this operator is responsible for.
func ExtractComponents(releases releasev1alpha1.ReleaseList) []releasev1alpha1.ReleaseSpecComponent {
	var componentDedupe = map[string]bool{}
	var componentSlice = []releasev1alpha1.ReleaseSpecComponent{}

	for _, release := range releases.Items {
		for _, component := range release.Spec.Components {
			_, ok := componentDedupe[BuildAppName(component)]
			if component.ReleaseOperatorDeploy && !ok {
				componentDedupe[BuildAppName(component)] = true
				componentSlice = append(componentSlice, component)
			}
		}
	}

	return componentSlice
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

func IsSameApp(component releasev1alpha1.ReleaseSpecComponent, app applicationv1alpha1.App) bool {
	return BuildAppName(component) == app.Name &&
		component.Catalog == app.Spec.Catalog &&
		GetComponentRef(component) == app.Spec.Version
}

func ComponentAppDeployed(component releasev1alpha1.ReleaseSpecComponent, apps []applicationv1alpha1.App) bool {
	for _, a := range apps {
		if IsSameApp(component, a) && a.Status.Release.Status == AppStatusDeployed {
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
