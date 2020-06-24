package key

import (
	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// AppCatalog is the name of the app catalog where releases and release
	// components are stored.
	AppCatalog = "control-plane"

	// Namespace is the namespace where App CRs are created.
	Namespace = "giantswarm"

	LabelAppOperatorVersion = "app-operator.giantswarm.io/version"
	LabelManagedBy          = "giantswarm.io/managed-by"
	LabelServiceType        = "giantswarm.io/service-type"

	ValueServiceTypeManaged = "managed"
)

type DeletionTimestampGetter interface {
	GetDeletionTimestamp() *metav1.Time
}

func GetApps(g8sclient versioned.Interface, namespace string) ([]*applicationv1alpha1.App, error) {
	return nil, nil
}

func GetReleases(g8sclient versioned.Interface, namespace string) ([]*releasev1alpha1.Release, error) {
	return nil, nil
}

func IsDeleted(cr DeletionTimestampGetter) bool {
	return cr.GetDeletionTimestamp() != nil
}

// ReleaseVersion returns the version of the given release.
func ReleaseVersion(releaseCR releasev1alpha1.Release) string {
	return releaseCR.Name
}

// ToAppCR converts v into an App CR.
func ToAppCR(v interface{}) (*applicationv1alpha1.App, error) {
	x, ok := v.(*applicationv1alpha1.App)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", x, v)
	}

	return x, nil
}

// ToReleaseCR converts v into a Release CR.
func ToReleaseCR(v interface{}) (*releasev1alpha1.Release, error) {
	x, ok := v.(*releasev1alpha1.Release)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", x, v)
	}

	return x, nil
}
