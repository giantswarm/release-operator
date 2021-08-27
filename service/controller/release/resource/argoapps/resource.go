package argoapps

import (
	"context"
	"fmt"
	"reflect"

	appv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/argoapp/pkg/argoapp"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/release-operator/v2/service/controller/key"
)

const (
	Name = "argoapps"
)

var (
	argoAPISchema       = schema.GroupVersion{"argoproj.io", "v1alpha1"}
	argoApplicationKind = "Application"
	argoApplicationList = "ApplicationList"
	argoCDNamespace     = "argocd"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensureState(ctx context.Context) error {
	var releases releasev1alpha1.ReleaseList
	{
		err := r.k8sClient.CtrlClient().List(
			ctx,
			&releases,
		)
		if err != nil {
			return microerror.Mask(err)
		}
		releases = key.ExcludeDeletedRelease(releases)
		releases = key.ExcludeUnusedDeprecatedReleases(releases)
	}

	var components map[string]releasev1alpha1.ReleaseSpecComponent
	{
		components = key.ExtractComponents(releases)
	}

	var apps []unstructured.Unstructured
	{
		list, err := r.listApplications(ctx, "*")
		if err != nil {
			return microerror.Mask(err)
		}
		apps = list.Items
	}

	appsToDelete := calculateObsoleteApps(components, apps)
	for i, app := range appsToDelete.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting app %#q in namespace %#q", app.Name, app.Namespace))

		err := r.k8sClient.CtrlClient().Delete(
			ctx,
			&appsToDelete.Items[i],
		)
		if apierrors.IsNotFound(err) {
			// fall through.
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted app %#q in namespace %#q", app.Name, app.Namespace))
	}

	appsToCreate := calculateMissingApps(components, apps)
	for i, app := range appsToCreate.Items {
		appConfig := key.GetAppConfig(app, configs)
		if appConfig.ConfigMapRef.Name == "" && appConfig.SecretRef.Name == "" {
			// Skip this app
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("skipping app %#q as its config is not ready", app.Name))
			continue
		}

		appsToCreate.Items[i].Spec.Config.ConfigMap.Name = appConfig.ConfigMapRef.Name
		appsToCreate.Items[i].Spec.Config.ConfigMap.Namespace = appConfig.ConfigMapRef.Namespace
		appsToCreate.Items[i].Spec.Config.Secret.Name = appConfig.SecretRef.Name
		appsToCreate.Items[i].Spec.Config.Secret.Namespace = appConfig.SecretRef.Namespace

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating app %#q in namespace %#q", app.Name, app.Namespace))

		err := r.k8sClient.CtrlClient().Create(
			ctx,
			&appsToCreate.Items[i],
		)
		if apierrors.IsAlreadyExists(err) {
			// fall through.
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created app %#q in namespace %#q", app.Name, app.Namespace))
	}

	return nil
}

func calculateMissingApps(components map[string]releasev1alpha1.ReleaseSpecComponent, apps unstructured.UnstructuredList) unstructured.UnstructuredList {
	var missingApps unstructured.UnstructuredList

	for _, component := range components {
		// TODO(kuba): move code to key.ComponentAppCreated
		// if !key.ComponentAppCreated(component, apps.Items) {
		ac := argoapp.ApplicationConfig{
			Name:       key.BuildAppName(component),
			AppName:    component.Name,
			AppVersion: key.GetComponentRef(component),
			AppCatalog: component.Catalog,
			// TODO(kuba): Where does release-operator get this now? Do we copy
			// code that calls to github from config-controller? Do we need
			// those 2 values to compare apps at all?
			AppDestinationNamespace: "???",
			ConfigRef:               "???",
		}
		// NOTE(kuba): Comparing unstructured argo apps is a bit of a hassle.
		// Also release-operator would need access to catalog index. Is there
		// any way we can make this easier?

		// TODO(kuba): handle this err
		missingApp, _ := argoapp.NewApplication(ac)
		for _, app := range apps.Items {
			if !compareArgoApplications(missingApp, app) {
				missingApps.Items = append(missingApps.Items, missingApp)
			}
		}
	}
}

func calculateObsoleteApps(components map[string]releasev1alpha1.ReleaseSpecComponent, apps appv1alpha1.AppList) appv1alpha1.AppList {
	var obsoleteApps appv1alpha1.AppList

	for _, app := range apps.Items {
		if !key.AppReferenced(app, components) {
			obsoleteApps.Items = append(obsoleteApps.Items, app)
		}
	}

	return obsoleteApps
}

func (r *Resource) getApplication(ctx context.Context, name string) (argoapp.ApplicationConfig, error) {
	var a argoapp.ApplicationConfig

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(argoAPISchema.WithKind(argoApplicationKind))
	err := c.Get(ctx, client.ObjectKey{Namespace: argoCDNamespace, Name: name}, u)
	if err == nil {
		return a, microerror.Mask(err)
	}
	a, err = unstructuredToArgoApplicationConfig(u)
	if err == nil {
		return a, microerror.Mask(err)
	}

	return a, nil
}

func compareArgoApplications(a, b unstructured.Unstructured) bool {
	aName, ok, err := unstructured.NestedString(a, "metadata", "name")
	if !ok || err {
		return false
	}
	aName, ok, err := unstructured.NestedString(b, "metadata", "name")
	if !ok || err {
		return false
	}

	if aName != bName {
		return false
	}

	aSpec, ok, err := unstructured.NestedMap(a, "spec")
	if !ok || err {
		return false
	}
	bSpec, ok, err := unstructured.NestedMap(b, "spec")
	if !ok || err {
		return false
	}

	return reflect.DeepEqual(aSpec, bSpec)

}

// func unstructuredToArgoApplicationConfig(u unstructured.Unstructured) (ac argoapp.ApplicationConfig, err error) {
// 	var ok bool
//
// 	ac.Name, ok, err = unstructured.NestedString(u, "metadata", "name")
// 	if err != nil {
// 		return ac, microerror.Mask(err)
// 	} else if !ok {
// 		return microerror.Maskf(executionFailedError, "unstructured key missing")
// 	}
//
// 	env, ok, err := unstructured.NestedSlice(u, "spec", "source", "plugin", "env")
// 	if err != nil {
// 		return ac, microerror.Mask(err)
// 	} else if !ok {
// 		return microerror.Maskf(executionFailedError, "unstructured key missing")
// 	}
//
// 	for _, envItem := range env {
// 		m, ok := envItem.(map[string]string)
// 		if !ok {
// 			return microerror.Maskf(executionFailedError, "could not cast to map[string]string: %q", envItem)
// 		}
//
// 		name, nameOk := m["name"]
// 		value, valueOk := m["value"]
// 		if !nameOk || !valueOk {
// 			return microerror.Maskf(executionFailedError, "could extract name/value: %q", m)
// 		}
//
// 		switch name {
// 		case "KONFIGURE_APP_NAME":
// 			ac.AppName = value
// 		case "KONFIGURE_APP_VERSION":
// 			ac.AppVersion = value
// 		}
//
// 	}
//
// }

// func (r *Resource) listApplications(ctx context.Context) (u *unstructured.UnstructuredList, err error) {
// 	u = &unstructured.UnstructuredList{}
// 	u.SetGroupVersionKind(argoAPISchema.WithKind(argoApplicationListKind))
// 	err = r.k8sClient.CtrlClient().List(ctx, u,
// 		&client.ListOptions{
// 			Namespace: argoCDNamespace,
// 			LabelSelector: labels.SelectorFromSet(labels.Set{
// 				key.LabelManagedBy: project.Name(),
// 			}),
// 		},
// 	)
// 	return
// }
