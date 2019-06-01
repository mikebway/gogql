/*
Package gplclient is a simple client package for accessing GrpapQL APIs.
This file contains unit test code for gplclient.
*/
package gplclient

import (
	"testing"
)

// This file defines unit tests for the gplclient library.

// TestPackQuery exercises the utility method that is used to reduce easily read, multi-line
// GraphQL expressions down to simple strings with no excess whitepsace or formatting ready
// for transmission as JSON values.
func TestPackQuery(t *testing.T) {

	input := "\n\tthis     \t    is a \n\t test of     whitespace\n\t\n\tremoval\n\n"
	expected := "this is a test of whitespace removal"
	output := packQuery(input)
	if output != expected {
		t.Errorf("Expected [%s] but found [%s]", expected, output)
	}
}
