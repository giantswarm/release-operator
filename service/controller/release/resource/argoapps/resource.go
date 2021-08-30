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

	var componentApps map[string]unstructured.Unstructured
	{
		for name, component := range components {
			argoApp, err := r.componentToArgoApplication(ctx, component)
			if err != nil {
				return microerror.Mask(err)
			}
			componentApps[name] = argoApp
		}
	}

	var apps unstructured.UnstructuredList
	{
		list, err := r.listApplications(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		apps.Items = list.Items
	}

	appsToDelete := calculateObsoleteApps(componentApps, apps)
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

	appsToCreate := calculateMissingApps(componentApps, apps)
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

func calculateMissingApps(componentApps map[string]unstructured.Unstructured, apps unstructured.UnstructuredList) unstructured.UnstructuredList {
	var missingApps unstructured.UnstructuredList

	for _, component := range componentApps {
		for _, app := range apps.Items {
			if !compareArgoApplications(componentApp, app) {
				missingApps.Items = append(missingApps.Items, componentApp)
			}
		}
	}

	return missingApps
}

func calculateObsoleteApps(componentApps map[string]unstructured.Unstructured, apps unstructured.UnstructuredList) unstructured.UnstructuredList {
	var obsoleteApps unstructured.UnstructuredList

	for _, app := range apps.Items {
		obsoleteApp := true
		for _, componentApp := range componentApps {
			if compareArgoApplications(componentApp, app) {
				obsoleteApp = false
				break
			}
		}
		if obsoleteApp {
			obsoleteApps = append(obsoleteApps.Items, app)
		}
	}

	return obsoleteApps
}

func compareArgoApplications(a, b unstructured.Unstructured) bool {
	aName, ok, err := unstructured.NestedString(a, "metadata", "name")
	if !ok || err {
		return false
	}
	bName, ok, err := unstructured.NestedString(b, "metadata", "name")
	if !ok || err {
		return false
	}

	if aName != bName {
		return false
	}

	// .spec.source.plugin contains information about app name, version, and catalog.
	aSpec, ok, err := unstructured.NestedMap(a, "spec", "source", "plugin")
	if !ok || err {
		return false
	}
	bSpec, ok, err := unstructured.NestedMap(b, "spec", "source", "plugin")
	if !ok || err {
		return false
	}

	return reflect.DeepEqual(aSpec, bSpec)

}

func (r *Resource) listApplications(ctx context.Context) (u *unstructured.UnstructuredList, err error) {
	u = &unstructured.UnstructuredList{}
	u.SetGroupVersionKind(argoAPISchema.WithKind(argoApplicationListKind))
	err = r.k8sClient.CtrlClient().List(ctx, u,
		&client.ListOptions{
			Namespace: argoCDNamespace,
			LabelSelector: labels.SelectorFromSet(labels.Set{
				key.LabelManagedBy: project.Name(),
			}),
		},
	)
	return
}

func (r *Resource) componentToArgoApplication(ctx context.Context, component releasev1alpha1.ReleaseSpecComponent) (unstructured.Unstructured, error) {
	ac := argoapp.ApplicationConfig{
		Name:                    key.BuildAppName(component),
		AppName:                 component.Name,
		AppVersion:              key.GetComponentRef(component),
		AppCatalog:              component.Catalog,
		AppDestinationNamespace: key.Namespace,
		// TODO(kuba): check catalog for version
		ConfigRef: "does-not-matter",
	}

	app, err := argoapp.NewApplication(ac)
	if err != nil {
		return app, microerror.Mask(err)
	}
	return app, nil
}
