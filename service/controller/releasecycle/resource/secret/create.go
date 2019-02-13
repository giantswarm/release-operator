package secret

import (
	"context"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/releasecycle/controllercontext"
)

func (r Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var resourceVersion string
	{
		secret, err := r.k8sClient.CoreV1().Secrets(r.namespace).Get(r.name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		resourceVersion = secret.GetResourceVersion()
	}

	{
		c, err := controllercontext.FromContext(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		c.Secret.Name = r.name
		c.Secret.Namespace = r.namespace
		c.Secret.ResourceVersion = resourceVersion
	}

	return nil
}
