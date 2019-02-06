package key

import (
	releasev1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
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

func ReleaseVersion(customResource releasev1.Release) string {
	return customResource.Spec.Version
}

func ToReleaseCycleCR(v interface{}) (releasev1.ReleaseCycle, error) {
	customResourcePointer, ok := v.(*releasev1.ReleaseCycle)
	if !ok {
		return releasev1.ReleaseCycle{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &releasev1.ReleaseCycle{}, v)
	}
	customResource := *customResourcePointer

	return customResource, nil
}
