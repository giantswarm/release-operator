// +build k8srequired

package releasehandling

import (
	"reflect"
	"testing"
	"time"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var releaseCR1 = &releasev1alpha1.Release{
	TypeMeta: releasev1alpha1.NewReleaseTypeMeta(),
	ObjectMeta: metav1.ObjectMeta{
		Name: "aws.v6.1.0",
	},
	Spec: releasev1alpha1.ReleaseSpec{
		Changelog: []releasev1alpha1.ReleaseSpecChangelogEntry{
			{
				Component:   "cloudconfig",
				Description: "Replace cloudinit with ignition.",
				Kind:        "changed",
			},
		},
		Components: []releasev1alpha1.ReleaseSpecComponent{
			{
				Name:    "aws-operator",
				Version: "4.6.0",
			},
			{
				Name:    "cert-operator",
				Version: "0.1.0",
			},
		},
		ParentVersion: "6.0.0",
		Version:       "6.1.0",
	},
}

var releaseCR2 = &releasev1alpha1.Release{
	TypeMeta: releasev1alpha1.NewReleaseTypeMeta(),
	ObjectMeta: metav1.ObjectMeta{
		Name: "aws.v6.2.0",
	},
	Spec: releasev1alpha1.ReleaseSpec{
		Changelog: []releasev1alpha1.ReleaseSpecChangelogEntry{
			{
				Component:   "cloudconfig",
				Description: "Replace cloudinit with ignition.",
				Kind:        "changed",
			},
		},
		Components: []releasev1alpha1.ReleaseSpecComponent{
			{
				Name:    "aws-operator",
				Version: "4.7.0",
			},
			{
				Name:    "cert-operator",
				Version: "0.1.0",
			},
		},
		ParentVersion: "6.1.0",
		Version:       "6.2.0",
	},
}

// TestMultiReleaseHandling runs following steps:
//
//	- Creates a Release CR 1.
//	- Creates a Release CR 2 sharing a component (cert-operator@0.1.0) with
//	  the Release CR 1.
//	- Checks if App CRs for all components of Release CR 1 and Release CR
//	  2 are created.
//
func TestMultiReleaseHandling(t *testing.T) {
	// Create the Release CR 1.
	{
		_, err := config.K8sClients.G8sClient().ReleaseV1alpha1().Releases().Create(releaseCR1)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}
	}

	// Create the Release CR 2 sharing a component (cert-operator@0.1.0)
	// with the Release CR 1.
	{
		_, err := config.K8sClients.G8sClient().ReleaseV1alpha1().Releases().Create(releaseCR2)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}
	}

	// Check if App CRs  for all components of Release CR 1 and Release CR
	// 2 are created.
	{
		o := func() error {
			list, err := config.K8sClients.G8sClient().ApplicationV1alpha1().Apps("").List(metav1.ListOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			var appCRNames []string
			for _, obj := range list.Items {
				appCRNames = append(appCRNames, obj.Name)
			}

			expectedAppCRNames := []string{
				"aws-operator.4.6.0",
				"aws-operator.4.7.0",
				"cert-operator.0.1.0",
			}

			if !reflect.DeepEqual(appCRNames, expectedAppCRNames) {
				return microerror.Maskf(waitError, "\n\n%s\n", cmp.Diff(appCRNames, expectedAppCRNames))
			}

			return nil
		}
		b := backoff.NewMaxRetries(30, 5*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}
	}
}