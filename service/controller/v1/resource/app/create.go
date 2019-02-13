package app

import (
	"context"
	"fmt"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	appCRs, err := toAppCRs(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(appCRs) != 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating %d App CRs in the Kubernetes API", len(appCRs)))

		for _, c := range appCRs {
			_, err = r.g8sClient.ApplicationV1alpha1().Apps(r.namespace).Create(c)
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created %d App CRs in the Kubernetes API", len(appCRs)))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "App CRs do not have to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentAppCRs, err := toAppCRs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredAppCRs, err := toAppCRs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var appCRsToCreate []*applicationv1alpha1.App
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing create state")

		for _, d := range desiredAppCRs {
			_, ok := getAppCR(currentAppCRs, d.Namespace, d.Name)
			if !ok {
				appCRsToCreate = append(appCRsToCreate, d)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed create state")
	}

	return appCRsToCreate, nil
}
