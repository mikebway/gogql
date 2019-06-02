# A Simple GraphQL Client For Go

## Early Stage Project

As the project stands at this stage, the "client" only supports queries and not mutations
or subscriptions. Whether I wil take it much beyond this point I am not yet sure. It was
a useful learning exercise to create it but there are perhaps more complete implementations
availble form other sources that make further effor moot?

## Installation

The `gqlclient` package is designed to be used in conjunction with the Go modules dependency
management system introduced in Go 1.11 and 1.12. Simply add the following to the import
block at the head of your source code files:

```go
import (
    "github.com/mikebway/gogql/gqlclient"
)
```

## Unit Testing

Note that to run the unit tests, you must have set the `GITHUB_TOKEN` environment variable
with a github 'developer access token'. Instructions for creating such a token for your
github login are described at the very bottom of this page.

Also at the bottom of the page are instructions for making the `GITHUB_TOKEN` environment
variable available for running the demo app under the VSCode IDE and make it available
for unit test execution with VSCode.

## Querying

The code examples below are lifted from the [`demo.go`](/demo.go) demostration application in
the project root and the [`clientdemo/github.go`](/clientdemo/github.go) package code that the
app uses. After pulling down this project, you can run the demonstration app as follows:

```shell
go run demo.go
```

Add a `-h` flag to display the following command line usage information:

```text
Usage of /var/folders/b6/n6cvhy2d6455c88qks3cxkr00000gn/T/go-build099997439/b001/exe/demo:
  -github string
        URL of the github service GraphQL API (default "https://api.github.com/graphql")
  -name string
        The name of the repository to be evaluated (default "gogql")
  -owner string
        The organization or user that owns the repository to be evaluated (default "mikebway")
  -verify
        true if to skip SSL certificate verification (default true)

The GITHUB_TOKEN should be set to a github developer personal access token
value with sufficient rights to access the values referenced by the
github.com/mikebway/gogql/github.getRepoDataQuery GraphQL query.
```

Instructions for creating such a `GITHUB_TOKEN` for your github login are described at the very
bottom of this page.

### Disabling TLS / SSL Certificate Validation

Normally, you would not need or want to skip validation of SSL certificates but it is not uncommon
in development enviroments for custom certificates to have been used that are not backed by
a certificate authority that the Go enviroment is aware of. The quicketst way to work around,
as illustrated by the [`demo.go`](/demo.go) app, is as follows:

```go
    // If we are to ignore unknown SSL certificate authorities ...
    if disableCertificateVerification {

        // Disable security checks on HTTPS requests
        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
    }
```

### Declaring the Query Text

GraphQL queries are declared as a multiline strings containing as much newline and indentation formatting
as you like. For example (from [`clientdemo/github.go`](/clientdemo/github.go)):

```go
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
```

### Invoking the Query

The [`demo.go`](/demo.go) demostration application in the project root is slightly more sophisticated but,
crudely, without illustrating how paging of large result sets might be handled, a query like that above
can be invoked and checked as follows:

```go
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

    // Construct a GraphQL client
    client := gqlclient.CreateClient(githubAPIURL, &githubAuthorization)

    // Assemble the query parameters into a map
    queryParms := make(map[string]interface{})
    queryParms["owner"] = owner
    queryParms["name"] = repoName

    // Establish a place to recieve the results of the query
    response := gqlclient.QueryResponse{Data: new(GetRepoDataResponse)}

    // Run the query
    err := client.Query(&getRepoDataQuery, &queryParms, &response)
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
        repository.PrimaryLanguage,
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

The demo app defaults all parameters so all you need to do is set access token environment variable
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

## Unit Testing with Macs and VSCode (or similar)

If you are trying to run unit tests on a Mac from within an IDE that does not give you the
to set environment variables for unit test execution, you can workaround the problem by 
setting them in a `.plist` file as follows:

1. Create a text file named `~/Library/LaunchAgents/githubtoken.plist`

2. Use a text editor to write the XML text below to the file, substituting your personal access token
for the `cbe9869a0...b945e2b21` value.

```xml
<?xml version="1.0" encoding="UTF-8"?>

<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
  <plist version="1.0">
  <dict>
  <key>Label</key>
  <string>githubtoken</string>
  <key>ProgramArguments</key>
  <array>
    <string>/bin/launchctl</string>
    <string>setenv</string>
    <string>GITHUB_TOKEN</string>
    <string>cbe9869a0...b945e2b21</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
</dict>
</plist>
```

3. Run the following command in a terminal window:

```shell
launchctl load ~/Library/LaunchAgents/githubtoken.plist
```

4. The `GITHUB_TOKEN` environmnet variable will now be accessible from all pplications and 
terminal shell contexts: exit and relaunch VSCode to have it available to the unit tests
from within the IDE.

Unfortunately, this means that it is permanently visible to anyone that runs 
`echo $GITHUB_TOKEN` on your Mac. The only way to remove it is to delete it
is to run

```shell
launchctl unload ~/Library/LaunchAgents/githubtoken.plist
```

and then delete the file.
