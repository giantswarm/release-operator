package setup

import (
	"context"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/release-operator/integration/env"
	"github.com/giantswarm/release-operator/integration/teardown"
)

func WrapTestMain(h *framework.Host, helmClient *helmclient.Client, l micrologger.Logger, m *testing.M) (int, error) {
	var v int
	var err error
	var errors []error

	err = h.CreateNamespace("giantswarm")
	if err != nil {
		errors = append(errors, err)
		v = 1
	}

	err = helmClient.EnsureTillerInstalled(context.Background())
	if err != nil {
		errors = append(errors, err)
		v = 1
	}

	var values string
	{
		c := chartvalues.ReleaseOperatorConfig{
			RegistryPullSecret: env.RegistryPullSecret(),
		}
		values, err = chartvalues.NewReleaseOperator(c)
		if err != nil {
			errors = append(errors, err)
			v = 1
		}
	}

	err = h.InstallBranchOperator("release-operator", "releasecycle", values)
	if err != nil {
		errors = append(errors, err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if !env.KeepResources() {
		// Only do full teardown when not on CI.
		if !env.CircleCI() {
			err := teardown.Teardown(h, helmClient)
			if err != nil {
				errors = append(errors, err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			h.Teardown()
		}
	}

	if len(errors) > 0 {
		return v, microerror.Mask(errors[0])
	}

	return v, nil
}
