package wiz

import "time"

// PageInfo represents GraphQL pagination information using Relay cursor pagination.
type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

// SourceRule represents the Wiz rule that triggered the issue.
type SourceRule struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EntitySnapshot represents the cloud resource entity associated with an issue.
type EntitySnapshot struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Name           string `json:"name"`
	NativeType     string `json:"nativeType"`
	ExternalID     string `json:"externalId"`
	CloudPlatform  string `json:"cloudPlatform"`
	SubscriptionID string `json:"subscriptionId"`
}

// Issue represents a Wiz security issue.
type Issue struct {
	ID              string         `json:"id"`
	Status          string         `json:"status"`
	Severity        string         `json:"severity"`
	CreatedAt       time.Time      `json:"createdAt"`
	StatusChangedAt time.Time      `json:"statusChangedAt"`
	SourceRule      SourceRule     `json:"sourceRule"`
	EntitySnapshot  EntitySnapshot `json:"entitySnapshot"`
}

// IssueConnection represents a paginated list of issues.
type IssueConnection struct {
	Nodes    []Issue  `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

// GraphQL response wrapper types.
type graphQLResponse struct {
	Data   interface{}    `json:"data"`
	Errors []graphQLError `json:"errors,omitempty"`
}

type graphQLError struct {
	Message string        `json:"message"`
	Path    []interface{} `json:"path,omitempty"`
}

// Specific response types for each query.
type issuesQueryResponse struct {
	IssuesV2 IssueConnection `json:"issuesV2"`
}
