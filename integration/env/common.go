package env

import (
	"crypto/sha1"
	"fmt"
	"os"
	"strings"
)

const (
	EnvVarCircleCI           = "CIRCLECI"
	EnvVarCircleSHA          = "CIRCLE_SHA1"
	EnvVarE2EKeepResources   = "E2E_KEEP_RESOURCES"
	EnvVarE2ETestDir         = "E2E_TEST_DIR"
	EnvVarGithubBotToken     = "GITHUB_BOT_TOKEN"
	EnvVarRegistryPullSecret = "REGISTRY_PULL_SECRET"
)

var (
	circleCI           string
	circleSHA          string
	githubToken        string
	keepResources      string
	registryPullSecret string
	testDir            string
)

func init() {
	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarE2EKeepResources)

	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}

	testDir = os.Getenv(EnvVarE2ETestDir)
	if testDir == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarE2ETestDir))
	}

	githubToken = os.Getenv(EnvVarGithubBotToken)
	if githubToken == "" {
		panic(fmt.Sprintf("env var %q must not be empty", EnvVarGithubBotToken))
	}

	registryPullSecret = os.Getenv(EnvVarRegistryPullSecret)
	if registryPullSecret == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarRegistryPullSecret))
	}
}

func CircleCI() bool {
	return circleCI == strings.ToLower("true")
}

func CircleSHA() string {
	return circleSHA
}

// ClusterID returns a cluster ID unique to a run integration test. It might
// look like ci-3cc75-5e958.
//
//     - ci is a static identifier stating a CI run.
//     - 3cc75 is the Git SHA.
//     - 5e958 is a hash of the integration test dir.
//
// NOTE: If a test requires multiple clusters this value should be used as
// a cluster ID prefix.
func ClusterID() string {
	var parts []string

	var testHash string
	{
		h := sha1.New()
		h.Write([]byte(TestDir()))
		testHash = fmt.Sprintf("%x", h.Sum(nil))[0:5]
	}

	parts = append(parts, "ci")
	parts = append(parts, CircleSHA()[0:5])
	parts = append(parts, testHash)

	return strings.Join(parts, "-")
}

func KeepResources() bool {
	return keepResources == strings.ToLower("true")
}

func GithubToken() string {
	return githubToken
}

func RegistryPullSecret() string {
	return registryPullSecret
}

func TestDir() string {
	return testDir
}
