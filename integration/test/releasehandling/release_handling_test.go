// +build k8srequired

package releasehandling

import (
	"testing"
	"time"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var releaseCR = &releasev1alpha1.Release{
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

// TestReleaseHandling runs following steps:
//
//	- Creates a Release CR.
//	- Checks if the CR has "release-operator.giantswarm.io/release-cycle-phase: upcoming" label reconciled.
//	- Checks if the CR has ".status.cycle.phase: upcoming" status reconciled.
//
func TestReleaseHandling(t *testing.T) {
	// Create the CR and make sure it doesn't have labels.
	{
		obj, err := config.K8sClients.G8sClient().ReleaseV1alpha1().Releases().Create(releaseCR)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		if len(obj.Labels) != 0 {
			t.Fatalf("len(obj.Labels) = %d, want 0", len(obj.Labels))
		}
	}

	// There is no corresponding ReleaseCycle CR so
	// "release-operator.giantswarm.io/release-cycle-phase: upcoming" label
	// should be reconciled on the created CR.
	{
		o := func() error {
			obj, err := config.K8sClients.G8sClient().ReleaseV1alpha1().Releases().Get(releaseCR.Name, metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			if obj.Labels == nil {
				return microerror.Maskf(waitError, "obj.Labels = %#v, want non-nil", obj.Labels)
			}

			k := "release-operator.giantswarm.io/release-cycle-phase"
			v := obj.Labels[k]
			if v != releasev1alpha1.CyclePhaseUpcoming.String() {
				return microerror.Maskf(waitError, "obj.Labels[%q] = %q, want %q", obj.Labels[k], releasev1alpha1.CyclePhaseUpcoming.String())
			}

			return nil
		}
		b := backoff.NewMaxRetries(30, 1*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}
	}

	// There is no corresponding ReleaseCycle CR so ".status.cycle.phase:
	// upcoming" status should be reconciled on the created CR.
	{
		o := func() error {
			obj, err := config.K8sClients.G8sClient().ReleaseV1alpha1().Releases().Get(releaseCR.Name, metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			if obj.Status.Cycle.Phase != releasev1alpha1.CyclePhaseUpcoming {
				return microerror.Maskf(waitError, "obj.Status.Cycle.Phase = %#v, want %q", obj.Status.Cycle.Phase, releasev1alpha1.CyclePhaseUpcoming)
			}

			return nil
		}
		b := backoff.NewMaxRetries(30, 1*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}
	}
}
