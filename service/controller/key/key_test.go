package key

import (
	"strconv"
	"testing"

	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/google/go-cmp/cmp"
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
		Version: "7.0.0",
	},
}

func Test_ExtractAllOperators(t *testing.T) {
	testCases := []struct {
		name              string
		releases          releasev1alpha1.ReleaseList
		expectedOperators map[string]releasev1alpha1.ReleaseSpecComponent
	}{
		{
			name: "case 0: extracts all operators",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testOperators[0],
								testOperators[1],
							},
						},
					},
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testOperators[2],
							},
						},
					},
				},
			},
			expectedOperators: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testOperators[0]): testOperators[0],
				BuildAppName(testOperators[1]): testOperators[1],
				BuildAppName(testOperators[2]): testOperators[2],
			},
		},
		{
			name: "case 1: does not extract non-operators",
			releases: releasev1alpha1.ReleaseList{
				Items: []releasev1alpha1.Release{
					{
						Spec: releasev1alpha1.ReleaseSpec{
							Components: []releasev1alpha1.ReleaseSpecComponent{
								testOperators[0],
								{
									Name:    "something-else",
									Version: "9.0.0",
								},
							},
						},
					},
				},
			},
			expectedOperators: map[string]releasev1alpha1.ReleaseSpecComponent{
				BuildAppName(testOperators[0]): testOperators[0],
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultOperators := ExtractAllOperators(tc.releases)

			if !cmp.Equal(resultOperators, tc.expectedOperators) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOperators, resultOperators))
			}
		})
	}
}
