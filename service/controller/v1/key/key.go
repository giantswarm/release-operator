package key

import (
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	LabelApp          = "app"
	LabelManagedBy    = "giantswarm.io/managed-by"
	LabelOrganization = "giantswarm.io/organization"
	LabelServiceType  = "giantswarm.io/service-type"
)

const (
	ProjectName        = "release-operator"
	OrganizationName   = "giantswarm"
	ServiceTypeManaged = "managed"
)

func OperatorChannelName(customResource v1alpha1.Release) string {
	return strings.Replace(OperatorVersion(customResource), ".", "-", -1)
}

func OperatorChartName(customResource v1alpha1.Release) string {
	return fmt.Sprintf("%s-chart", OperatorName(customResource))
}

func OperatorName(customResource v1alpha1.Release) string {
	return customResource.Spec.Operator.Name
}

func OperatorVersion(customResource v1alpha1.Release) string {
	return customResource.Spec.Operator.Version
}

func ReleaseName(customResource v1alpha1.Release) string {
	return fmt.Sprintf("%s-%s", OperatorName(customResource), OperatorChannelName(customResource))
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
