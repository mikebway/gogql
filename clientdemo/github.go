/*
Package clientdemo illustrates how gqlclient can be used to access a github GrapghQL Query API.
*/
package clientdemo

import (
	"errors"
	"strings"
	"time"

	"github.com/mikebway/gogql/gqlclient"
)

// RepoCommit is a structure type that represents a single commit to a github repository
type RepoCommit struct {
	CommittedAt time.Time // The data and time at which the commit was made
	Headline    string    // The headlin explanation of why the commit was made
}

// RepoData is a structure used to return information about a single github repository.
type RepoData struct {
	Name            string       // The repository name
	Owner           string       // The user or organization that owns the repository
	Description     string       // The short description of the repository
	CreatedAt       time.Time    // The date and time at which the repository was created
	PrimaryLanguage string       // The language used for most of the code in the repository
	IsPrivate       bool         // true if the repository is private to the owner
	RecentCommits   []RepoCommit // A list of the most recent commits (if any)
}

// The Graphql query we use to retrieve some data about a given repository
var getRepoDataQuery = `query FetchRepoInfo($owner: String!, $name: String!) {
	repository(owner: $owner, name: $name) {
	  name
	  owner {
			login
	  }
	  description
	  createdAt
	  primaryLanguage {
			name
	  }
	  diskUsage
	  isPrivate
	  ref(qualifiedName: "master") {
			target {
		  	... on Commit {
					history(first: 5) {
						edges {
							node {
								committedDate
								messageHeadline
							}
						}
					}
				}
			}
		}
	}
}`

// GetRepoDataResponse is a JSON annotated structure used to parse the response from the GraphQL call into
type GetRepoDataResponse struct {
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Description     string `json:"description"`
		CreatedAt       string `json:"createdAt"`
		PrimaryLanguage struct {
			Name string `json:"name"`
		} `json:"primaryLanguage"`
		IsPrivate bool `json:"isPrivate"`
		Ref       struct {
			Target struct {
				History struct {
					Edges []struct {
						Node struct {
							CommittedDate   string `json:"committedDate"`
							MessageHeadline string `json:"messageHeadline"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"history"`
			} `json:"target"`
		} `json:"ref"`
	} `json:"repository"`
}

// GetRepoData serves the dual purpose of illustrating the use of the GraphQL
// client and getting line coverage up when called from a unit test by retrieving
// a few bits of data about a given repository.
func GetRepoData(githubAPIURL string, githubToken string, owner string, repoName string) (*RepoData, error) {

	// Construct a GraphQL client
	client := gqlclient.CreateClient(githubAPIURL, &githubToken)

	// Assemble the query parameters into a map
	queryParms := make(map[string]interface{})
	queryParms["owner"] = &owner
	queryParms["name"] = &repoName

	// Establish a place to recieve the results of the query
	response := gqlclient.QueryResponse{Data: new(GetRepoDataResponse)}

	// Run the query
	err := client.Query(&getRepoDataQuery, &queryParms, &response)
	if err != nil {
		return nil, err
	}

	// Were there any errors reported by the GraphQL service itself?
	if response.Errors != nil {

		// 	Assemble the error messages into a single string
		var sb strings.Builder
		sb.WriteString("Errors found in GraphQL Response:\n\n")
		for _, e := range response.Errors {
			sb.WriteString(e.Message)
			sb.WriteString("\n")
		}

		// Report this back to the caller
		return nil, errors.New(sb.String())
	}

	// All is well, translate the query response into our simpler result structure
	repoDataResponse, ok := response.Data.(*GetRepoDataResponse)
	if !ok {
		return nil, errors.New("Response did not contain the expected structure")
	}
	repository := repoDataResponse.Repository
	result := &RepoData{
		Name:            repository.Name,
		Owner:           repository.Owner.Login,
		Description:     repository.Description,
		PrimaryLanguage: repository.PrimaryLanguage.Name,
		IsPrivate:       repository.IsPrivate,
	}

	// The other stuff is more fiddly: parse the repo creation time
	result.CreatedAt, _ = time.Parse(time.RFC3339, repository.CreatedAt)

	// Loop over the commit messages
	for _, c := range repository.Ref.Target.History.Edges {
		committedDate, _ := time.Parse(time.RFC3339, c.Node.CommittedDate)
		result.RecentCommits = append(result.RecentCommits, RepoCommit{
			CommittedAt: committedDate,
			Headline:    c.Node.MessageHeadline,
		})
	}

	// And we are all done, return the result
	return result, nil
}
