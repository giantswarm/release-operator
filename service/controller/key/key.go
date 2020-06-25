package key

import (
	"fmt"
	"strings"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/pkg/project"
)

const (
	// AppCatalog is the name of the app catalog where releases and release
	// components are stored.
	AppCatalog = "control-plane-catalog"

	// Namespace is the namespace where App CRs are created.
	Namespace = "giantswarm"

	LabelAppOperatorVersion = "app-operator.giantswarm.io/version"
	LabelManagedBy          = "giantswarm.io/managed-by"
	LabelServiceType        = "giantswarm.io/service-type"

	ValueServiceTypeManaged = "managed"
)

func BuildAppName(operatorName, operatorRef string) string {
	return fmt.Sprintf("%s-%s-hackathon", operatorName, operatorRef)
}

func ExtractOperators(comps []releasev1alpha1.ReleaseSpecComponent) []releasev1alpha1.ReleaseSpecComponent {
	var operators []releasev1alpha1.ReleaseSpecComponent
	for _, c := range comps {
		if strings.Contains(c.Name, "operator") {
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

func AppInApps(apps []applicationv1alpha1.App, appName string, appVersion string) bool {
	for _, a := range apps {
		if a.Name == appName && a.Spec.Version == appVersion {
			return true
		}
	}

	return false
}

func AppInOperators(operators []releasev1alpha1.ReleaseSpecComponent, app applicationv1alpha1.App) bool {
	for _, operator := range operators {
		if BuildAppName(operator.Name, GetOperatorRef(operator)) == app.Name && GetOperatorRef(operator) == app.Spec.Version {
			return true
		}
	}

	return false
}

func ConstructApp(operatorName, operatorRef string) applicationv1alpha1.App {
	return applicationv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      BuildAppName(operatorName, operatorRef),
			Namespace: Namespace,
			Labels: map[string]string{
				// TALK to team batman to find correct version!
				LabelAppOperatorVersion: "0.0.0",
				LabelManagedBy:          project.Name(),
			},
		},
		Spec: applicationv1alpha1.AppSpec{
			Catalog:   AppCatalog,
			Name:      operatorName,
			Namespace: Namespace,
			Version:   operatorRef,
		},
	}
}
