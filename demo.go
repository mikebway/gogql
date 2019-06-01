/*
Package main demonstrates how gqlclient can be used to access a GrapghQL Query API.
*/
package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"

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
	flag.StringVar(&githubURL, "github", "https://github.com/api/graphql", "URL of the github service GraphQL API")
	flag.StringVar(&repoOwner, "owner", "mikebway", "The organization or user that owns the repository to be evaluated")
	flag.StringVar(&repoName, "name", "ktor-portfolio", "The name of the repository to be evaluated")
	flag.BoolVar(&disableCertificateVerification, "verify", true, "true if to skip SSL certificate verification")
	defaultUsage := flag.Usage
	flag.Usage = func() {
		defaultUsage()
		log.Println()
		log.Println("The GITHUB_TOKEN should be set to a github developer personal access token")
		log.Println("value with sufficient rights to access the values referenced by the")
		log.Println("github.com/mikebway/gogql/github.getRepoDataQuery GraphQL query.")
		log.Println()
	}

	// Parse the command line
	flag.Parse()

	// If we are still here, then the command line did not dissapoint
	// but is the GITHUB_TOKEN environment variable set?
	githubToken := os.Getenv("GITHUB_TOKEN")
	if len(githubToken) == 0 {

		// The token is not set! Dang!!
		flag.Usage()
		log.Printf("ERROR: GITHUB_TOKEN environment is not set\n\n")
	}

	// With the command line understood, now do the actual work of the demonstration
	// If we are to ignore unknown SSL certificate authorities ...
	if disableCertificateVerification {

		// Disable security checks on HTTPS requests
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Have our client demonstration package do the real work
	result, err := clientdemo.GetRepoData(githubURL, githubToken, repoOwner, repoName)
	if err != nil {
		log.Print("\n*** GraphQL Client Demo FAILED ***\n\n")
		log.Fatal(err)
	}

	// From here on, log output to stdout, not stderr
	log.SetOutput(os.Stdout)

	// Log the basic repository data
	log.Printf("\n"+
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
		result.PrimaryLangauge,
		result.IsPrivate)

	// Are there commits to show?

	// And we are done done
	log.Println("\nGraphQL Client Demo finished OK\n ")
}
