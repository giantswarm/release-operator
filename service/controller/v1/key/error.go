package key

import "github.com/giantswarm/microerror"

var emptyValueError = &microerror.Error{
	Kind: "emptyValueError",
}

// IsEmptyValueError asserts emptyValueError.
func IsEmptyValueError(err error) bool {
	return microerror.Cause(err) == emptyValueError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
