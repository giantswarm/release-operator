package app

import (
	"context"

	"github.com/giantswarm/operatorkit/controller"
)

// ApplyUpdateChange has no effect.
func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	return nil
}

// NewUpdatePatch has no effect.
func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}
