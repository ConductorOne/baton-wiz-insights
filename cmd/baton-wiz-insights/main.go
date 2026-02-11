//go:build !generate

package main

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
	cfg "github.com/conductorone/baton-wiz-insights/pkg/config"
	"github.com/conductorone/baton-wiz-insights/pkg/connector"
)

var version = "dev"

func main() {
	ctx := context.Background()

	config.RunConnector(
		ctx,
		"baton-wiz-insights",
		version,
		cfg.Config,
		connector.New,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilder(&connector.Connector{}),
		// connectorrunner.WithSessionStoreEnabled(), if the connector needs a cache.
	)
}
