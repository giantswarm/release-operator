package controllercontext

import (
	"context"

	"github.com/giantswarm/microerror"
)

type contextKey string

const controllerKey contextKey = "controller"

type Context struct {
	ConfigMap ContextConfigMap
	Secret    ContextSecret
}

type ContextConfigMap struct {
	Name            string
	Namespace       string
	ResourceVersion string
}

type ContextSecret struct {
	Name            string
	Namespace       string
	ResourceVersion string
}

func (c *Context) Validate() error {
	if c.ConfigMap.Name == "" {
		return microerror.Maskf(invalidContextError, "%T.ConfigMap.Name must not be empty", c)
	}
	if c.ConfigMap.Namespace == "" {
		return microerror.Maskf(invalidContextError, "%T.ConfigMap.Namespace must not be empty", c)
	}
	if c.ConfigMap.ResourceVersion == "" {
		return microerror.Maskf(invalidContextError, "%T.ConfigMap.ResourceVersion must not be empty", c)
	}

	if c.Secret.Name == "" {
		return microerror.Maskf(invalidContextError, "%T.Secret.Name must not be empty", c)
	}
	if c.Secret.Namespace == "" {
		return microerror.Maskf(invalidContextError, "%T.Secret.Namespace must not be empty", c)
	}
	if c.Secret.ResourceVersion == "" {
		return microerror.Maskf(invalidContextError, "%T.Secret.ResourceVersion must not be empty", c)
	}

	return nil
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
