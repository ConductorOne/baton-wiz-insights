package connector

import (
	"context"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const eventFeedID = "wiz_issues_feed"

// issuesEventFeed implements connectorbuilder.EventFeed by polling
// the issuesV2 GraphQL query filtered by statusChangedAt to only
// return issues modified since the last check.
type issuesEventFeed struct {
	connector *Connector
}

func newIssuesEventFeed(connector *Connector) *issuesEventFeed {
	return &issuesEventFeed{connector: connector}
}

func (e *issuesEventFeed) EventFeedMetadata(_ context.Context) *v2.EventFeedMetadata {
	return v2.EventFeedMetadata_builder{
		Id: eventFeedID,
		SupportedEventTypes: []v2.EventType{
			v2.EventType_EVENT_TYPE_RESOURCE_CHANGE,
		},
	}.Build()
}

func (e *issuesEventFeed) ListEvents(
	ctx context.Context,
	earliestEvent *timestamppb.Timestamp,
	pToken *pagination.StreamToken,
) ([]*v2.Event, *pagination.StreamState, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	// Decode cursor from the stream token. On first call, cursor is empty
	// and we use earliestEvent as the start time.
	cursor, err := decodeEventCursor(pToken, earliestEvent)
	if err != nil {
		return nil, nil, nil, err
	}

	l.Debug("wiz-event-feed: querying issues",
		zap.String("since", cursor.Since.Format(time.RFC3339)),
		zap.String("page_cursor", cursor.PageEndCursor))

	// Query issuesV2 filtered by statusChangedAt >= cursor.Since
	var pageCursor *string
	if cursor.PageEndCursor != "" {
		pageCursor = &cursor.PageEndCursor
	}

	issuesResp, err := e.connector.client.ListIssuesSince(ctx, cursor.Since, pageCursor)
	if err != nil {
		return nil, nil, nil, err
	}

	// Convert each issue to a RESOURCE_CHANGE event
	var events []*v2.Event
	for _, issue := range issuesResp.Nodes {
		event := v2.Event_builder{
			Id:         issue.ID,
			OccurredAt: timestamppb.New(issue.StatusChangedAt),
			ResourceChangeEvent: v2.ResourceChangeEvent_builder{
				ResourceId: v2.ResourceId_builder{
					ResourceType: issueResourceType.GetId(),
					Resource:     issue.ID,
				}.Build(),
			}.Build(),
		}.Build()
		events = append(events, event)

		// Track the latest statusChangedAt we've seen
		if issue.StatusChangedAt.After(cursor.LatestSeen) {
			cursor.LatestSeen = issue.StatusChangedAt
		}
	}

	// Build next cursor
	hasMore := issuesResp.PageInfo.HasNextPage
	if hasMore {
		cursor.PageEndCursor = issuesResp.PageInfo.EndCursor
	} else {
		// Done with this sweep. Next call starts from the latest timestamp we saw.
		cursor.Since = cursor.LatestSeen
		cursor.PageEndCursor = ""
	}

	nextCursor, err := cursor.encode()
	if err != nil {
		return nil, nil, nil, err
	}

	streamState := &pagination.StreamState{
		Cursor:  nextCursor,
		HasMore: hasMore,
	}

	l.Debug("wiz-event-feed: processed issues",
		zap.Int("count", len(events)),
		zap.Bool("has_more", hasMore))

	return events, streamState, nil, nil
}
