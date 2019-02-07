package app

import (
	"context"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	var appCRs []*applicationv1alpha1.App
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing desired state")

		for _, component := range releaseCR.Spec.Components {
			appCR, err := r.newAppCR(ctx, *releaseCR, component)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			appCRs = append(appCRs, appCR)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed desired state")
	}

	return appCRs, nil
}

func (r *Resource) newAppCR(ctx context.Context, releaseCR releasev1alpha1.Release, component releasev1alpha1.ReleaseSpecComponent) (*applicationv1alpha1.App, error) {
	appCR := &applicationv1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "application.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: component.Name,
			Labels: map[string]string{
				key.LabelApp: component.Name,
				// TODO: define app-operator version.
				key.LabelAppOperatorVersion: "",
				key.LabelManagedBy:          key.ProjectName,
				key.LabelOrganization:       key.OrganizationName,
				key.LabelReleaseVersion:     key.ReleaseVersion(releaseCR),
				key.LabelServiceType:        key.ServiceTypeManaged,
			},
		},
		Spec: applicationv1alpha1.AppSpec{
			Name:      component.Name,
			Namespace: r.namespace,
			Version:   component.Version,
			Catalog:   "",
			Config: applicationv1alpha1.AppSpecConfig{
				ConfigMap: applicationv1alpha1.AppSpecConfigConfigMap{
					Name:      "",
					Namespace: "",
				},
				Secret: applicationv1alpha1.AppSpecConfigSecret{
					Name:      "",
					Namespace: "",
				},
			},
			KubeConfig: applicationv1alpha1.AppSpecKubeConfig{
				Secret: applicationv1alpha1.AppSpecKubeConfigSecret{
					Name:      "",
					Namespace: "",
				},
			},
		},
	}

	return appCR, nil
}
