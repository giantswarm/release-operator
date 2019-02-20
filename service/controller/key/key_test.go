package key

import (
	"testing"
)

func TestSplitReleaseName(t *testing.T) {
	testCases := []struct {
		name             string
		input            string
		expectedProvider string
		expectedVersion  string
		errorMatcher     func(err error) bool
	}{
		{
			name:             "case 0: valid semver",
			input:            "aws.v6.0.1",
			expectedProvider: "aws",
			expectedVersion:  "v6.0.1",
			errorMatcher:     nil,
		},
		{
			name:             "case 1: valid any version",
			input:            "aws.any",
			expectedProvider: "aws",
			expectedVersion:  "any",
			errorMatcher:     nil,
		},
		{
			name:             "case 1: empty provider",
			input:            ".v6.0.1",
			expectedProvider: "",
			expectedVersion:  "",
			errorMatcher:     IsInvalidReleaseName,
		},
		{
			name:             "case 2: empty version",
			input:            "aws.",
			expectedProvider: "",
			expectedVersion:  "",
			errorMatcher:     IsInvalidReleaseName,
		},
		{
			name:             "case 3: invalid name",
			input:            "foo",
			expectedProvider: "",
			expectedVersion:  "",
			errorMatcher:     IsInvalidReleaseName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider, version, err := SplitReleaseName(tc.input)
			t.Logf("provider=%#q  version=%#q  err=%v", provider, version, err)

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

			if provider != tc.expectedProvider {
				t.Fatalf("wrong provider: got %#q, want %#q", provider, tc.expectedProvider)
			}

			if version != tc.expectedVersion {
				t.Fatalf("wrong version: got %#q, want %#q", version, tc.expectedVersion)
			}
		})
	}
}
