package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	LabelApp            = "app"
	LabelManagedBy      = "giantswarm.io/managed-by"
	LabelOrganization   = "giantswarm.io/organization"
	LabelReleaseVersion = "release.giantswarm.io/version"
	LabelServiceType    = "giantswarm.io/service-type"
)

const (
	ProjectName        = "release-operator"
	OrganizationName   = "giantswarm"
	ServiceTypeManaged = "managed"
)

func ReleaseVersion(customResource v1alpha1.Release) string {
	return customResource.Spec.Version
}

func ToCustomResource(v interface{}) (v1alpha1.Release, error) {
	customResourcePointer, ok := v.(*v1alpha1.Release)
	if !ok {
		return v1alpha1.Release{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.Release{}, v)
	}
	customResource := *customResourcePointer

	return customResource, nil
}

func VersionBundleVersion(customResource v1alpha1.Release) string {
	return customResource.Spec.VersionBundle.Version
}
