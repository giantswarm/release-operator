package key

import (
	"fmt"
	"strings"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/pkg/project"
)

const (
	AppStatusDeployed = "DEPLOYED"

	// Namespace is the namespace where App CRs are created.
	Namespace = "giantswarm"

	LabelAppOperatorVersion = "app-operator.giantswarm.io/version"
	LabelManagedBy          = "giantswarm.io/managed-by"
	LabelServiceType        = "giantswarm.io/service-type"

	ValueServiceTypeManaged = "managed"
)

func AppReferenced(app applicationv1alpha1.App, components map[string]releasev1alpha1.ReleaseSpecComponent) bool {
	component, ok := components[app.Name]
	if ok && IsSameApp(component, app) {
		return true
	}

	return false
}

func BuildAppName(component releasev1alpha1.ReleaseSpecComponent) string {
	return fmt.Sprintf("%s-%s-hackathon", component.Name, component.Version)
}

func ConstructApp(component releasev1alpha1.ReleaseSpecComponent) applicationv1alpha1.App {
	return applicationv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BuildAppName(component),
			Namespace: Namespace,
			Labels: map[string]string{
				// TALK to team batman to find correct version!
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

func ExtractAllRelevantComponents(releases releasev1alpha1.ReleaseList) map[string]releasev1alpha1.ReleaseSpecComponent {
	var relevantComponents = make(map[string]releasev1alpha1.ReleaseSpecComponent)

	for _, release := range releases.Items {
		for _, component := range release.Spec.Components {
			if IsOperator(component) && (relevantComponents[BuildAppName(component)] == releasev1alpha1.ReleaseSpecComponent{}) {
				relevantComponents[BuildAppName(component)] = component
			}
		}
	}
	return relevantComponents
}

func ExtractRelevantComponents(comps []releasev1alpha1.ReleaseSpecComponent) []releasev1alpha1.ReleaseSpecComponent {
	var relevantComponents []releasev1alpha1.ReleaseSpecComponent
	for _, c := range comps {
		if IsOperator(c) {
			relevantComponents = append(relevantComponents, c)
		}
	}
	return relevantComponents
}

func GetComponentRef(comp releasev1alpha1.ReleaseSpecComponent) string {
	if comp.Reference != "" {
		return comp.Reference
	}
	return comp.Version
}

func IsOperator(component releasev1alpha1.ReleaseSpecComponent) bool {
	return strings.HasSuffix(component.Name, "-operator") && component.Name != "chart-operator" && component.Name != "app-operator"
}

func IsSameApp(component releasev1alpha1.ReleaseSpecComponent, app applicationv1alpha1.App) bool {
	return BuildAppName(component) == app.Name &&
		component.Catalog == app.Spec.Catalog &&
		GetComponentRef(component) == app.Spec.Version
}

func ComponentCreated(component releasev1alpha1.ReleaseSpecComponent, apps []applicationv1alpha1.App) bool {
	for _, a := range apps {
		if IsSameApp(component, a) {
			return true
		}
	}

	return false
}

func ComponentDeployed(component releasev1alpha1.ReleaseSpecComponent, apps []applicationv1alpha1.App) bool {
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
