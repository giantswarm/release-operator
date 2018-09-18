package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

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
