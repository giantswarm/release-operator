package apps

import (
	"strconv"
	"testing"

	appv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/key"
)

var testOperators = []releasev1alpha1.ReleaseSpecComponent{
	{
		Name:    "test-operator",
		Version: "1.0.0",
	},
	{
		Name:    "abc-operator",
		Version: "123.0.0",
	},
	{
		Name:    "other-operator",
		Version: "2.0.0",
	},
	{
		Name:    "not-exist-operator",
		Version: "1.0.0",
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
				key.BuildAppName(testOperators[0]): testOperators[0],
				key.BuildAppName(testOperators[1]): testOperators[1],
				key.BuildAppName(testOperators[2]): testOperators[2],
			},
			apps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					appForOperator(testOperators[0]),
				},
			},
			expectedApps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					key.ConstructApp(testOperators[1]),
					key.ConstructApp(testOperators[2]),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultApps := calculateMissingApps(tc.operators, tc.apps)

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
				key.BuildAppName(testOperators[0]): testOperators[0],
				key.BuildAppName(testOperators[1]): testOperators[1],
				key.BuildAppName(testOperators[2]): testOperators[2],
			},
			apps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					key.ConstructApp(testOperators[0]),
					key.ConstructApp(testOperators[1]),
					key.ConstructApp(testOperators[2]),
					key.ConstructApp(testOperators[3]),
				},
			},
			expectedApps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					key.ConstructApp(testOperators[3]),
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

func Test_excludeDeletedRelease(t *testing.T) {
	testCases := []struct {
		name             string
		releases         releasev1alpha1.ReleaseList
		expectedReleases releasev1alpha1.ReleaseList
	}{
		{
			name: "case 0: some releases are being deleted",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "being-deleted",
							DeletionTimestamp: &metav1.Time{},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:              "not-being-deleted",
							DeletionTimestamp: nil,
						},
					},
				},
			},
			expectedReleases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultReleases := excludeDeletedRelease(tc.releases)

			if !cmp.Equal(resultReleases, tc.expectedReleases) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedReleases, resultReleases))
			}
		})
	}
}

func appForOperator(operator releasev1alpha1.ReleaseSpecComponent) appv1alpha1.App {
	return appv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name: key.BuildAppName(operator),
		},
		Spec: appv1alpha1.AppSpec{
			Name: operator.Name,
			//TODO this should be ref
			Version: operator.Version,
		},
	}
}
