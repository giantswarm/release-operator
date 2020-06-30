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
	// AppCatalog is the name of the app catalog where releases and release
	// components are stored.
	AppCatalog = "control-plane-catalog"

	AppStatusDeployed = "DEPLOYED"

	// Namespace is the namespace where App CRs are created.
	Namespace = "giantswarm"

	LabelAppOperatorVersion = "app-operator.giantswarm.io/version"
	LabelManagedBy          = "giantswarm.io/managed-by"
	LabelServiceType        = "giantswarm.io/service-type"

	ValueServiceTypeManaged = "managed"
)

func AppReferenced(app applicationv1alpha1.App, operators map[string]releasev1alpha1.ReleaseSpecComponent) bool {
	if operator, ok := operators[app.Name]; ok && GetOperatorRef(operator) == app.Spec.Version {
		return true
	}

	return false
}

func BuildAppName(operator releasev1alpha1.ReleaseSpecComponent) string {
	return fmt.Sprintf("%s-%s-hackathon", operator.Name, operator.Version)
}

func ConstructApp(operator releasev1alpha1.ReleaseSpecComponent) applicationv1alpha1.App {
	return applicationv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BuildAppName(operator),
			Namespace: Namespace,
			Labels: map[string]string{
				// TALK to team batman to find correct version!
				LabelAppOperatorVersion: "0.0.0",
				LabelManagedBy:          project.Name(),
			},
		},
		Spec: applicationv1alpha1.AppSpec{
			Catalog: AppCatalog,
			KubeConfig: applicationv1alpha1.AppSpecKubeConfig{
				InCluster: true,
			},
			Name:      operator.Name,
			Namespace: Namespace,
			Version:   GetOperatorRef(operator),
		},
	}
}

func ExtractAllOperators(releases releasev1alpha1.ReleaseList) map[string]releasev1alpha1.ReleaseSpecComponent {
	var operators = make(map[string]releasev1alpha1.ReleaseSpecComponent)

	for _, release := range releases.Items {
		for _, component := range release.Spec.Components {
			if IsOperator(component) && (operators[BuildAppName(component)] == releasev1alpha1.ReleaseSpecComponent{}) {
				operators[BuildAppName(component)] = component
			}
		}
	}
	return operators
}

func ExtractOperators(comps []releasev1alpha1.ReleaseSpecComponent) []releasev1alpha1.ReleaseSpecComponent {
	var operators []releasev1alpha1.ReleaseSpecComponent
	for _, c := range comps {
		if IsOperator(c) {
			operators = append(operators, c)
		}
	}
	return operators
}

func GetOperatorRef(comp releasev1alpha1.ReleaseSpecComponent) string {
	if comp.Reference != "" {
		return comp.Reference
	}
	return comp.Version
}

func IsOperator(component releasev1alpha1.ReleaseSpecComponent) bool {
	return strings.HasSuffix(component.Name, "-operator") && component.Name != "chart-operator" && component.Name != "app-operator"
}

func OperatorCreated(operator releasev1alpha1.ReleaseSpecComponent, apps []applicationv1alpha1.App) bool {
	for _, a := range apps {
		if a.Name == BuildAppName(operator) && a.Spec.Version == GetOperatorRef(operator) {
			return true
		}
	}

	return false
}

func OperatorDeployed(operator releasev1alpha1.ReleaseSpecComponent, apps []applicationv1alpha1.App) bool {
	for _, a := range apps {
		if a.Name == BuildAppName(operator) && a.Spec.Version == GetOperatorRef(operator) && a.Status.Release.Status == AppStatusDeployed {
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
