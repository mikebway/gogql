/*
Package main demonstrates how gqlclient can be used to access a GrapghQL Query API.
*/
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mikebway/gogql/clientdemo"
)

// URL of the github service GraphQL API; set by command line flag
var githubURL string

// True if to disable SSL certificate verification; set by command line flag
var disableCertificateVerification bool

// The organization or user that owns the repository to be evaluated
var repoOwner string

// The name of the repository to be evaluated
var repoName string

// This demonstration app retrieves the github data for a specified project,
// defaulting to this project itself.
func main() {

	// Declare our command line flags
	flag.StringVar(&githubURL, "github", "https://api.github.com/graphql", "URL of the github service GraphQL API")
	flag.StringVar(&repoOwner, "owner", "mikebway", "The organization or user that owns the repository to be evaluated")
	flag.StringVar(&repoName, "name", "gogql", "The name of the repository to be evaluated")
	flag.BoolVar(&disableCertificateVerification, "skipverify", false, "Use to to skip SSL certificate verification")
	defaultUsage := flag.Usage
	flag.Usage = func() {
		defaultUsage()
		fmt.Println()
		fmt.Println("The GITHUB_TOKEN should be set to a github developer personal access token")
		fmt.Println("value with sufficient rights to access the values referenced by the")
		fmt.Println("github.com/mikebway/gogql/github.getRepoDataQuery GraphQL query.")
		fmt.Println()
	}

	// Parse the command line
	flag.Parse()

	// If we are still here, then the command line did not dissapoint
	// but is the GITHUB_TOKEN environment variable set?
	githubToken := os.Getenv("GITHUB_TOKEN")
	if len(githubToken) == 0 {

		// The token is not set! Dang!!
		fmt.Printf("\nERROR: GITHUB_TOKEN environment variable is not set\n\n")
		flag.Usage()
		os.Exit(2)
	}

	// Passed as an HTTP Authorization header, the token value must be prefixed by "token "
	githubAuthorization := "token " + githubToken

	// With the command line understood, now do the actual work of the demonstration
	// If we are to ignore unknown SSL certificate authorities ...
	if disableCertificateVerification {

		// Disable security checks on HTTPS requests
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Have our client demonstration package do the real work
	result, err := clientdemo.GetRepoData(githubURL, githubAuthorization, repoOwner, repoName)
	if err != nil {
		log.Print("\n*** GraphQL Client Demo FAILED ***\n\n")
		log.Fatal(err)
	}

	// Log the basic repository data
	fmt.Printf("\n"+
		"\nRepository Name:               %v"+
		"\nRepository owner/organization  %v"+
		"\nDescription:                   %v"+
		"\nCreated at:                    %v"+
		"\nPrimary language:              %v"+
		"\nIs Private:                    %v",
		result.Name,
		result.Owner,
		result.Description,
		result.CreatedAt,
		result.PrimaryLanguage,
		result.IsPrivate)

	// Are there commits to show? There must be at least one in any non-virgin repo!
	fmt.Println("\nMost recent commits:")
	for _, c := range result.RecentCommits {
		fmt.Printf("  %s\n    %s\n", c.CommittedAt.Format(time.RFC1123), c.Headline)
	}

	// And we are done done
	fmt.Println("\n\nGraphQL Client Demo finished OK\n ")
}
