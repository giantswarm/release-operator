package app

import (
	"context"

	"github.com/giantswarm/operatorkit/resource/crud"
)

// ApplyDeleteChange has no effect.
func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

// NewDeletePatch has no effect.
func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	return nil, nil
}
