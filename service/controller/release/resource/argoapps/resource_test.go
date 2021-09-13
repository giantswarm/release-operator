package argoapps

import (
	"sort"
	"strconv"
	"testing"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	argoappv1alpha1 "github.com/giantswarm/argoapp/pkg/apis/application/v1alpha1"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/release-operator/v2/service/controller/key"
)

var testComponents = []releasev1alpha1.ReleaseSpecComponent{
	{
		Catalog: "giantswarm",
		Name:    "test",
		Version: "1.0.0",
	},
	{
		Catalog: "giantswarm",
		Name:    "abc",
		Version: "123.0.0",
	},
	{
		Catalog:         "giantswarm",
		ConfigReference: "test-branch",
		Name:            "other",
		Version:         "2.0.0",
	},
}

func Test_calculateMissingApps(t *testing.T) {
	testCases := []struct {
		name         string
		operators    []argoappv1alpha1.Application
		apps         argoappv1alpha1.ApplicationList
		expectedApps argoappv1alpha1.ApplicationList
	}{
		{
			name: "case 0: an app is missing",
			operators: []argoappv1alpha1.Application{
				mustComponentToArgoApplication(testComponents[0]),
				mustComponentToArgoApplication(testComponents[1]),
				mustComponentToArgoApplication(testComponents[2]),
			},
			apps: argoappv1alpha1.ApplicationList{
				Items: []argoappv1alpha1.Application{
					mustComponentToArgoApplication(testComponents[0]),
				},
			},

			expectedApps: argoappv1alpha1.ApplicationList{
				Items: []argoappv1alpha1.Application{
					mustComponentToArgoApplication(testComponents[1]),
					mustComponentToArgoApplication(testComponents[2]),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultApps := calculateMissingApps(tc.operators, tc.apps)

			sortSlice := func(slice []argoappv1alpha1.Application) {
				sort.SliceStable(slice, func(i, j int) bool { return slice[i].Name < slice[j].Name })
			}

			// calculateMissingApps iterates over a map -> random order
			sortSlice(tc.expectedApps.Items)
			sortSlice(resultApps.Items)

			if !cmp.Equal(resultApps, tc.expectedApps) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedApps, resultApps))
			}
		})
	}
}

func Test_calculateObsoleteApps(t *testing.T) {
	testCases := []struct {
		name         string
		operators    []argoappv1alpha1.Application
		apps         argoappv1alpha1.ApplicationList
		expectedApps argoappv1alpha1.ApplicationList
	}{
		{
			name: "case 0: there is an obsolete app",
			operators: []argoappv1alpha1.Application{
				mustComponentToArgoApplication(testComponents[0]),
				mustComponentToArgoApplication(testComponents[1]),
			},
			apps: argoappv1alpha1.ApplicationList{
				Items: []argoappv1alpha1.Application{
					mustComponentToArgoApplication(testComponents[0]),
					mustComponentToArgoApplication(testComponents[1]),
					mustComponentToArgoApplication(testComponents[2]),
				},
			},

			expectedApps: argoappv1alpha1.ApplicationList{
				Items: []argoappv1alpha1.Application{
					mustComponentToArgoApplication(testComponents[2]),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultApps := calculateObsoleteApps(tc.operators, tc.apps)

			if !cmp.Equal(resultApps, tc.expectedApps) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedApps, resultApps))
			}
		})
	}
}

func mustComponentToArgoApplication(component releasev1alpha1.ReleaseSpecComponent) argoappv1alpha1.Application {
	app, err := key.ComponentToArgoApplication(component)
	if err != nil {
		panic(err)
	}
	return *app
}
