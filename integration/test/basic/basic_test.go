// +build k8srequired

package basic

import (
	"fmt"
	"testing"

	"github.com/giantswarm/e2esetup/chart/env"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
)

const (
	cr = "apiextensions-release-e2e"
)

func TestInstall(t *testing.T) {
	version := fmt.Sprintf(":%s", env.CircleSHA())
	err := h.InstallOperator("release-operator", "release", "", version)
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}

	// Test Creation
	{
		releaseValuesConfig := chartvalues.APIExtensionsReleaseE2EConfig{
			Namespace: "giantswarm",
			Operator: chartvalues.APIExtensionsReleaseE2EConfigOperator{
				Name:    "test-operator",
				Version: "1.0.0",
			},
			VersionBundle: chartvalues.APIExtensionsReleaseE2EConfigVersionBundle{
				Version: "1.0.0",
			},
		}

		releaseValues, err := chartvalues.NewAPIExtensionsReleaseE2E(releaseValuesConfig)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = r.Install(cr, releaseValues, "stable")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

}
