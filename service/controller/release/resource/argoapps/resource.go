package argoapps

import (
	"context"
	"fmt"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	argoappv1alpha1 "github.com/giantswarm/argoapp/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/argoapp/pkg/argoapp"

	"github.com/giantswarm/release-operator/v2/pkg/project"
	"github.com/giantswarm/release-operator/v2/service/controller/key"
)

const (
	Name = "argoapps"
)

var (
	argoCDNamespace = "argocd"
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

	componentApps := map[string]argoappv1alpha1.Application{}
	{
		for name, component := range components {
			argoApp, err := r.componentToArgoApplication(ctx, component)
			if err != nil {
				return microerror.Mask(err)
			}
			componentApps[name] = *argoApp
		}
	}

	var apps argoappv1alpha1.ApplicationList
	{
		err := r.k8sClient.CtrlClient().List(
			ctx,
			&apps,
			&client.ListOptions{
				Namespace: argoCDNamespace,
				LabelSelector: labels.SelectorFromSet(labels.Set{
					key.LabelManagedBy: project.Name(),
				}),
			},
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	appsToDelete := calculateObsoleteApps(componentApps, apps)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("calculated %d obsolete apps", len(appsToDelete.Items)))
	for i, app := range appsToDelete.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting app %#q in namespace %#q", app.GetName(), app.GetNamespace()))

		err := r.k8sClient.CtrlClient().Delete(
			ctx,
			&appsToDelete.Items[i],
		)
		if apierrors.IsNotFound(err) {
			// fall through.
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted app %#q in namespace %#q", app.GetName(), app.GetNamespace()))
	}

	appsToCreate := calculateMissingApps(componentApps, apps)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("calculated %d missing apps", len(appsToCreate.Items)))
	for i, app := range appsToCreate.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating app %#q in namespace %#q", app.GetName(), app.GetNamespace()))

		// ensure giantswarm.io/managed-by label
		if appsToCreate.Items[i].Labels == nil {
			appsToCreate.Items[i].Labels = map[string]string{}
		}
		appsToCreate.Items[i].Labels[key.LabelManagedBy] = project.Name()

		err := r.k8sClient.CtrlClient().Create(
			ctx,
			&appsToCreate.Items[i],
		)
		if apierrors.IsAlreadyExists(err) {
			// fall through.
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created app %#q in namespace %#q", app.GetName(), app.GetNamespace()))
	}

	return nil
}

func calculateMissingApps(componentApps map[string]argoappv1alpha1.Application, apps argoappv1alpha1.ApplicationList) argoappv1alpha1.ApplicationList {
	var missingApps argoappv1alpha1.ApplicationList

	for _, componentApp := range componentApps {
		missingApp := true
		for _, app := range apps.Items {
			if compareArgoApplications(componentApp, app) {
				missingApp = false
				break
			}
		}
		if missingApp {
			missingApps.Items = append(missingApps.Items, componentApp)
		}
	}

	return missingApps
}

func calculateObsoleteApps(componentApps map[string]argoappv1alpha1.Application, apps argoappv1alpha1.ApplicationList) argoappv1alpha1.ApplicationList {
	var obsoleteApps argoappv1alpha1.ApplicationList

	for _, app := range apps.Items {
		obsoleteApp := true
		for _, componentApp := range componentApps {
			if compareArgoApplications(componentApp, app) {
				obsoleteApp = false
				break
			}
		}
		if obsoleteApp {
			obsoleteApps.Items = append(obsoleteApps.Items, app)
		}
	}

	return obsoleteApps
}

func compareArgoApplications(a, b argoappv1alpha1.Application) bool {
	if a.Name != b.Name {
		return false
	}

	konfigureA := extractKonfigureVariables(a)
	konfigureB := extractKonfigureVariables(b)

	match := true
	match = match && konfigureA.AppName != "" && konfigureA.AppName == konfigureB.AppName
	match = match && konfigureA.AppVersion != "" && konfigureA.AppVersion == konfigureB.AppVersion
	match = match && konfigureA.AppCatalog != "" && konfigureA.AppCatalog == konfigureB.AppCatalog

	return match

}

type konfigureVariables struct {
	AppName    string
	AppVersion string
	AppCatalog string
}

func extractKonfigureVariables(app argoappv1alpha1.Application) *konfigureVariables {
	v := &konfigureVariables{}

	if app.Spec.Source.Plugin.Name != "konfigure" {
		return v
	}

	for _, envVar := range app.Spec.Source.Plugin.Env {
		switch envVar.Name {
		case "KONFIGURE_APP_NAME":
			v.AppName = envVar.Value
		case "KONFIGURE_APP_VERSION":
			v.AppVersion = envVar.Value
		case "KONFIGURE_APP_CATALOG":
			v.AppCatalog = envVar.Value
		}
	}

	return v
}

func (r *Resource) componentToArgoApplication(ctx context.Context, component releasev1alpha1.ReleaseSpecComponent) (app *argoappv1alpha1.Application, err error) {
	configRef := "main"
	if component.ConfigReference != "" {
		configRef = component.ConfigReference
	}

	ac := argoapp.ApplicationConfig{
		Name:                    key.BuildAppName(component),
		AppName:                 component.Name,
		AppVersion:              key.GetComponentRef(component),
		AppCatalog:              component.Catalog,
		AppDestinationNamespace: key.Namespace,
		ConfigRef:               configRef,
	}

	app, err = argoapp.NewApplication(ac)
	if err != nil {
		return app, microerror.Mask(err)
	}

	return app, nil
}
