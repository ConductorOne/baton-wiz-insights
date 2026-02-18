package connector

import (
	"context"
	"fmt"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	cfg "github.com/conductorone/baton-wiz-insights/pkg/config"
	"github.com/conductorone/baton-wiz-insights/pkg/wiz"
)

type Connector struct {
	client wiz.Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (c *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
		newIssueBuilder(c.client),
	}
}

// EventFeeds returns the event feeds supported by this connector.
// This makes the Connector satisfy EventProviderV2 and the SDK
// will automatically report CAPABILITY_EVENT_FEED_V2.
func (c *Connector) EventFeeds(_ context.Context) []connectorbuilder.EventFeed {
	return []connectorbuilder.EventFeed{
		newIssuesEventFeed(c),
	}
}

// Close releases any resources held by the connector's client.
func (c *Connector) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (c *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, fmt.Errorf("baton-wiz-insights: asset retrieval not supported")
}

// Metadata returns metadata about the connector.
func (c *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Wiz Insights",
		Description: "Wiz cloud security platform connector for syncing security issues as insights",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (c *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	if err := c.client.ValidateCredentials(ctx); err != nil {
		return nil, fmt.Errorf("baton-wiz-insights: failed to validate Wiz API credentials: %w", err)
	}
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context,
	connectorConfig *cfg.WizInsights,
	cliOpts *cli.ConnectorOpts,
) (connectorbuilder.ConnectorBuilderV2,
	[]connectorbuilder.Opt,
	error,
) {
	// Initialize the Wiz API client
	client, err := wiz.NewClient(
		ctx,
		connectorConfig.WizApiUrl,
		connectorConfig.WizClientId,
		connectorConfig.WizClientSecret,
		connectorConfig.WizAuthEndpoint,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Wiz client: %w", err)
	}

	return &Connector{client: client}, nil, nil
}
