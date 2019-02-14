package controllercontext

import (
	"context"

	"github.com/giantswarm/microerror"
)

type contextKey string

const controllerKey contextKey = "controller"

type Context struct {
}

func NewContext(ctx context.Context, c Context) context.Context {
	return context.WithValue(ctx, controllerKey, &c)
}

func FromContext(ctx context.Context) (*Context, error) {
	c, ok := ctx.Value(controllerKey).(*Context)
	if !ok {
		return nil, microerror.Mask(notFoundError)
	}

	return c, nil
}
