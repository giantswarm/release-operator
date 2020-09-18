package apps

import (
	"sort"
	"strconv"
	"testing"

	"github.com/giantswarm/release-operator/service/controller/unittest"

	appv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/release-operator/service/controller/key"
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

func Test_excludeDeprecatedUnusedRelease(t *testing.T) {
	testCases := []struct {
		name             string
		releases         releasev1alpha1.ReleaseList
		expectedReleases releasev1alpha1.ReleaseList
	}{
		{
			name: "case 0: an unused, deprecated release is deleted",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ancient-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: releasev1alpha1.StateDeprecated,
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: false,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
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
		{
			name: "case 1: an unused active release is not deleted",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "active-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: releasev1alpha1.StateActive,
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: false,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
			expectedReleases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "active-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: "active",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
		},
		{
			name: "case 2: a used deprecated release is not deleted",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "deprecated-used-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: releasev1alpha1.StateDeprecated,
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: true,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
			expectedReleases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "deprecated-used-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: "deprecated",
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: true,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
		},
		{
			name: "case 3: an unused wip release is not deleted",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "deprecated-used-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: releasev1alpha1.StateWIP,
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: false,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "not-being-deleted",
						},
					},
				},
			},
			expectedReleases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "deprecated-used-release",
						},
						Spec: releasev1alpha1.ReleaseSpec{
							State: "wip",
						},
						Status: releasev1alpha1.ReleaseStatus{
							InUse: false,
						},
					},
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

			fakeK8sClient := unittest.FakeK8sClient()

			var newResource *Resource
			{
				c := Config{
					K8sClient: fakeK8sClient,
					Logger:    microloggertest.New(),
				}
				var err error
				newResource, err = New(c)
				if err != nil {
					t.Fatal("expected", nil, "got", err)
				}
			}

			resultReleases := newResource.excludeUnusedDeprecatedReleases(tc.releases)

			if !cmp.Equal(resultReleases, tc.expectedReleases) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedReleases, resultReleases))
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
