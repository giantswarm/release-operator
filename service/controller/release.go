package controller

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/release-operator/pkg/project"
	controllerv1 "github.com/giantswarm/release-operator/service/controller/v1"
)

const (
	releaseOperatorVersionLabel = "release-operator.giantswarm.io/version"
)

type ReleaseConfig struct {
	G8sClient    versioned.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	ProjectName string
}

type Release struct {
	*controller.Controller
}

func NewRelease(config ReleaseConfig) (*Release, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}

	var err error

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.Config{
			K8sExtClient: config.K8sExtClient,
			Logger:       config.Logger,
		}

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Logger:  config.Logger,
			Watcher: config.G8sClient.Release().ReleaseCycles(""),

			ListOptions: metav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=%s", releaseOperatorVersionLabel, project.Version()),
			},
			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var v1ResourceSet *controller.ResourceSet
	{
		c := controllerv1.ResourceSetConfig{
			G8sClient:   config.G8sClient,
			K8sClient:   config.K8sClient,
			Logger:      config.Logger,
			ProjectName: config.ProjectName,
		}

		v1ResourceSet, err = controllerv1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var releaseController *controller.Controller
	{
		c := controller.Config{
			CRD:       releasev1alpha1.NewReleaseCycleCRD(),
			CRDClient: crdClient,
			Informer:  newInformer,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				v1ResourceSet,
			},
			RESTClient: config.G8sClient.ReleaseV1alpha1().RESTClient(),

			Name: config.ProjectName,
		}

		releaseController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Release{
		Controller: releaseController,
	}

	return c, nil
}
