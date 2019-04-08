package app

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_resourceStateGetter_getDesiredComponents(t *testing.T) {
	testCases := []struct {
		name                  string
		inputRelease          *releasev1alpha1.Release
		inputExistingReleases []*releasev1alpha1.Release
		expectedComponents    []releasev1alpha1.ReleaseSpecComponent
		errorMatcher          func(err error) bool
	}{
		{
			name:                  "case 0: non-EOL",
			inputExistingReleases: nil,
			inputRelease: &releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Name:    "c1",
							Version: "v1",
						},
						{
							Name:    "c2",
							Version: "v2",
						},
					},
				},
				Status: releasev1alpha1.ReleaseStatus{
					Cycle: releasev1alpha1.ReleaseCycleSpec{
						Phase: releasev1alpha1.CyclePhaseEnabled,
					},
				},
			},
			expectedComponents: []releasev1alpha1.ReleaseSpecComponent{
				{
					Name:    "c1",
					Version: "v1",
				},
				{
					Name:    "c2",
					Version: "v2",
				},
			},
			errorMatcher: nil,
		},
		{
			name:                  "case 1: simple case EOL",
			inputExistingReleases: nil,
			inputRelease: &releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Name:    "c1",
							Version: "v1",
						},
						{
							Name:    "c2",
							Version: "v2",
						},
					},
				},
				Status: releasev1alpha1.ReleaseStatus{
					Cycle: releasev1alpha1.ReleaseCycleSpec{
						Phase: releasev1alpha1.CyclePhaseEOL,
					},
				},
			},
			expectedComponents: nil,
			errorMatcher:       nil,
		},
		{
			name: "case 2: complex case EOL",
			inputExistingReleases: []*releasev1alpha1.Release{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "r0",
						Labels: map[string]string{
							"release-operator.giantswarm.io/release-cycle-phase": "eol",
						},
					},
					Spec: releasev1alpha1.ReleaseSpec{
						Components: []releasev1alpha1.ReleaseSpecComponent{
							{
								Name:    "c1",
								Version: "v1",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "r1",
						Labels: map[string]string{
							"release-operator.giantswarm.io/release-cycle-phase": "upcoming",
						},
					},
					Spec: releasev1alpha1.ReleaseSpec{
						Components: []releasev1alpha1.ReleaseSpecComponent{
							{
								Name:    "c2",
								Version: "v2",
							},
						},
					},
				},
			},
			inputRelease: &releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Name:    "c1",
							Version: "v1",
						},
						{
							Name:    "c2",
							Version: "v2",
						},
						{
							Name:    "c3",
							Version: "v3",
						},
					},
				},
				Status: releasev1alpha1.ReleaseStatus{
					Cycle: releasev1alpha1.ReleaseCycleSpec{
						Phase: releasev1alpha1.CyclePhaseEOL,
					},
				},
			},
			expectedComponents: []releasev1alpha1.ReleaseSpecComponent{
				{
					Name:    "c2",
					Version: "v2",
				},
			},
			errorMatcher: nil,
		},
		{
			name:                  "case 3: simple marked as deleted",
			inputExistingReleases: nil,
			inputRelease: &releasev1alpha1.Release{
				ObjectMeta: metav1.ObjectMeta{
					// DeletionTimestamp is set so it should be ignored.
					DeletionTimestamp: &metav1.Time{Time: time.Date(2019, 4, 5, 12, 0, 0, 0, time.UTC)},
				},
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Name:    "c1",
							Version: "v1",
						},
						{
							Name:    "c2",
							Version: "v2",
						},
					},
				},
				Status: releasev1alpha1.ReleaseStatus{
					Cycle: releasev1alpha1.ReleaseCycleSpec{
						Phase: releasev1alpha1.CyclePhaseEnabled,
					},
				},
			},
			expectedComponents: nil,
			errorMatcher:       nil,
		},
		{
			name: "case 4: complex marked as deleted",
			inputExistingReleases: []*releasev1alpha1.Release{
				{
					ObjectMeta: metav1.ObjectMeta{
						// DeletionTimestamp is set so this one should be ignored.
						DeletionTimestamp: &metav1.Time{Time: time.Date(2019, 4, 5, 12, 0, 0, 0, time.UTC)},
						Name:              "r0",
						Labels: map[string]string{
							"release-operator.giantswarm.io/release-cycle-phase": "enabled",
						},
					},
					Spec: releasev1alpha1.ReleaseSpec{
						Components: []releasev1alpha1.ReleaseSpecComponent{
							{
								Name:    "c1",
								Version: "v1",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "r1",
						Labels: map[string]string{
							"release-operator.giantswarm.io/release-cycle-phase": "enabled",
						},
					},
					Spec: releasev1alpha1.ReleaseSpec{
						Components: []releasev1alpha1.ReleaseSpecComponent{
							{
								Name:    "c2",
								Version: "v2",
							},
						},
					},
				},
			},
			inputRelease: &releasev1alpha1.Release{
				Spec: releasev1alpha1.ReleaseSpec{
					Components: []releasev1alpha1.ReleaseSpecComponent{
						{
							Name:    "c1",
							Version: "v1",
						},
						{
							Name:    "c2",
							Version: "v2",
						},
						{
							Name:    "c3",
							Version: "v3",
						},
					},
				},
				Status: releasev1alpha1.ReleaseStatus{
					Cycle: releasev1alpha1.ReleaseCycleSpec{
						Phase: releasev1alpha1.CyclePhaseEOL,
					},
				},
			},
			expectedComponents: []releasev1alpha1.ReleaseSpecComponent{
				{
					Name:    "c2",
					Version: "v2",
				},
			},
			errorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx := context.Background()

			var stateGetter *resourceStateGetter
			{
				var objs []runtime.Object
				for _, r := range tc.inputExistingReleases {
					objs = append(objs, r)
				}

				fakeG8sClient := fake.NewSimpleClientset(objs...)

				stateGetter = &resourceStateGetter{
					g8sClient: fakeG8sClient,
					logger:    microloggertest.New(),

					appCatalog: "test-app-catalog",
				}
			}

			components, err := stateGetter.getDesiredComponents(ctx, tc.inputRelease)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(components, tc.expectedComponents) {
				t.Fatalf("\n\n%s\n", cmp.Diff(components, tc.expectedComponents))
			}
		})
	}
}
