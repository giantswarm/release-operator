package key

import (
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

func ReleaseVersion(customResource releasev1alpha1.Release) string {
	return customResource.Spec.Version
}

func ToAppCR(v interface{}) (*applicationv1alpha1.App, error) {
	appCR, ok := v.(*applicationv1alpha1.App)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", applicationv1alpha1.App{}, v)
	}

	return appCR, nil
}

func ToReleaseCycleCR(v interface{}) (releasev1alpha1.ReleaseCycle, error) {
	releaseCycleCR, ok := v.(*releasev1alpha1.ReleaseCycle)
	if !ok {
		return releasev1alpha1.ReleaseCycle{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &releasev1alpha1.ReleaseCycle{}, v)
	}

	return *releaseCycleCR, nil
}
