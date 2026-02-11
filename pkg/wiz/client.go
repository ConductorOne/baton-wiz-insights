package wiz

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// principalEntityTypes are the Wiz entity types that represent user and service
// accounts — the kinds of identities that match users synced from other Baton
// connectors (baton-aws, baton-github, baton-okta, etc.).
// See https://docs.wiz.io/dev/sec-graph-object-normalization for all entity types.
var principalEntityTypes = []string{
	"USER_ACCOUNT",
	"SERVICE_ACCOUNT",
}

// Client defines the interface for interacting with the Wiz API.
type Client interface {
	ListIssues(ctx context.Context, cursor *string) (*IssueConnection, error)
	ListIssuesSince(ctx context.Context, since time.Time, cursor *string) (*IssueConnection, error)
}

// client implements the Client interface.
type client struct {
	wrapper *uhttp.BaseHttpClient
	apiURL  string
}

// NewClient creates a new Wiz API client with OAuth2 authentication.
func NewClient(ctx context.Context, apiURL, clientID, clientSecret, authEndpoint string) (Client, error) {
	// Configure OAuth2 client credentials flow
	// Wiz requires the "audience=wiz-api" parameter for token requests
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     authEndpoint,
		AuthStyle:    oauth2.AuthStyleInParams,
		EndpointParams: map[string][]string{
			"audience": {"wiz-api"},
		},
	}

	// Create an HTTP client that automatically handles token management
	httpClient := config.Client(ctx)

	// Wrap with baton-sdk's HTTP client wrapper for proper error handling and retries
	wrapper, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client wrapper: %w", err)
	}

	return &client{
		wrapper: wrapper,
		apiURL:  apiURL,
	}, nil
}

// graphQLRequest makes a GraphQL request to the Wiz API using baton-sdk's HTTP wrapper.
// The wrapper handles retries, rate limiting, and error wrapping automatically.
func (c *client) graphQLRequest(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	// Parse the API URL
	parsedURL, err := url.Parse(c.apiURL)
	if err != nil {
		return fmt.Errorf("failed to parse API URL: %w", err)
	}

	req, err := c.wrapper.NewRequest(ctx, http.MethodPost, parsedURL, uhttp.WithJSONBody(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Use a temporary struct to capture the GraphQL response envelope
	var gqlResp graphQLResponse
	gqlResp.Data = result

	// Execute the request with JSON response handling
	resp, err := c.wrapper.Do(req, uhttp.WithJSONResponse(&gqlResp))
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	// Check for GraphQL-specific errors in the response
	if len(gqlResp.Errors) > 0 {
		return status.Errorf(codes.Unknown, "graphql errors: %+v", gqlResp.Errors)
	}

	return nil
}

const issuesQuery = `query IssuesV2($after: String, $first: Int, $filterBy: IssueFilters) {
  issuesV2(after: $after, first: $first, filterBy: $filterBy) {
    nodes {
      id
      status
      severity
      createdAt
      statusChangedAt
      sourceRule {
        id
        name
      }
      entitySnapshot {
        id
        type
        name
        nativeType
        externalId
        cloudPlatform
        subscriptionId
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}`

// principalEntityFilter returns the relatedEntity filter that restricts
// results to only principal/identity entity types.
func principalEntityFilter() map[string]interface{} {
	return map[string]interface{}{
		"relatedEntity": map[string]interface{}{
			"type": principalEntityTypes,
		},
	}
}

// ListIssues retrieves a paginated list of principal-related issues from Wiz.
// Results are filtered to only issues whose related entity is a principal type
// (USER, SERVICE_ACCOUNT, ACCESS_ROLE, ACCESS_ROLE_BINDING, IDENTITY).
func (c *client) ListIssues(ctx context.Context, cursor *string) (*IssueConnection, error) {
	variables := map[string]interface{}{
		"first":    100,
		"filterBy": principalEntityFilter(),
	}
	if cursor != nil && *cursor != "" {
		variables["after"] = *cursor
	}

	var result issuesQueryResponse
	if err := c.graphQLRequest(ctx, issuesQuery, variables, &result); err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	return &result.IssuesV2, nil
}

// ListIssuesSince retrieves a paginated list of principal-related issues from Wiz
// filtered by statusChangedAt >= since. Used by the event feed for incremental sync.
func (c *client) ListIssuesSince(ctx context.Context, since time.Time, cursor *string) (*IssueConnection, error) {
	filter := principalEntityFilter()
	filter["statusChangedAt"] = map[string]interface{}{
		"after": since.Format(time.RFC3339),
	}

	variables := map[string]interface{}{
		"first":    100,
		"filterBy": filter,
	}
	if cursor != nil && *cursor != "" {
		variables["after"] = *cursor
	}

	var result issuesQueryResponse
	if err := c.graphQLRequest(ctx, issuesQuery, variables, &result); err != nil {
		return nil, fmt.Errorf("failed to list issues since %s: %w", since.Format(time.RFC3339), err)
	}

	return &result.IssuesV2, nil
}
