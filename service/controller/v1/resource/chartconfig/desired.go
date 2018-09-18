package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	return nil, nil
}

func (r *Resource) newChartCR(customResource v1alpha1.Release) (*v1alpha1.ChartConfig, error) {
	newChartCR := &v1alpha1.ChartConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ChartConfig",
			APIVersion: "core.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: key.OperatorChartName(customResource),
			Labels: map[string]string{
				key.LabelApp:          key.OperatorName(customResource),
				key.LabelManagedBy:    key.ProjectName,
				key.LabelOrganization: key.OrganizationName,
				key.LabelServiceType:  key.ServiceTypeManaged,
			},
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      key.OperatorChartName(customResource),
				Namespace: metav1.NamespaceSystem,
				Channel:   key.OperatorChannelName(customResource),
				ConfigMap: TODO,
				Release:   TODO,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: TODO,
			},
		},
	}

	return newChartCR, nil
}
