/*
Package main demonstrates how gqlclient can be used to access a GrapghQL Query API.
*/
package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mikebway/gogql/clientdemo"
)

// URL of the github service GraphQL API; set by command line flag
var githubURL string

// The name of the environment variable to be used to load the github access token
var tokenVarName = "GITHUB_TOKEN"

// True if to disable SSL certificate verification; set by command line flag
var disableCertificateVerification bool

// The organization or user that owns the repository to be evaluated
var repoOwner string

// The name of the repository to be evaluated
var repoName string

// We allow unti testing to override program exit handling
var exitDemo = func(code int) {
	os.Exit(code)
}

// This demonstration app retrieves the github data for a specified project,
// defaulting to this project itself.
func main() {

	// Declare our command line flags
	flag.StringVar(&githubURL, "github", "https://api.github.com/graphql", "URL of the github service GraphQL API")
	flag.StringVar(&tokenVarName, "token-env", "GITHUB_TOKEN", "The name of the environment variable that provides the github access token")
	flag.StringVar(&repoOwner, "owner", "mikebway", "The organization or user that owns the repository to be evaluated")
	flag.StringVar(&repoName, "name", "gogql", "The name of the repository to be evaluated")
	flag.BoolVar(&disableCertificateVerification, "skipverify", false, "Use to to skip SSL certificate verification")
	defaultUsage := flag.Usage
	flag.Usage = func() {
		defaultUsage()
		fmt.Println()
		fmt.Println("The GITHUB_TOKEN enironment variable should be set to a github developer")
		fmt.Println("personal access token value with sufficient rights to access the values")
		fmt.Println("referenced by the github.com/mikebway/gogql/github.getRepoDataQuery GraphQL")
		fmt.Println("query.")
		fmt.Println()
		fmt.Println("You can use the -token-env command line flag to override the name of the")
		fmt.Println("envvironment variable and so support more than one token value for multiple")
		fmt.Println("github services (i.e. public and corporate).")
	}

	// Parse the command line. Note that we have to pass the arguments because we are
	// not useing the default flags.Parse() function.
	flag.Parse()

	// For the sake of easier unit testing, separate the actual work of the demo into
	// parameterized function. Likewise, we don't use os.Exit(n) directly so that
	// unit tests can oveeride that behavior
	err := runDemo(githubURL, repoOwner, repoName, disableCertificateVerification)
	if err != nil {
		fmt.Printf("GraphQL Client Demo FAILED:\n\n %v\n\n", err)
		flag.Usage()
		exitDemo(2)
	} else {
		fmt.Println("\n\nGraphQL Client Demo finished OK\n ")
		exitDemo(0)
	}
}

// Do the actual work of the demo as a function that can be more easily unit tested
func runDemo(githubURL, repoOwner, repoName string, disableCertificateVerification bool) error {

	// Is the GITHUB_TOKEN environment variable set?
	githubToken := os.Getenv(tokenVarName)
	if len(githubToken) == 0 {

		// The token is not set! Dang!!
		msg := fmt.Sprintf("the %s environment variable is not set", tokenVarName)
		return errors.New(msg)
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
		return err
	}

	// Log the basic repository data
	fmt.Printf("\n"+
		"\nRepository Name:               %v"+
		"\nRepository owner/organization  %v"+
		"\nDescription:                   %v"+
		"\nCreated at:                    %v"+
		"\nPrimary language:              %v"+
		"\nDisk usage (K):                %v"+
		"\nIs Private:                    %v",
		result.Name,
		result.Owner,
		result.Description,
		result.CreatedAt,
		result.PrimaryLanguage,
		result.DiskUsage,
		result.IsPrivate)

	// Are there commits to show? There must be at least one in any non-virgin repo!
	fmt.Println("\nMost recent commits:")
	for _, c := range result.RecentCommits {
		fmt.Printf("  %s\n    %s\n", c.CommittedAt.Format(time.RFC1123), c.Headline)
	}

	// And we are done done
	return nil
}
