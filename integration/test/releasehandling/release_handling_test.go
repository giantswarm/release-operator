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

var releaseCycleCREnabled = &releasev1alpha1.ReleaseCycle{
	TypeMeta: releasev1alpha1.NewReleaseCycleTypeMeta(),
	ObjectMeta: metav1.ObjectMeta{
		Name: "aws.v6.1.0",
	},
	Spec: releasev1alpha1.ReleaseCycleSpec{
		DisabledDate: releasev1alpha1.DeepCopyDate{time.Date(2019, 1, 12, 0, 0, 0, 0, time.UTC)},
		EnabledDate:  releasev1alpha1.DeepCopyDate{time.Date(2019, 1, 8, 0, 0, 0, 0, time.UTC)},
		Phase:        releasev1alpha1.CyclePhaseEnabled,
	},
}

var releaseCycleCREOL = &releasev1alpha1.ReleaseCycle{
	TypeMeta: releasev1alpha1.NewReleaseCycleTypeMeta(),
	ObjectMeta: metav1.ObjectMeta{
		Name: "aws.v6.1.0",
	},
	Spec: releasev1alpha1.ReleaseCycleSpec{
		DisabledDate: releasev1alpha1.DeepCopyDate{time.Date(2019, 4, 8, 0, 0, 0, 0, time.UTC)},
		EnabledDate:  releasev1alpha1.DeepCopyDate{time.Date(2019, 1, 8, 0, 0, 0, 0, time.UTC)},
		Phase:        releasev1alpha1.CyclePhaseEOL,
	},
}

// TestReleaseHandling tests the Release CR reconciliation.
//
// It checks for Release status, labels, and components App CRs.
//
// It runs following steps:
//
//	- Creates a Release CR.
//	- Checks if the CR has "release-operator.giantswarm.io/release-cycle-phase: upcoming" label reconciled.
//	- Checks if the CR has ".status.cycle.phase: upcoming" status reconciled.
//	- Verifies App CRs for the Release CR components exist.
//	- Marks the Release as enabled.
//	- Verifies Release status is enabled.
//
func TestReleaseHandling(t *testing.T) {
	ctx := context.Background()

	// Create the CR and make sure it doesn't have labels.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating Release CR %#q", releaseCR.Name))

		obj, err := config.K8sClients.G8sClient().ReleaseV1alpha1().Releases().Create(releaseCR)
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

	// Create the ReleaseCycle for the Release CR marking it as "enabled" release.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating ReleaseCycle CR %#q", releaseCycleCREnabled.Name))

		_, err := config.K8sClients.G8sClient().ReleaseV1alpha1().ReleaseCycles().Create(releaseCycleCREnabled)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created ReleaseCycle CR %#q", releaseCycleCREnabled.Name))
	}

	// After creating ReleaseCycle with "enabled" phase the
	// "release-operator.giantswarm.io/release-cycle-phase: enabled" label
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
			if v != releasev1alpha1.CyclePhaseEnabled.String() {
				return microerror.Maskf(waitError, "obj.Labels[%q] = %q, want %q", obj.Labels[k], releasev1alpha1.CyclePhaseEnabled.String())
			}

			return nil
		}
		b := backoff.NewMaxRetries(35, 6*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked Release CR %#q labels", releaseCR.Name))
	}

	// Verify that release was reconciled, status and label should be
	// updated with values from release cycle.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("check if Release CR %#q status was updated", releaseCR.Name))

		o := func() error {
			obj, err := config.K8sClients.G8sClient().ReleaseV1alpha1().Releases().Get(releaseCR.Name, metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			if !cmp.Equal(obj.Status.Cycle, releaseCycleCREnabled.Spec) {
				return microerror.Maskf(waitError, "\n\n%s\nobj.Status.Cycle = %#v\nreleaseCycleCR.Spec = %#v\n\n", cmp.Diff(obj.Status.Cycle, releaseCycleCREnabled.Spec), obj.Status.Cycle, releaseCycleCREnabled.Spec)
			}

			return nil
		}
		b := backoff.NewMaxRetries(35, 6*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked if Release CR %#q was updated", releaseCR.Name))
	}

	// Update the release to eol by updating the release cycle with phase=eol.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating ReleaseCycle CR %#q", releaseCycleCREnabled.Name))

		c, err := config.K8sClients.G8sClient().ReleaseV1alpha1().ReleaseCycles().Get(releaseCycleCREnabled.GetName(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		u := releaseCycleCREOL.DeepCopy()
		u.ObjectMeta = c.ObjectMeta

		_, err = config.K8sClients.G8sClient().ReleaseV1alpha1().ReleaseCycles().Update(u)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated ReleaseCycle CR %#q", u.Name))
	}

	// Verify that release was reconciled, status should be updated with values from release cycle.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking if Release CR status %#q was updated", releaseCR.Name))

		o := func() error {
			obj, err := config.K8sClients.G8sClient().ReleaseV1alpha1().Releases().Get(releaseCR.Name, metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			if !cmp.Equal(obj.Status.Cycle, releaseCycleCREOL.Spec) {
				return microerror.Maskf(waitError, "\n\n%s\nobj.Status.Cycle = %#v\nreleaseCycleCREOL.Spec = %#v\n\n", cmp.Diff(obj.Status.Cycle, releaseCycleCREOL.Spec), obj.Status.Cycle, releaseCycleCREOL.Spec)
			}

			return nil
		}
		b := backoff.NewMaxRetries(150, 1*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked if Release CR status %#q was updated", releaseCR.Name))
	}

	// Verify that release was reconciled, label should be updated with values from release cycle.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking if Release CR labels %#q were updated", releaseCR.Name))

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
			if v != releasev1alpha1.CyclePhaseEOL.String() {
				return microerror.Maskf(waitError, "obj.Labels[%q] = %q, want %q", obj.Labels[k], releasev1alpha1.CyclePhaseEOL.String())
			}

			return nil
		}
		b := backoff.NewMaxRetries(150, 1*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked if Release CR labels %#q were updated", releaseCR.Name))
	}

	// Verifies that components App CRs are gone.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking if App CR were removed", releaseCR.Name))

		o := func() error {
			list, err := config.K8sClients.G8sClient().ApplicationV1alpha1().Apps("").List(metav1.ListOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			expectedAppCRNames := []string{
				"aws-operator.4.6.0",
				"cert-operator.0.1.0",
			}
			for _, obj := range list.Items {
				for _, name := range expectedAppCRNames {
					if obj.GetName() == name {
						return microerror.Maskf(waitError, "not expected to found App CR %s", obj.GetName())
					}
				}
			}

			return nil
		}
		b := backoff.NewMaxRetries(30, 5*time.Second)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("err == %v, want %v", err, nil)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked if App CR were removed", releaseCR.Name))
	}
}
