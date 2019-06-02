/*
Package gqlclient is a simple client package for accessing GrpapQL APIs.
*/
package gqlclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// GqlClient is a structure/class that is used to wrap configuration data including the
// target GraphQL URL and any autheorization token that may be required. Queries are
// invoked through methods associated with this structure type.
//
// Valid GqlClient instances can only be obtained through the CreateClient(...) function.
type GqlClient struct {
	targetURL     string  // The GraphQL server URL, e.g. https://api.github.com/graphql
	authorization *string // If not nil, the authoorization header value to be supplied with GraphQL calls
}

// CreateClient returns a reference to an initialized GqlClient instance. The target URL for the
// GraphQL must be provided. The authorization string my be nil if no token or basic auth header
// is required by the server. A typical authirization value for a target URL, say, https://api.github.com/graphql
// the authorization value would be of the form "token f69acf817105a9e024f3e94a80bbf09e2879abef". Note that
// the authorization value is write only - once set in the GqlClient it cannot be accessed outside of the
// `gqlclient` package. While the targetURL can be retrieved vai the GetTargetURL() function, it cannot be
// modified.
func CreateClient(targetURL string, authorization *string) *GqlClient {
	return &GqlClient{targetURL, authorization}
}

// GetTargetURL returns the target API URL of the GqlClient.
func (gc *GqlClient) GetTargetURL() string {
	return gc.targetURL
}

// QueryResponse is a structure pattern that should be followed by all response structures provided to the
// gqlclient.Query(...) method. Package clients should set the Data variable to point to a struture instance
// that has been declared to match the expected JSON result of the query. For example:
//
// 		res := gqlclient.QueryResponse{Data: new(RepositorySearch)}
//
type QueryResponse struct {
	Data interface {
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Query sends a GraphQL query string to the given URL and parses the response into the provided object reference.
// An error is returned if any showstopping problem occurs.
//
// The query string may be formatted with whitespace and carriage returns for readbility, any such whitespace shall
// be removed prior to submission to the GraphQL server. The queryParms may be nil if the query does not require
// any parameters.
func (gc *GqlClient) Query(queryStr *string, queryParms *map[string]interface{}, response *QueryResponse) error {

	// Build the GraphQL query into JSON that we can POST
	q := query{packQuery(queryStr), *queryParms}
	queryBytes, err := json.Marshal(q)
	if err != nil {
		return err
	}

	// Form up an HTTP POST request, supplying the github access token
	req, _ := http.NewRequest("POST", gc.targetURL, bytes.NewReader(queryBytes))
	req.Header.Set("Content-Type", "application/json")
	if gc.authorization != nil {
		req.Header.Add("Authorization", *gc.authorization)
	}

	// Submit the POST and wait for the response
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If the response status code is not 200, report an error
	if resp.StatusCode != 200 {
		if resp.StatusCode == 401 {
			return errors.New("Recieved 401 UNAUTHORIZED response! Did you need to provide an authorization key?")
		}
		return errors.New("Expected 200 response but received: " + resp.Status)
	}

	// Load the raw response body
	body, _ := ioutil.ReadAll(resp.Body)

	// Unmarshal the response into the provided object
	return json.Unmarshal(body, &response)
}

// packQuery strips whitespace and newlines from a formatted GraphQL query.
func packQuery(str *string) string {

	// Reduce all whitespace character sequences to single spaces
	return strings.Join(strings.Fields(*str), " ")
}

// For GraphQL over HTTP 1.1, the query and its parameters must be wrapped in a JSON object.
type query struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// httpClient is a package scoped http client declaration that can be overriden by unit tests
// to mock up various error conditions.
var httpClient = &http.Client{
	Timeout: time.Second * 10,
}
