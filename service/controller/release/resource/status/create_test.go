package status

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var testClusters = []tenantCluster{
	{
		ID:               "abc12",
		OperatorVersion:  "1.2.3",
		ProviderOperator: "test-operator",
		ReleaseVersion:   "9.8.7",
	},
	{
		ID:               "def34",
		OperatorVersion:  "4.5.6",
		ProviderOperator: "test-operator",
		ReleaseVersion:   "8.7.6",
	},
	// These next two clusters differ only in their IDs.
	// The versions should be deduplicated in the result map.
	{
		ID:               "ghi56",
		OperatorVersion:  "1.2.1-4",
		ProviderOperator: "another-operator",
		ReleaseVersion:   "9.8.9-1",
	},
	{
		ID:               "jkl78",
		OperatorVersion:  "1.2.1-4",
		ProviderOperator: "another-operator",
		ReleaseVersion:   "9.8.9-1",
	},
}

func Test_consolidateClusterVersions(t *testing.T) {
	testCases := []struct {
		name              string
		clusters          []tenantCluster
		expectedReleases  map[string]bool
		expectedOperators map[string]map[string]bool
	}{
		{
			name:     "case 0: release and operator versions are all present in map",
			clusters: testClusters,
			expectedReleases: map[string]bool{
				"9.8.7":   true,
				"8.7.6":   true,
				"9.8.9-1": true,
			},
			expectedOperators: map[string]map[string]bool{
				"test-operator": map[string]bool{
					"1.2.3": true,
					"4.5.6": true,
				},
				"another-operator": map[string]bool{
					"1.2.1-4": true,
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			resultReleases, resultOperators := consolidateClusterVersions(tc.clusters)
			if !cmp.Equal(resultReleases, tc.expectedReleases) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedReleases, resultReleases))
			}
			if !cmp.Equal(resultOperators, tc.expectedOperators) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedOperators, resultOperators))
			}
		})
	}
}
