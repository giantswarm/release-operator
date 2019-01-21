package app

import (
	"context"

	applicationv1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	corev1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseCR, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var appCRs []*applicationv1.App
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing desired state")

		for _, authority := range releaseCR.Spec.Authorities {
			appCR, err := r.newAppCR(ctx, releaseCR, authority)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			appCRs = append(appCRs, appCR)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed desired state")
	}

	return appCRs, nil
}

func (r *Resource) newAppCR(ctx context.Context, releaseCR corev1.Release, authority corev1.ReleaseSpecAuthority) (*applicationv1.App, error) {
	appCR := &applicationv1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "application.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: authority.HelmReleaseName(),
			Labels: map[string]string{
				key.LabelApp:            authority.Name,
				key.LabelManagedBy:      key.ProjectName,
				key.LabelOrganization:   key.OrganizationName,
				key.LabelReleaseVersion: key.ReleaseVersion(releaseCR),
				key.LabelServiceType:    key.ServiceTypeManaged,
			},
		},
		Spec: applicationv1.AppSpec{
			Name:      authority.Name,
			Namespace: r.namespace,
			Release:   authority.Version,
			Catalog:   "",
			Config: applicationv1.AppSpecConfig{
				ConfigMap: applicationv1.AppSpecConfigConfigMap{
					Name:      "",
					Namespace: "",
				},
				Secret: applicationv1.AppSpecConfigSecret{
					Name:      "",
					Namespace: "",
				},
			},
			KubeConfig: applicationv1.AppSpecKubeConfig{
				Secret: applicationv1.AppSpecKubeConfigSecret{
					Name:      "",
					Namespace: "",
				},
			},
		},
	}

	return appCR, nil
}
