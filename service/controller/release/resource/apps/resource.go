package apps

import (
	"context"

	appv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/release-operator/pkg/project"
	"github.com/giantswarm/release-operator/service/controller/key"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		releases = excludeDeletedRelease(releases)
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

	appsToCreate := calculateMissingApps(releases, apps)
	err := r.k8sClient.CtrlClient().Create(
		ctx,
		&appsToCreate,
	)
	if apierrors.IsAlreadyExists(err) {
		// fall through.
	} else if err != nil {
		return microerror.Mask(err)
	}

	appsToDelete := calculateObsoleteApps(releases, apps)
	err = r.k8sClient.CtrlClient().Delete(
		ctx,
		&appsToDelete,
	)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func calculateMissingApps(releases releasev1alpha1.ReleaseList, apps appv1alpha1.AppList) appv1alpha1.AppList {
	var missingApps appv1alpha1.AppList

	for _, release := range releases.Items {
		operators := key.ExtractOperators(release.Spec.Components)
		for _, operator := range operators {
			ref := key.GetOperatorRef(operator)
			if !key.AppInApps(apps.Items, key.BuildAppName(operator.Name, ref), ref) {
				missingApp := key.ConstructApp(operator.Name, ref)
				missingApps.Items = append(missingApps.Items, missingApp)
			}
		}
	}
	return missingApps
}

func calculateObsoleteApps(releases releasev1alpha1.ReleaseList, apps appv1alpha1.AppList) appv1alpha1.AppList {
	var obsoleteApps appv1alpha1.AppList

	var operators []releasev1alpha1.ReleaseSpecComponent
	for _, release := range releases.Items {
		operators = append(operators, key.ExtractOperators(release.Spec.Components)...)
	}

	for _, app := range apps.Items {
		if !key.AppInOperators(operators, app) {
			obsoleteApps.Items = append(obsoleteApps.Items, app)
		}
	}

	return obsoleteApps
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
