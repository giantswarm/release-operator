package apps

import (
	"context"
	"fmt"

	appv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/k8sclient/v4/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/release-operator/v2/pkg/project"
	"github.com/giantswarm/release-operator/v2/service/controller/key"
)

const (
	Name = "apps"
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

	var apps appv1alpha1.AppList
	{
		err := r.k8sClient.CtrlClient().List(
			ctx,
			&apps,
			&client.ListOptions{
				LabelSelector: labels.SelectorFromSet(labels.Set{
					key.LabelManagedBy: project.Name(),
				}),
			},
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var configs corev1alpha1.ConfigList
	{
		err := r.k8sClient.CtrlClient().List(
			ctx,
			&configs,
			&client.ListOptions{
				LabelSelector: labels.SelectorFromSet(labels.Set{
					key.LabelManagedBy: project.Name(),
				}),
			},
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	appsToDelete := calculateObsoleteApps(components, apps)
	for _, app := range appsToDelete.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting app %#q in namespace %#q", app.Name, app.Namespace))

		err := r.k8sClient.CtrlClient().Delete(
			ctx,
			&app,
		)
		if apierrors.IsNotFound(err) {
			// fall through.
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted app %#q in namespace %#q", app.Name, app.Namespace))
	}

	appsToCreate := calculateMissingApps(components, apps)
	for _, app := range appsToCreate.Items {
		cmRef, secretRef := key.GetAppConfig(app, configs)
		if cmRef.Name == "" && secretRef.Name == "" {
			// Skip this app
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("skipping app %#q as its config is not ready", app.Name))
			continue
		}

		app.Spec.Config.ConfigMap.Name = cmRef.Name
		app.Spec.Config.ConfigMap.Namespace = cmRef.Namespace
		app.Spec.Config.Secret.Name = secretRef.Name
		app.Spec.Config.Secret.Namespace = secretRef.Namespace

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating app %#q in namespace %#q", app.Name, app.Namespace))

		err := r.k8sClient.CtrlClient().Create(
			ctx,
			&app,
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

func calculateMissingApps(components map[string]releasev1alpha1.ReleaseSpecComponent, apps appv1alpha1.AppList) appv1alpha1.AppList {
	var missingApps appv1alpha1.AppList

	for _, component := range components {
		if !key.ComponentAppCreated(component, apps.Items) {
			missingApp := key.ConstructApp(component)
			missingApps.Items = append(missingApps.Items, missingApp)
		}
	}

	return missingApps
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
