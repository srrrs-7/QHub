package testutil

import "net/http"

const TestBearerToken = "test-bearer-token"

// SetAuthHeader sets the Authorization header with a Bearer token for testing
func SetAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+TestBearerToken)
}
