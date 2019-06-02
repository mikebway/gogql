/*
Package gqlclient is a simple client package for accessing GrpapQL APIs.
This file contains unit test code for gqlclient.
*/
package gqlclient

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This file defines unit tests for the gqlclient library.

// The github GraphQL API endpoint URL
const githubAPIURL = "https://api.github.com/graphql"

// Owner / organization and repository names to use in happy tests
var owner = "mikebway"
var repoName = "gogql"

// The Graphql query we use to retrieve some data about a given repository
var SimpleRepoDataQuery = `query FetchRepoInfo($owner: String!, $name: String!) {
	repository(owner: $owner, name: $name) {
		name
		owner {
			login
		}
	}
}`

// GetRepoDataResponse is a JSON annotated structure used to parse the response from the GraphQL call into
type SimpleRepoDataResponse struct {
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

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

// TestPackQuery exercises the utility method that is used to reduce easily read, multi-line
// GraphQL expressions down to simple strings with no excess whitepsace or formatting ready
// for transmission as JSON values.
func TestPackQuery(t *testing.T) {

	input := "\n\tthis     \t    is a \n\t test of     whitespace\n\t\n\tremoval\n\n"
	expected := "this is a test of whitespace removal"
	output := packQuery(&input)
	assert.Equal(t, expected, output, "Query packing gave unexpected result")
}

// TestHappyPath uses the `clientdemo.GetRepoData(...)` function to access information about a github project.
func TestHappyPath(t *testing.T) {

	// Get the authorization tokne from the `GITHUB_TOKEN` environment variable
	authToken := getAuthorization(t)

	// Construct a GraphQL client
	client := CreateClient(githubAPIURL, &authToken)

	// Assemble the query parameters into a map
	queryParms := make(map[string]interface{})
	queryParms["owner"] = &owner
	queryParms["name"] = &repoName

	// Establish a place to recieve the results of the query
	response := QueryResponse{Data: new(SimpleRepoDataResponse)}

	// Get the repository data for a public repository
	err := client.Query(&SimpleRepoDataQuery, &queryParms, &response)
	if err != nil {
		t.Errorf("Happy path invocation failed: %v", err)
	}

	// There should be no errors reported in the GraphQL response
	assert.Empty(t, response.Errors, "There should be no GraphQL reported errors")

	// Check the values that we got back
	repoDataResponse, ok := response.Data.(*SimpleRepoDataResponse)
	assert.True(t, ok, "Response did not contain the expected structure")
	repository := repoDataResponse.Repository
	assert.Equal(t, owner, repository.Owner.Login)
	assert.Equal(t, repoName, repository.Name)
}
