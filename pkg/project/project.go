package project

var (
	description = "The release-operator manages chart configs for new releases."
	gitSHA      = "n/a"
	name        = "release-operator"
	source      = "https://github.com/giantswarm/release-operator"
	version     = "4.2.2-dev"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
