package configs

import (
	"context"
	"fmt"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	corev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
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
		releases = key.ExcludeDeletedRelease(releases)
		releases = key.ExcludeUnusedDeprecatedReleases(releases)
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


	r.logger.LogCtx(ctx, "level", "debug", "message", "Component list:")
	for i, conponent := range components.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("\t Component: %s", conponent.Name))
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "Config list:")

	for i, config := range configs.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("config %#q in namespace %#q", config.Name, config.Namespace))
	}


	configsToDelete := calculateObsoleteConfigs(components, configs)
	for i, config := range configsToDelete.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting config %#q in namespace %#q", config.Name, config.Namespace))

		err := r.k8sClient.CtrlClient().Delete(
			ctx,
			&configsToDelete.Items[i],
		)
		if apierrors.IsNotFound(err) {
			// fall through.
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted config %#q in namespace %#q", config.Name, config.Namespace))
	}

	configsToCreate := calculateMissingConfigs(components, configs)
	for i, config := range configsToCreate.Items {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating config %#q in namespace %#q", config.Name, config.Namespace))

		err := r.k8sClient.CtrlClient().Create(
			ctx,
			&configsToCreate.Items[i],
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

func calculateMissingConfigs(components map[string]releasev1alpha1.ReleaseSpecComponent, configs corev1alpha1.ConfigList) corev1alpha1.ConfigList {
	var missingConfigs corev1alpha1.ConfigList

	for _, component := range components {
		if !key.ComponentConfigCreated(component, configs.Items) {
			missingConfig := key.ConstructConfig(component)
			missingConfigs.Items = append(missingConfigs.Items, missingConfig)
		}
	}

	return missingConfigs
}

func calculateObsoleteConfigs(components map[string]releasev1alpha1.ReleaseSpecComponent, configs corev1alpha1.ConfigList) corev1alpha1.ConfigList {
	var obsoleteConfigs corev1alpha1.ConfigList

	for _, config := range configs.Items {
		if !key.ConfigReferenced(config, components) {
			obsoleteConfigs.Items = append(obsoleteConfigs.Items, config)
		}
	}

	return obsoleteConfigs
}
