package apps

import (
	"strings"

	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

// IsResourceNotFound asserts resource not found error from the Kubernetes API.
func IsResourceNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(microerror.Cause(err).Error(), "the server could not find the requested resource")
}
