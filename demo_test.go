/*
Package main demonstrates how gqlclient can be used to access a GrapghQL Query API.
*/
package main

import (
	"flag"
	"testing"
)

// This file defines unit tests for the main demonstration package

// TestHappyPath of main() entry point function.
func TestHappyPath(t *testing.T) {

	// Run the main function
	main()

	// Display the usage configured by main()
	flag.Usage()
}
