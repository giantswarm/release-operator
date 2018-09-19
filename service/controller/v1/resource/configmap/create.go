package configmap

import (
	"context"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/v1/controllercontext"
)

func (r Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var resourceVersion string
	{
		configMap, err := r.k8sClient.CoreV1().ConfigMaps(r.namespace).Get(r.name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		resourceVersion = configMap.GetResourceVersion()
	}

	{
		c, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		c.ConfigMap.Name = r.name
		c.ConfigMap.Namespace = r.namespace
		c.ConfigMap.ResourceVersion = resourceVersion
	}

	return nil
}
