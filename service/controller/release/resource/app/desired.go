package app

import (
	"context"
	"fmt"
	"reflect"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/pkg/project"
	"github.com/giantswarm/release-operator/service/controller/key"
)

func (r *resourceStateGetter) GetDesiredState(ctx context.Context, obj interface{}) ([]*applicationv1alpha1.App, error) {
	cr, err := key.ToReleaseCR(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var desiredComponents []releasev1alpha1.ReleaseSpecComponent
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding desired components for release %#q", cr.Name))

		desiredComponents, err = r.getDesiredComponents(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d desired components for release %#q", len(desiredComponents), cr.Name))
	}

	var desiredAppCRs []*applicationv1alpha1.App
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computing desired app CRs for release %#q components", cr.Name))

		for _, c := range desiredComponents {
			appCR := newAppCR(appCRName(c), c.Name, c.Version, r.appCatalog)
			desiredAppCRs = append(desiredAppCRs, appCR)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed %d desired app CRs for release %#q components", len(desiredAppCRs), cr.Name))
	}

	return desiredAppCRs, nil
}

func (r *resourceStateGetter) getDesiredComponents(ctx context.Context, cr *releasev1alpha1.Release) ([]releasev1alpha1.ReleaseSpecComponent, error) {
	// If all App CRs for components of this EOL release were blindly
	// deleted there is a chance some active (non-EOL) release sharing some
	// of those components would stop working (that would be eventually
	// fixed when the broken release is reconciled again and App CRs for
	// missing components are recreated). To avoid such situations this
	// release's components that are also part of other active releases are
	// still desired.
	if cr.Status.Cycle.Phase != releasev1alpha1.CyclePhaseEOL && !key.IsDeleted(cr) {
		return cr.Spec.Components, nil
	}

	var activeReleases []releasev1alpha1.Release
	{
		opts := metav1.ListOptions{
			LabelSelector: key.LabelReleaseCyclePhase + "!=" + releasev1alpha1.CyclePhaseEOL.String(),
		}

		result, err := r.g8sClient.ReleaseV1alpha1().Releases().List(opts)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		activeReleases = result.Items
	}

	var desiredComponents []releasev1alpha1.ReleaseSpecComponent

	// If this an EOL release only components that are also part of other
	// non-EOL (active) releases are desired.
	for _, activeCR := range activeReleases {
		cs := componentIntersection(cr, &activeCR)
		desiredComponents = append(desiredComponents, cs...)
	}

	return desiredComponents, nil
}

func newAppCR(crName, appName, appVersion, appCatalog string) *applicationv1alpha1.App {
	appCR := &applicationv1alpha1.App{
		TypeMeta: applicationv1alpha1.NewAppTypeMeta(),
		ObjectMeta: metav1.ObjectMeta{
			Name:      crName,
			Namespace: key.Namespace,
			Labels: map[string]string{
				key.LabelAppOperatorVersion: project.Version(),
				key.LabelManagedBy:          project.Name(),
				key.LabelServiceType:        key.ValueServiceTypeManaged,
			},
		},
		Spec: applicationv1alpha1.AppSpec{
			Catalog: appCatalog,
			KubeConfig: applicationv1alpha1.AppSpecKubeConfig{
				InCluster: true,
			},
			Name:      appName,
			Namespace: key.Namespace,
			Version:   appVersion,
		},
	}

	return appCR
}

func componentIntersection(cr1, cr2 *releasev1alpha1.Release) []releasev1alpha1.ReleaseSpecComponent {
	var intersection []releasev1alpha1.ReleaseSpecComponent

	for _, c1 := range cr1.Spec.Components {
		for _, c2 := range cr2.Spec.Components {
			if reflect.DeepEqual(c1, c2) {
				intersection = append(intersection, c1)
			}
		}
	}

	return intersection
}
