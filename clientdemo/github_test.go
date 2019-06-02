
/*
Package clientdemo illustrates how gqlclient can be used to access a github GrapghQL Query API.
*/
package clientdemo

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// The github GraphQL API endpoint URL
const githubAPIURL = "https://api.github.com/graphql"

// Shared function to form a github GraphQL authorization header value from
// a developer's personal access key defined in the `GITHUB_TOKEN` environment
// variable. Fails the test if the key is not present in the environment.
func getAuthorization(t *testing.T) string {

	// Fetch the access key from the environment
	githubToken := os.Getenv("GITHUB_TOKEN")
	if len(githubToken) == 0 {

		// The token is not set! Dang!!
		t.Errorf("\nGITHUB_TOKEN environment variable is not set\n\n")
	}

	// To be assed as an HTTP Authorization header, the access key must be prefixed by "token "
	return "token " + githubToken
}

// TestHappyPath of GetRepoData(...) function to access information about a github project.
func TestHappyPath(t *testing.T) {

	// Get the authorization token from the `GITHUB_TOKEN` environment variable
	authToken := getAuthorization(t)

	// Get the repository data for a public repository
	result, err := GetRepoData(githubAPIURL, authToken, "mikebway", "gogql")
	assert.Nil(t, err, "github graphql invocation should not have failed")

	// Check that the basic values are what we expect them to be
	assert.Equal(t, "gogql", result.Name, "Repository name doees not match")
	assert.Equal(t, "mikebway", result.Owner, "Repository owner doees not match")
	assert.Equal(t, "A basic GraphQL client library for Go", result.Description, "Repository description doees not match")
	expectedCreatedAt, _ := time.Parse(time.RFC3339, "2019-06-01T19:07:06Z")
	assert.Equal(t, expectedCreatedAt, result.CreatedAt, "Repository create time doees not match")
	assert.Equal(t, "Go", result.PrimaryLanguage, "Repository primary language doees not match")
	assert.Equal(t, false, result.IsPrivate, "Repository privacy doees not match")

	// We can't check that the commit data matches what we expect - it will have changed by now - but
	// we do now that there should be five recent commits
	assert.Equal(t, 5, len(result.RecentCommits), "There should have been five recent commits")

	// Confirm that first has a time stamp and a headline message
	assert.NotEmpty(t, result.RecentCommits[0].CommittedAt, "First commit time should be present")
	assert.NotEmpty(t, result.RecentCommits[0].Headline, "First commit headline should be present")
}

// TestInvalidURL examines handling of an invalid github GraphQL API URL
func TestInvalidURL(t *testing.T) {

	// Get the authorization token from the `GITHUB_TOKEN` environment variable
	authToken := getAuthorization(t)

	// Get the repository data for a public repository ... from a bad API URL
	_, err := GetRepoData("http://mikebroadway.com", authToken, "mikebway", "gogql")
	assert.NotEmpty(t, err, "Should not have been able to send a query to https://www.mikebroadway.com")
}

// TestFailedQuery examines handling of an GraphQL reported errors
func TestFailedQuery(t *testing.T) {

	// Get the authorization token from the `GITHUB_TOKEN` environment variable
	authToken := getAuthorization(t)

	// Ask for the repository data for a repository that does not exist
	_, err := GetRepoData(githubAPIURL, authToken, "mikebway", "i-dont-exist")
	assert.NotEmpty(t, err, "GetRepoData should have failed")
	assert.Contains(t, err.Error(), "Errors found in GraphQL Response:", err.Error(), "GetRepoData should have reported GraphQL errors")
}