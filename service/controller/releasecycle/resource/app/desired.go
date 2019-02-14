package app

import (
	"context"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/pkg/project"
	"github.com/giantswarm/release-operator/service/controller/key"
)

// GetDesiredState computes the desired App CR for the release referenced in obj ReleaseCycle CR.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	releaseCycleCR, err := key.ToReleaseCycleCR(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	appName := key.ReleaseAppCRName(releaseCycleCR)
	r.logger.LogCtx(ctx, "level", "debug", "message", "computing desired state", "app", appName)

	releaseProvider, releaseVersion, err := key.SplitReleaseName(releaseCycleCR.GetName())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseRepo := key.ReleasePrefix(releaseProvider)
	appCR := r.newAppCR(appName, releaseRepo, releaseVersion)

	r.logger.LogCtx(ctx, "level", "debug", "message", "computed desired state", "app", appName)

	return appCR, nil
}

func (r *Resource) newAppCR(name, repository, version string) *applicationv1alpha1.App {
	appCR := &applicationv1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "application.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				key.LabelAppOperatorVersion: project.Version(),
				key.LabelManagedBy:          project.Name(),
				key.LabelServiceType:        key.ServiceTypeManaged,
			},
		},
		Spec: applicationv1alpha1.AppSpec{
			Name:      repository,
			Namespace: r.namespace,
			Version:   version,
		},
	}

	return appCR
}
