// +build k8srequired

package basic

import (
	"testing"

	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
)

const (
	cr = "apiextensions-release-e2e"
)

func TestInstall(t *testing.T) {

	// Test Creation
	{
		releaseValuesConfig := chartvalues.APIExtensionsReleaseE2EConfig{
			Active: "true",
			Authorities: []chartvalues.APIExtensionsReleaseE2EConfigAuthority{
				{
					Name:    "test-operator",
					Version: "1.0.0",
				},
			},
			Date:      "0001-01-01T00:00:00Z",
			Name:      "1.0.0",
			Namespace: "giantswarm",
			Provider:  "aws",
			Version:   "1.0.0",
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
