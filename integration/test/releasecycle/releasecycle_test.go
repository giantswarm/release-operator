// +build k8srequired

package releasecycle

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/release-operator/integration/env"
	"github.com/giantswarm/release-operator/service/controller/key"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestReleaseAppCRCreate attests that release App CR is created for the corresponding ReleaseCycle CR.
//
// Steps:
//   - create release cycle: aws.v6.1.0
//   - wait for app: release-aws.v6.1.0
//   - check values of app: release-aws.v6.1.0
func TestReleaseAppCRCreate(t *testing.T) {
	g8sClient := config.K8sClients.G8sClient()

	// Create Release Cycle CR.
	releaseCycleCR := &releasev1alpha1.ReleaseCycle{
		TypeMeta: releasev1alpha1.NewReleaseCycleTypeMeta(),
		ObjectMeta: metav1.ObjectMeta{
			Name: "aws.v6.1.0",
		},
		Spec: releasev1alpha1.ReleaseCycleSpec{
			Phase: releasev1alpha1.CyclePhaseEnabled,
		},
	}
	rc, err := g8sClient.ReleaseV1alpha1().ReleaseCycles().Create(releaseCycleCR)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for release App CR.
	var appCR *applicationv1alpha1.App
	ctx := context.Background()
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
	n := backoff.NewNotifier(config.Logger, ctx)
	o := func() (err error) {
		appCR, err = g8sClient.ApplicationV1alpha1().Apps(key.Namespace).Get("release-aws.v6.1.0", v1.GetOptions{})
		if err != nil {
			fmt.Println("app not found: test")
		} else {
			fmt.Printf("app found: %#q\n", appCR.GetName())
		}
		return err
	}

	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		t.Fatal(err)
	}

	// Test release App CR.
	expectedAppCR := &applicationv1alpha1.App{
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

	if !reflect.DeepEqual(appCR.GetName(), expectedAppCR.GetName()) {
		t.Errorf("wrong name, expected %#v got %#v", expectedAppCR.GetName(), appCR.GetName())
	}
	if !reflect.DeepEqual(appCR.GetNamespace(), expectedAppCR.GetNamespace()) {
		t.Errorf("wrong namespace, expected %#v  got %#v", expectedAppCR.GetNamespace(), appCR.GetNamespace())
	}
	if !reflect.DeepEqual(appCR.GetLabels(), expectedAppCR.GetLabels()) {
		t.Errorf("wrong labels\n>>> expected:\n%#v\n>>> got:\n%#v\n", expectedAppCR.GetLabels(), appCR.GetLabels())
	}
	if !reflect.DeepEqual(appCR.Spec, expectedAppCR.Spec) {
		t.Errorf("wrong spec\n>>> expected:\n%#v\n>>> got:\n%#v\n", expectedAppCR.Spec, appCR.Spec)
	}
}
