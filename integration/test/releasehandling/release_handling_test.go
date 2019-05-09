// +build k8srequired

package releasehandling

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/release-operator/integration/env"
	"github.com/google/go-cmp/cmp"
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
//	- Verifies App CRs for the Release CR components exist.
//
func TestReleaseHandling(t *testing.T) {
	ctx := context.Background()

	// Create the CR and make sure it doesn't have labels.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating Release CR %#q", releaseCR.Name))

		_, err := config.K8sClients.G8sClient().ReleaseV1alpha1().Releases().Create(releaseCR)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		if len(obj.Labels) != 0 {
			t.Fatalf("len(obj.Labels) = %d, want 0", len(obj.Labels))
		}

		defer func() {
			if env.CircleCI() || env.KeepResources() {
				return
			}

			config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cleaning up Release CR %#q", releaseCR.Name))

			err := config.K8sClients.G8sClient().ReleaseV1alpha1().Releases().Delete(releaseCR.Name, &metav1.DeleteOptions{})
			if err != nil {
				config.Logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("failed to clean up Release CR %#q", releaseCR.Name), "stack", fmt.Sprintf("%#v", err))
			}

			config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cleaned up Release CR %#q", releaseCR.Name))
		}()

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created Release CR %#q", releaseCR.Name))
	}

	// There is no corresponding ReleaseCycle CR so
	// "release-operator.giantswarm.io/release-cycle-phase: upcoming" label
	// should be reconciled on the created CR.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking Release CR %#q labels", releaseCR.Name))

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

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked Release CR %#q labels", releaseCR.Name))
	}

	// There is no corresponding ReleaseCycle CR so ".status.cycle.phase:
	// upcoming" status should be reconciled on the created CR.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking Release CR %#q status", releaseCR.Name))

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

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked Release CR %#q status", releaseCR.Name))
	}

	// Verifies App CRs for the Release CR components exist.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking App CRs for Release CR %#q components", releaseCR.Name))

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
				"cert-operator.0.1.0",
			}

			sort.Strings(appCRNames)
			sort.Strings(expectedAppCRNames)

			if !cmp.Equal(appCRNames, expectedAppCRNames) {
				return microerror.Maskf(waitError, "\n\n%s\nappCRNames = %#v\nexpectedAppCRNames = %#v\n\n", cmp.Diff(appCRNames, expectedAppCRNames), appCRNames, expectedAppCRNames)
			}

			return nil
		}
		b := backoff.NewMaxRetries(30, 5*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked App CRs for Release CR %#q components", releaseCR.Name))
	}
}
