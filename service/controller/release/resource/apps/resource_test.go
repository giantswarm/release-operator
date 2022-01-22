package apps

import (
	"sort"
	"strconv"
	"testing"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	releasev1alpha1 "github.com/giantswarm/release-operator/v3/api/v1alpha1"
	"github.com/giantswarm/release-operator/v3/service/controller/key"
)

var testComponents = []releasev1alpha1.ReleaseSpecComponent{
	{
		Name:    "test",
		Version: "1.0.0",
	},
	{
		Name:    "abc",
		Version: "123.0.0",
	},
	{
		Name:    "other",
		Version: "2.0.0",
	},
}

func Test_calculateMissingApps(t *testing.T) {
	testCases := []struct {
		name         string
		operators    map[string]releasev1alpha1.ReleaseSpecComponent
		apps         appv1alpha1.AppList
		expectedApps appv1alpha1.AppList
	}{
		{
			name: "case 0: an app is missing",
			operators: map[string]releasev1alpha1.ReleaseSpecComponent{
				key.BuildAppName(testComponents[0]): testComponents[0],
				key.BuildAppName(testComponents[1]): testComponents[1],
				key.BuildAppName(testComponents[2]): testComponents[2],
			},
			apps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					appForComponent(testComponents[0]),
				},
			},

			expectedApps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					key.ConstructApp(testComponents[1]),
					key.ConstructApp(testComponents[2]),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultApps := calculateMissingApps(tc.operators, tc.apps)

			sortSlice := func(slice []appv1alpha1.App) {
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
		operators    map[string]releasev1alpha1.ReleaseSpecComponent
		apps         appv1alpha1.AppList
		expectedApps appv1alpha1.AppList
	}{
		{
			name: "case 0: there is an obsolete app",
			operators: map[string]releasev1alpha1.ReleaseSpecComponent{
				key.BuildAppName(testComponents[0]): testComponents[0],
				key.BuildAppName(testComponents[1]): testComponents[1],
			},
			apps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					key.ConstructApp(testComponents[0]),
					key.ConstructApp(testComponents[1]),
					key.ConstructApp(testComponents[2]),
				},
			},
			expectedApps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					key.ConstructApp(testComponents[2]),
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

func appForComponent(operator releasev1alpha1.ReleaseSpecComponent) appv1alpha1.App {
	return appv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name: key.BuildAppName(operator),
		},
		Spec: appv1alpha1.AppSpec{
			Name:    operator.Name,
			Version: operator.Version,
		},
	}
}
