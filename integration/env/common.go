package env

import (
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
	return strings.ToLower(circleCI) =="true"
}

func CircleSHA() string {
	return circleSHA
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
