package app

import (
	"context"

	applicationv1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	var appCRs []*applicationv1.App
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding current state")

		o := metav1.ListOptions{}
		list, err := r.g8sClient.ApplicationV1alpha1().Apps(r.namespace).List(o)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, a := range list.Items {
			appCRs = append(appCRs, &a)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found current state")
	}

	return appCRs, nil
}
