// +build k8srequired

package releasecycle

import (
	"reflect"
	"testing"
	"time"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/release-operator/integration/env"
	"github.com/giantswarm/release-operator/service/controller/key"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var releaseCycleCR = &releasev1alpha1.ReleaseCycle{
	TypeMeta: releasev1alpha1.NewReleaseCycleTypeMeta(),
	ObjectMeta: metav1.ObjectMeta{
		Name: "aws.v6.1.0",
	},
	Spec: releasev1alpha1.ReleaseCycleSpec{
		Phase: releasev1alpha1.CyclePhaseEnabled,
	},
}

var expectedAppCR = &applicationv1alpha1.App{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app-operator.giantswarm.io/version": "1.0.0-" + env.CircleSHA(),
			"giantswarm.io/managed-by":           "release-operator",
			"giantswarm.io/service-type":         "managed",
		},
		Name:      "release-aws.v6.1.0",
		Namespace: "giantswarm",
	},
	Spec: applicationv1alpha1.AppSpec{
		Catalog: "control-plane",
		KubeConfig: applicationv1alpha1.AppSpecKubeConfig{
			InCluster: true,
		},
		Name:      "release-aws",
		Namespace: "giantswarm",
		Version:   "v6.1.0",
	},
}

// TestReleaseAppCRCreate attests that release App CR is created for the corresponding ReleaseCycle CR.
//
// Steps:
//   - create release cycle: aws.v6.1.0
//   - wait for app: release-aws.v6.1.0
//   - check values of app: release-aws.v6.1.0
func TestReleaseAppCRCreate(t *testing.T) {
	// Create Release Cycle CR.
	{
		_, err := config.K8sClients.G8sClient().ReleaseV1alpha1().ReleaseCycles().Create(releaseCycleCR)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}
	}

	// Verifies that release App CR was correctly reconciled.
	{
		o := func() (err error) {
			appCR, err := config.K8sClients.G8sClient().ApplicationV1alpha1().Apps(key.Namespace).Get("release-aws.v6.1.0", v1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			if !reflect.DeepEqual(appCR.GetName(), expectedAppCR.GetName()) {
				return microerror.Maskf(waitError, "obj.GetName() = %#v, want %#v", appCR.GetName(), expectedAppCR.GetName())
			}
			if !reflect.DeepEqual(appCR.GetNamespace(), expectedAppCR.GetNamespace()) {
				return microerror.Maskf(waitError, "obj.GetNamespace() = %#v, want %#v", appCR.GetNamespace(), expectedAppCR.GetNamespace())
			}
			if !reflect.DeepEqual(appCR.GetLabels(), expectedAppCR.GetLabels()) {
				return microerror.Maskf(waitError, ">>> obj.GetLabels()\n%#v\n>>> want\n%#v\n", appCR.GetLabels(), expectedAppCR.GetLabels())
			}
			if !reflect.DeepEqual(appCR.Spec, expectedAppCR.Spec) {
				return microerror.Maskf(waitError, ">>> obj.Spec\n%#v\n>>> want\n%#v\n", appCR.Spec, expectedAppCR.Spec)
			}

			return err
		}
		b := backoff.NewMaxRetries(30, 1*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}
	}
}
