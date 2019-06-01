# A Simple GraphQL Client For Go

## Installation

The `gqlclient` package is designed to be used in conjunction with the Go modules dependency
management system introduced in Go 1.11 and 1.12. Simply add the following to the import
block at the head of your source code files:

```go
import (
    "github.com/mikebway/gogql/gqlclient"
)
```

## Querying

The following example code is lifted from the [`demo.go`](/demo.go) demostration application in
the project root and the [`clientdemo/github.go`](/clientdemo/github.go) package code that the
app uses. After pulling down this project, you can run the demonstration app as follows:

```shell
go run demo.go
```

### Declaring the Query Text

GraphQL queries are declared as a multiline strings containing as much newline and indentation formatting
as you like. For example (from [`clientdemo/github.go`](/clientdemo/github.go)):

```go
// The Graphql query we use to retrieve some data about a given repository
const getRepoDataQuery = `query FetchRepoInfo($owner: String!, $name: String!) {
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
```

All of the whitespace will be removed by the client package prior to submission to the GraphQL server;
you don't have to worry about doing that yourself.

### Declaring the Response Data Structure

All GraphQL query responses are recieved as JSON in the following general form:

```json
{
    "data": {
        ... implementation specific JSON here ...
    },
    "errors": [
        {
            "message": "Error message text"
        }
    ]
}
```

The `gqlclient` package models this with the `gqlclient.QueryResponse` type definition:

```go
type QueryResponse struct {
    Data interface {
    } `json:"data"`
    Errors []struct {
        Message string `json:"message"`
    } `json:"errors"`
}
```

As you can see, the `Data` field is declared as an empty interface type. When issuing a query, clients of
the package pass a reference to an instance of `gqlclient.QueryResponse` with the `Data` field pointing to
to a custom structure type that they declare to match the response content that they expect to recieve.

For the github API query shown above, the [`clientdemo/github.go`](/clientdemo/github.go) code builds up its 
`GetRepoDataResponse` data type as follows:

```go
// GetRepoDataResponse is a JSON annotated structure used to parse the response from the GraphQL call into
type GetRepoDataResponse struct {
    Repository struct {
        Name  string `json:"name"`
        Owner struct {
            Login string `json:"login"`
        } `json:"owner"`
        Description     string `json:"description"`
        CreatedAt       string `json:"createdAt"`
        PrimaryLangauge struct {
            Name string `json:"name"`
        } `json:"primaryLangauge"`
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
```

### Invoking the Query

Crudely, without illustrating how paging of alrge result sets might be handled, a query like that 
above can now be invoked and checked as follows:

```go
const githubAPIURL = "https://api.github.com/graphql"

var owner = "mikebway"
var repoName = "ktor-portfolio"

const githubAPIURL = "https://api.github.com/graphql"

var owner = "mikebway"
var repoName = "ktor-portfolio"

func main() {

    // Retrieve the github developer personal access token
    githubToken := os.Getenv("GITHUB_TOKEN")
    if len(githubToken) == 0 {
        log.Fatalf("\nERROR: GITHUB_TOKEN environment is not set\n\n")
    }
    githubAuthorization := "token " + githubToken

    // Disable security checks on HTTPS requests
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

    // Construct a GraphQL client
    client := gqlclient.CreateClient(githubAPIURL, &githubAuthorization)

    // Assemble the query parameters into a map
    queryParms := make(map[string]interface{})
    queryParms["owner"] = owner
    queryParms["name"] = repoName

    // Establish a place to recieve the results of the query
    response := gqlclient.QueryResponse{Data: new(GetRepoDataResponse)}

    // Run the query
    err := client.Query(getRepoDataQuery, &queryParms, &response)
    if err != nil {
        log.Fatalf("\nFAILED: %v\n\n", err)
    }

    // Check the response for errors
    if len(response.Errors) > 0 {

        // Log the errors reported in the response
        for _, err := range response.Errors[0:] {
            log.Fatalf("GraphQL response error: %s", err.Message)
        }

    }

    // All is well, log some of the data retrieved
    repoDataResponse, ok := response.Data.(*GetRepoDataResponse)
    if !ok {
        log.Fatalf("Response did not contain the expected structure")
    }
    repository := repoDataResponse.Repository

    // From here on, log output to stdout, not stderr
    log.SetOutput(os.Stdout)
    log.Printf("\n"+
        "\nRepository Name:               %v"+
        "\nRepository owner/organization  %v"+
        "\nDescription:                   %v"+
        "\nCreated at:                    %v"+
        "\nPrimary language:              %v"+
        "\nIs Private:                    %v",
        repository.Name,
        repository.Owner,
        repository.Description,
        repository.CreatedAt,
        repository.PrimaryLangauge,
        repository.IsPrivate)
}
```

Note that the Go HTTP client does not recognize HomeAway's internal certificate authority, hence the
line at the top that disables TLS certificate verification for the HTTP transport.

## github Authentication (for the demo)

The [github GraphQL API](https://developer.github.com/v4/) requires the provision of an OAuth token
with the right scopes. Fortunately, obtaining a token is straight forward:

1. Loging to github
2. Go to your profile: [https://github.homeawaycorp.com/settings/profile]
3. Click **Developer settings** or go to [https://github.homeawaycorp.com/settings/developers]
4. Click **Personal access tokens** or go to [https://github.homeawaycorp.com/settings/tokens]
5. Click **Generate new token** or go to [https://github.homeawaycorp.com/settings/tokens/new]
6. Give the token a short description that you will be able to recognize in future
7. Check the **repo** box for 'Full control of private repositories'
8. Click **Generate token** button at the bottom of the page
9. **COPY THE TOKEN VALUE AND KEEP IT IN A SAFE AND SECURE PLACE**

Should you need to add more scopes later, you can return to the [tokens page](https://github.homeawaycorp.com/settings/tokens)
at any time and check as many boxes as you like ... but you will **NOT BE ABLE TO SEE THE TOKEN AGAIN**
so make sure you don't skip step 9!!

## Running The Program

The app takes no parameters so ... all you need to do is set access token environment variable 
and run it:

```shell
export GITHUB_TOKEN=cbe9869a0ae552aed6352a188f09370b945e2b21
go run demo.go
```

## Debugging with Visual Studio Code

Change the `"env": {}` line of your `.vscode\launch.json` file to be as follows (with your token value, obviously):

```
"env": {"GITHUB_TOKEN":"cbe9869a0ae552aed6352a188f09370b945e2b21"},
```