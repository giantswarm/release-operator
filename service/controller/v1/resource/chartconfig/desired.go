package chartconfig

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/release-operator/service/controller/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToIndexReleases(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigs := make([]*v1alpha1.ChartConfig, 0)

	for _, indexRelease := range customObject {
		for _, auths := range indexRelease.Authorities {
			chartConfig := &v1alpha1.ChartConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       key.ChartConfigKind,
					APIVersion: key.ChartConfigAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: key.ChartConfigName(auths.Name),
				},
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:      key.ChartName(auths.Name),
						Namespace: metav1.NamespaceSystem,
						Channel:   key.ChartChannel(auths.Version, auths.Name),
						ConfigMap: v1alpha1.ChartConfigSpecConfigMap{},
						Release:   auths.Name,
						Secret:    v1alpha1.ChartConfigSpecSecret{},
					},
					VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
						Version: key.ChartConfigVersionbundleVersion,
					},
				},
			}
			desiredChartConfigs = append(desiredChartConfigs, chartConfig)
		}
	}
	return nil, nil
}
