/*
Package main demonstrates how gqlclient can be used to access a GrapghQL Query API.
*/
package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This file defines unit tests for the main demonstration package

// The github GraphQL API endpoint URL
const testGithubURL = "https://api.github.com/graphql"

// Owner / organization and repository names to use in happy tests
var testOwner = "mikebway"
var testRepoName = "gogql"

// Our exit handling override records the exit codes set by main() here
var exitCodes []int

// We keep the original exit handling function here
var originalExitHandler func(code int)

// Utility function to override the exit handling of the main() function
func overrideFlagsAndExitHandling() {

	// Clear out the command line and any previously defined flags
	os.Args = os.Args[:1]
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Keep the original exit handler on hand
	originalExitHandler = exitDemo

	// Reset any previous exit codes collected
	exitCodes = exitCodes[:0]

	// Set up our own to track the codes we get
	exitDemo = func(code int) {
		exitCodes = append(exitCodes, code)
	}
}

// Restore the original main() exit handler function
func restoreExitHandling() {

	// Put back the originl exit handler
	exitDemo = originalExitHandler
}

// TestUsageOverride executes the main() entry point function to confirm
// that it exetnds the default `flag.Usage()` command line usage function.
func TestUsageOverride(t *testing.T) {

	// Override exit handling from the main() function, restoring after we are done
	overrideFlagsAndExitHandling()
	defer restoreExitHandling()

	// Go does not not support comaprison of function references no matter what trick you try
	// but you can use the runtime package to examine the stack frame. We can confirm that
	// and ovverride that we put in place for flag.Usage() has itself been overriden but
	// comparing the stack frames for multiple invocations of our function.

	// Record how many times our usage function gets invoked
	invocationCount := 0

	// Each time our usage gets invoked, record the file name of the function
	// that called it in this slice
	var callingFiles []string

	// Redfine the `flag.Usage` function to be sonmething we control
	ourUsage := func() {
		invocationCount++
		_, filename, _, _ := runtime.Caller(1)
		callingFiles = append(callingFiles, filename)
	}
	flag.Usage = ourUsage

	// Invoke for the first time directly from this point
	flag.Usage()
	assert.Equal(t, 1, invocationCount, "Our flag.Usage should have been invoked for the first time")
	assert.Contains(t, callingFiles[0], "demo_test.go", "This file should have recorded as the caller")

	// Run the main function
	main()

	// Invoke the usage configured by main()
	flag.Usage()

	// Our usage should have been called a second time but indirectly from the main() file
	assert.Equal(t, 2, invocationCount, "Our flag.Usage should have been invoked for the second time")
	assert.Contains(t, callingFiles[1], "demo.go", "This main() file should have recorded as the caller")

	// Main should have call demoExit() only once and with a code of zero
	assert.Equal(t, 1, len(exitCodes), "exitDemo(n) should only have been called once")
	assert.Equal(t, 0, exitCodes[0], "exit code should be zero")
}

// TestMissingGithubToken confirms that a github access key must be provided
func TestMissingGithubToken(t *testing.T) {

	// Override exit handling from the main() function, restoring after we are done
	overrideFlagsAndExitHandling()
	defer restoreExitHandling()

	// Collect the current github token value and make sure it gets restored after we are done
	githubToken := os.Getenv("GITHUB_TOKEN")
	if len(githubToken) == 0 {
		assert.Fail(t, "the GITHUB_TOKEN must have been set to start with")
	}
	defer os.Setenv("GITHUB_TOKEN", githubToken)

	// Unset the environment variable to upset the demo
	os.Unsetenv("GITHUB_TOKEN")

	// Run the main function
	main()

	// Main should have called demoExit() once with a code of 2
	assert.Equal(t, 1, len(exitCodes), "exitDemo(n) should onl have been called once")
	assert.Equal(t, 2, exitCodes[0], "exit code should be 2 for error handling and showing usage")
}

// Confirm that SSL certificate verification can be dissabled
func TestDisablingCertificateVerification(t *testing.T) {

	// Ensure that the initial confition is disabled and arrange to put it back the way it was
	ourConfig := tls.Config{InsecureSkipVerify: false}
	originalConfig := setTLSClientConfig(&ourConfig)
	defer setTLSClientConfig(originalConfig)

	// Invoke the demo
	err := runDemo(testGithubURL, testOwner, testRepoName, true)
	assert.Nil(t, err, "Should not have been an error running the demo")

	// Confirm that the TLS confoguration has been changed to ignore certificate issues
	insecureSkipVerify := http.DefaultTransport.(*http.Transport).TLSClientConfig.InsecureSkipVerify
	assert.True(t, insecureSkipVerify, "Certificate verification should have been disabled")
}

// Set the SLL.TLS configuration, returing the original TLSClientConfig
// (which may have been nil).
func setTLSClientConfig(newConfig *tls.Config) *tls.Config {

	// Get the current state
	original := http.DefaultTransport.(*http.Transport).TLSClientConfig

	// Set the state to the way we want it
	http.DefaultTransport.(*http.Transport).TLSClientConfig = newConfig

	// Return the orinal status, true if verification was previously disabled
	return original
}
