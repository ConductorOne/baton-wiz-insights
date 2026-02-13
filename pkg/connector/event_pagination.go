package connector

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/conductorone/baton-sdk/pkg/pagination"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// eventCursor tracks where we are in the issues polling loop.
type eventCursor struct {
	// Since is the statusChangedAt lower bound for the current query window.
	Since time.Time `json:"since"`

	// PageEndCursor is the GraphQL pagination cursor within the current window.
	PageEndCursor string `json:"page_end_cursor,omitempty"`

	// LatestSeen is the most recent statusChangedAt we encountered.
	// When we finish a sweep (no more pages), this becomes the next Since.
	LatestSeen time.Time `json:"latest_seen"`
}

func decodeEventCursor(token *pagination.StreamToken, defaultStart *timestamppb.Timestamp) (*eventCursor, error) {
	cursor := &eventCursor{}

	if token != nil && token.Cursor != "" {
		data, err := base64.StdEncoding.DecodeString(token.Cursor)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, cursor); err != nil {
			return nil, err
		}
		// Guard against zero LatestSeen from old or malformed tokens.
		if cursor.LatestSeen.IsZero() {
			cursor.LatestSeen = cursor.Since
		}
		return cursor, nil
	}

	// First call -- use the provided earliest event time, or default to 30 days ago.
	if defaultStart != nil {
		cursor.Since = defaultStart.AsTime()
	} else {
		cursor.Since = time.Now().Add(-30 * 24 * time.Hour)
	}
	cursor.LatestSeen = cursor.Since

	return cursor, nil
}

func (c *eventCursor) encode() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
