package teardown

import (
	"fmt"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
)

func Teardown(h *framework.Host, helmClient *helmclient.Client) error {
	err := framework.HelmCmd(fmt.Sprintf("delete %s-release-operator --purge", h.TargetNamespace()))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}