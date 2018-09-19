package chartconfig

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/release-operator/service/controller/v1/controllercontext"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	c, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	err = c.Validate()
	if controllercontext.IsInvalidContext(err) {
		// In case the controller context is not valid we miss certain information
		// necessary to compute the chart config CRs. In this case we stop here.
		r.logger.LogCtx(ctx, "level", "debug", "message", "invalid controller context")
		r.logger.LogCtx(ctx, "level", "debug", "message", "cannot compute chart config CR")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	return nil, nil
}
