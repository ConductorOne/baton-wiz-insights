package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-wiz-insights/pkg/wiz"
)

type issueBuilder struct {
	client wiz.Client
}

func (i *issueBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return issueResourceType
}

// List returns Wiz issues as security insight resources, one page at a time.
func (i *issueBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, attr resource.SyncOpAttrs) ([]*v2.Resource, *resource.SyncOpResults, error) {
	var resources []*v2.Resource

	// Get the page token from the sync attributes
	var cursor *string
	if attr.PageToken.Token != "" {
		cursor = &attr.PageToken.Token
	}

	// Fetch one page of issues
	resp, err := i.client.ListIssues(ctx, cursor)
	if err != nil {
		return nil, nil, fmt.Errorf("wiz-connector: failed to list issues: %w", err)
	}

	for _, issue := range resp.Nodes {
		insightOpts := []resource.SecurityInsightTraitOption{
			resource.WithIssue(issue.SourceRule.Name),
			resource.WithIssueSeverity(issue.Severity),
			resource.WithInsightObservedAt(issue.StatusChangedAt),
		}

		// Set the target based on the entity snapshot.
		// Use external resource target since the entity is an external cloud resource.
		if issue.EntitySnapshot.ExternalID != "" {
			insightOpts = append(insightOpts,
				resource.WithInsightExternalResourceTarget(
					issue.EntitySnapshot.ExternalID,
					issue.EntitySnapshot.CloudPlatform,
				),
			)
		} else {
			// Fallback: use the entity snapshot ID as external resource target
			insightOpts = append(insightOpts,
				resource.WithInsightExternalResourceTarget(
					issue.EntitySnapshot.ID,
					issue.EntitySnapshot.CloudPlatform,
				),
			)
		}

		displayName := fmt.Sprintf("[%s] %s", issue.Severity, issue.SourceRule.Name)

		insightResource, err := resource.NewSecurityInsightResource(
			displayName,
			issueResourceType,
			issue.ID,
			insightOpts...,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("wiz-connector: failed to create security insight resource for issue %s: %w", issue.ID, err)
		}

		resources = append(resources, insightResource)
	}

	// Prepare the sync results with next page token if there are more pages
	syncResults := &resource.SyncOpResults{}
	if resp.PageInfo.HasNextPage {
		syncResults.NextPageToken = resp.PageInfo.EndCursor
	}

	return resources, syncResults, nil
}

// Entitlements returns an empty slice for issues (security insights don't have entitlements).
func (i *issueBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ resource.SyncOpAttrs) ([]*v2.Entitlement, *resource.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants returns an empty slice for issues (security insights don't have grants).
func (i *issueBuilder) Grants(_ context.Context, _ *v2.Resource, _ resource.SyncOpAttrs) ([]*v2.Grant, *resource.SyncOpResults, error) {
	return nil, nil, nil
}

func newIssueBuilder(client wiz.Client) *issueBuilder {
	return &issueBuilder{client: client}
}
