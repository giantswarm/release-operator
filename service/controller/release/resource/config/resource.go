package config

import (
	"context"
	"fmt"

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
	Name = "configs"
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
		releases = excludeDeletedRelease(releases)
		releases = r.excludeUnusedDeprecatedReleases(releases)
	}

	var components map[string]releasev1alpha1.ReleaseSpecComponent
	{
		components = key.ExtractComponents(releases)
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

	configsToDelete := calculateObsoleteConfigs(components, configs)
	for _, config := range configsToDelete.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting config %#q in namespace %#q", config.Name, config.Namespace))

		err := r.k8sClient.CtrlClient().Delete(
			ctx,
			&config,
		)
		if apierrors.IsNotFound(err) {
			// fall through.
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted config %#q in namespace %#q", config.Name, config.Namespace))
	}

	configsToCreate := calculateMissingConfigs(components, configs)
	for _, config := range configsToCreate.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating config %#q in namespace %#q", config.Name, config.Namespace))

		err := r.k8sClient.CtrlClient().Create(
			ctx,
			&config,
		)
		if apierrors.IsAlreadyExists(err) {
			// fall through.
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created config %#q in namespace %#q", config.Name, config.Namespace))
	}

	return nil
}

func (r *Resource) excludeUnusedDeprecatedReleases(releases releasev1alpha1.ReleaseList) releasev1alpha1.ReleaseList {
	var active releasev1alpha1.ReleaseList

	for _, release := range releases.Items {
		if release.Spec.State == releasev1alpha1.StateDeprecated && !release.Status.InUse {
			// Skip this release
			r.logger.Log("level", "debug", "message", fmt.Sprintf("excluding release %s because it is deprecated and unused", release.Name))
		} else {
			active.Items = append(active.Items, release)
		}
	}

	return active
}

func calculateMissingConfigs(components map[string]releasev1alpha1.ReleaseSpecComponent, configs corev1alpha1.ConfigList) corev1alpha1.ConfigList {
	var missingConfigs corev1alpha1.ConfigList

	for _, component := range components {
		if !key.ComponentCreated(component, configs.Items) {
			missingConfig := key.ConstructApp(component)
			missingConfigs.Items = append(missingConfigs.Items, missingConfig)
		}
	}

	return missingConfigs
}

func calculateObsoleteApps(components map[string]releasev1alpha1.ReleaseSpecComponent, configs corev1alpha1.ConfigList) corev1alpha1.ConfigList {
	var obsoleteApps corev1alpha1.ConfigList

	for _, config := range configs.Items {
		if !key.AppReferenced(config, components) {
			obsoleteConfigs.Items = append(obsoleteConfigs.Items, config)
		}
	}

	return obsoleteConfigs
}

func excludeDeletedRelease(releases releasev1alpha1.ReleaseList) releasev1alpha1.ReleaseList {
	var active releasev1alpha1.ReleaseList
	for _, release := range releases.Items {
		if release.DeletionTimestamp == nil {
			active.Items = append(active.Items, release)
		}
	}
	return active
}
