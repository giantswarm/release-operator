package app

import (
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestAppCRFromFilled(modifyFunc func(*applicationv1alpha1.App)) *applicationv1alpha1.App {
	appCR := &applicationv1alpha1.App{
		TypeMeta: applicationv1alpha1.NewAppTypeMeta(),
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-app",
			Namespace: "test-namespace",
		},
		Spec: v1alpha1.AppSpec{
			Catalog: "test-spec-catalog",
			Config: v1alpha1.AppSpecConfig{
				ConfigMap: v1alpha1.AppSpecConfigConfigMap{
					Name:      "test-spec-config-configmap-name",
					Namespace: "test-spec-config-configmap-namespace",
				},
				Secret: v1alpha1.AppSpecConfigSecret{
					Name:      "test-spec-config-secret-name",
					Namespace: "test-spec-config-secret-namespace",
				},
			},
			KubeConfig: v1alpha1.AppSpecKubeConfig{
				Secret: v1alpha1.AppSpecKubeConfigSecret{
					Name:      "test-spec-kubeconfig-secret-name",
					Namespace: "test-spec-kubeconfig-secret-namespace",
				},
			},
			Name:      "test-spec-name",
			Namespace: "test-spec-namespace",
			UserConfig: v1alpha1.AppSpecConfig{
				ConfigMap: v1alpha1.AppSpecConfigConfigMap{
					Name:      "test-spec-userconfig-configmap-name",
					Namespace: "test-spec-userconfig-configmap-namespace",
				},
				Secret: v1alpha1.AppSpecConfigSecret{
					Name:      "test-spec-userconfig-secret-name",
					Namespace: "test-spec-userconfig-secret-namespace",
				},
			},
			Version: "test-spec-version",
		},
		Status: v1alpha1.AppStatus{
			AppVersion: "test-status",
			Release: v1alpha1.AppStatusRelease{
				LastDeployed: v1alpha1.DeepCopyTime{
					Time: time.Date(2019, 2, 12, 12, 4, 0, 0, time.UTC),
				},
				Status: "test-status-release-status",
			},
			Version: "test-status-version",
		},
	}

	modifyFunc(appCR)
	return appCR
}
