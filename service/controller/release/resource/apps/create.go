package apps

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	err := r.ensureState(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
