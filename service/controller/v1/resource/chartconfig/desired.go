package chartconfig

import (
	"context"

	corev1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	releasev1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/v1/controllercontext"
	"github.com/giantswarm/release-operator/service/controller/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseCycleCR, err := key.ToReleaseCycleCR(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseCR, err := r.g8sClient.ReleaseV1alpha1().Releases(releaseCycleCR.GetNamespace()).Get(releaseCycleCR.Spec.Release.Name, metav1.GetOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var desiredChartConfigCRs []*corev1.ChartConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing desired state")

		for _, a := range releaseCR.Spec.Components {
			chartConfigCR, err := r.newChartCR(ctx, *releaseCR, a)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			desiredChartConfigCRs = append(desiredChartConfigCRs, chartConfigCR)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed desired state")
	}

	return desiredChartConfigCRs, nil
}

func (r *Resource) newChartCR(ctx context.Context, customResource releasev1.Release, component releasev1.ReleaseSpecComponent) (*corev1.ChartConfig, error) {
	c, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	newChartCR := &corev1.ChartConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ChartConfig",
			APIVersion: "core.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: component.Name,
			Labels: map[string]string{
				key.LabelApp:            component.Name,
				key.LabelManagedBy:      key.ProjectName,
				key.LabelOrganization:   key.OrganizationName,
				key.LabelReleaseVersion: key.ReleaseVersion(customResource),
				key.LabelServiceType:    key.ServiceTypeManaged,
			},
		},
		Spec: corev1.ChartConfigSpec{
			Chart: corev1.ChartConfigSpecChart{
				Channel: component.Name,
				ConfigMap: corev1.ChartConfigSpecConfigMap{
					Name:            c.ConfigMap.Name,
					Namespace:       c.ConfigMap.Namespace,
					ResourceVersion: c.ConfigMap.ResourceVersion,
				},
				Name:      component.Name,
				Namespace: metav1.NamespaceSystem,
				Secret: corev1.ChartConfigSpecSecret{
					Name:            c.Secret.Name,
					Namespace:       c.Secret.Namespace,
					ResourceVersion: c.Secret.ResourceVersion,
				},
				Release: component.Name,
			},
			VersionBundle: corev1.ChartConfigSpecVersionBundle{
				Version: r.chartOperatorVersion,
			},
		},
	}

	return newChartCR, nil
}
