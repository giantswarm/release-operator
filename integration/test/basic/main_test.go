// +build k8srequired

package basic

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/framework/resource"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/release-operator/integration/setup"
)

const (
	testName = "basic"
)

var (
	h          *framework.Host
	helmClient *helmclient.Client
	l          micrologger.Logger
	r          *resource.Resource
)

func init() {
	var err error

	{
		c := micrologger.Config{}
		l, err = micrologger.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := framework.HostConfig{
			Logger: l,

			ClusterID:       "n/a",
			VaultToken:      "n/a",
			TargetNamespace: "giantswarm",
		}
		h, err = framework.NewHost(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := helmclient.Config{
			Logger:     l,
			K8sClient:  h.K8sClient(),
			RestConfig: h.RestConfig(),

			TillerNamespace: "giantswarm",
		}
		helmClient, err = helmclient.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := resource.Config{
			Logger:     l,
			HelmClient: helmClient,

			Namespace: "giantswarm",
		}
		r, err = resource.New(c)
		if err != nil {
			panic(err.Error())
		}
	}
}

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	ctx := context.Background()

	v, err := setup.WrapTestMain(h, helmClient, l, m)
	if err != nil {
		l.LogCtx(ctx, "level", "error", "message", "e2e test failed", "stack", fmt.Sprintf("%#v\n", err))
	}

	os.Exit(v)
}
