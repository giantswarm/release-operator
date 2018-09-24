package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/v1/controllercontext"
	"github.com/giantswarm/release-operator/service/controller/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customResource, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var desiredChartConfigCRs []*v1alpha1.ChartConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing desired state")

		for _, a := range customResource.Spec.Authorities {
			chartConfigCR, err := r.newChartCR(ctx, customResource, a)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			desiredChartConfigCRs = append(desiredChartConfigCRs, chartConfigCR)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed desired state")
	}

	return desiredChartConfigCRs, nil
}

func (r *Resource) newChartCR(ctx context.Context, customResource v1alpha1.Release, authority v1alpha1.ReleaseSpecAuthority) (*v1alpha1.ChartConfig, error) {
	c, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	newChartCR := &v1alpha1.ChartConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ChartConfig",
			APIVersion: "core.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: authority.HelmChartName(),
			Labels: map[string]string{
				key.LabelApp:            authority.Name,
				key.LabelManagedBy:      key.ProjectName,
				key.LabelOrganization:   key.OrganizationName,
				key.LabelReleaseVersion: key.ReleaseVersion(customResource),
				key.LabelServiceType:    key.ServiceTypeManaged,
			},
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Channel: authority.HelmReleaseName(),
				ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
					Name:            c.ConfigMap.Name,
					Namespace:       c.ConfigMap.Namespace,
					ResourceVersion: c.ConfigMap.ResourceVersion,
				},
				Name:      authority.HelmChartName(),
				Namespace: metav1.NamespaceSystem,
				Secret: v1alpha1.ChartConfigSpecSecret{
					Name:            c.Secret.Name,
					Namespace:       c.Secret.Namespace,
					ResourceVersion: c.Secret.ResourceVersion,
				},
				Release: authority.HelmReleaseName(),
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: r.chartOperatorVersion,
			},
		},
	}

	return newChartCR, nil
}
