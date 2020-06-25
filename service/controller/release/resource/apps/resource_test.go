package apps

import (
	"strconv"
	"testing"

	appv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/release-operator/service/controller/key"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_calculateMissingApps(t *testing.T) {

	testCases := []struct {
		name         string
		releases     releasev1alpha1.ReleaseList
		apps         appv1alpha1.AppList
		expectedApps appv1alpha1.AppList
	}{
		{
			name: "case 0: there is a missing app",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								{
									Name:    "test-operator",
									Version: "1.0.0",
								},
								{
									Name:    "abc-operator",
									Version: "123.0.0",
								},
							},
						},
					},
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								{
									Name:    "other-operator",
									Version: "2.0.0",
								},
							},
						},
					},
				},
			},
			apps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-operator-1.0.0",
						},
						Spec: appv1alpha1.AppSpec{
							Name:    "test-operator",
							Version: "1.0.0",
						},
					},
				},
			},
			expectedApps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					key.ConstructApp("abc-operator", "123.0.0"),
					key.ConstructApp("other-operator", "2.0.0"),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultApps := calculateMissingApps(tc.releases, tc.apps)

			if !cmp.Equal(resultApps, tc.expectedApps) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedApps, resultApps))
			}
		})
	}
}

func Test_calculateObsoleteApps(t *testing.T) {
	testCases := []struct {
		name         string
		releases     releasev1alpha1.ReleaseList
		apps         appv1alpha1.AppList
		expectedApps appv1alpha1.AppList
	}{
		{
			name: "case 0: there is a missing app",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								{
									Name:    "test-operator",
									Version: "1.0.0",
								},
								{
									Name:    "abc-operator",
									Version: "123.0.0",
								},
							},
						},
					},
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								{
									Name:    "other-operator",
									Version: "2.0.0",
								},
							},
						},
					},
				},
			},
			apps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					key.ConstructApp("test-operator", "1.0.0"),
					key.ConstructApp("abc-operator", "123.0.0"),
					key.ConstructApp("other-operator", "2.0.0"),
					key.ConstructApp("not-exist-operator", "1.0.0"),
				},
			},
			expectedApps: appv1alpha1.AppList{
				Items: []appv1alpha1.App{
					key.ConstructApp("not-exist-operator", "1.0.0"),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultApps := calculateObsoleteApps(tc.releases, tc.apps)

			if !cmp.Equal(resultApps, tc.expectedApps) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedApps, resultApps))
			}
		})
	}
}
