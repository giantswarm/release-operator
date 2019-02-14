package app

import (
	"context"

	"github.com/giantswarm/operatorkit/controller"
)

// ApplyDeleteChange has no effect.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

// NewDeletePatch has no effect.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}
