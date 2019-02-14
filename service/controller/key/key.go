package key

import (
	"fmt"
	"strings"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	LabelApp                = "app"
	LabelAppOperatorVersion = "app-operator.giantswarm.io/version"
	LabelManagedBy          = "giantswarm.io/managed-by"
	LabelOrganization       = "giantswarm.io/organization"
	LabelReleaseVersion     = "release.giantswarm.io/version"
	LabelServiceType        = "giantswarm.io/service-type"
)

const (
	ProjectName        = "release-operator"
	OrganizationName   = "giantswarm"
	ServiceTypeManaged = "managed"
)

// ReleaseAppCRName returns the name of the release App CR for the given release cycle.
func ReleaseAppCRName(releaseCycleCR releasev1alpha1.ReleaseCycle) string {
	return ReleasePrefix(releaseCycleCR.GetName())
}

// ReleasePrefix adds release- prefix to name.
func ReleasePrefix(name string) string {
	return fmt.Sprintf("release-%s", name)
}

// ReleaseVersion returns the version of the given release.
func ReleaseVersion(releaseCR releasev1alpha1.Release) string {
	return releaseCR.Spec.Version
}

// SplitReleaseName splits a release name into provider and version.
// It returns provider, version, and error, in this order.
//
// It expects name to be in the following format <provider>.<version>
// e.g. aws.v6.0.1
func SplitReleaseName(name string) (string, string, error) {
	split := strings.SplitN(name, ".", 2)
	if len(split) < 2 {
		return "", "", microerror.Maskf(invalidReleaseNameError, "expect <provider>.<version>, got %#q", name)
	}

	return split[0], split[1], nil
}

// ToAppCR converts v into an App CR.
func ToAppCR(v interface{}) (*applicationv1alpha1.App, error) {
	appCR, ok := v.(*applicationv1alpha1.App)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &applicationv1alpha1.App{}, v)
	}

	return appCR, nil
}

// ToReleaseCycleCR converts v into a ReleaseCycle CR.
func ToReleaseCycleCR(v interface{}) (releasev1alpha1.ReleaseCycle, error) {
	releaseCycleCR, ok := v.(*releasev1alpha1.ReleaseCycle)
	if !ok {
		return releasev1alpha1.ReleaseCycle{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &releasev1alpha1.ReleaseCycle{}, v)
	}

	return *releaseCycleCR, nil
}
