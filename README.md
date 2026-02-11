![Baton Logo](./baton-logo.png)

# `baton-wiz-insights` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-wiz-insights.svg)](https://pkg.go.dev/github.com/conductorone/baton-wiz-insights) ![ci](https://github.com/conductorone/baton-wiz-insights/actions/workflows/verify.yaml/badge.svg)

`baton-wiz-insights` is a connector for the Wiz.io cloud security platform built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It syncs security insights (issues) from Wiz that are related to user and service accounts, enabling identity-aware cloud security posture visibility.

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Prerequisites

- **Wiz Account**: You need an active Wiz account with API access
- **OAuth2 Credentials**: Create an OAuth2 client in Wiz with the following permissions:
  - `read:issues` - To sync security issues/insights
- **API Endpoints**: You'll need both the GraphQL API URL and the OAuth2 token endpoint for your Wiz region

# Getting Started

## brew

```bash
brew install conductorone/baton/baton conductorone/baton/baton-wiz-insights

baton-wiz-insights \
  --wiz-api-url "https://api.wiz.io/graphql" \
  --wiz-client-id "your-client-id" \
  --wiz-client-secret "your-client-secret" \
  --wiz-auth-endpoint "https://auth.wiz.io/oauth/token"

baton resources
```

## docker

```bash
docker run --rm -v $(pwd):/out \
  -e BATON_WIZ_API_URL="https://api.wiz.io/graphql" \
  -e BATON_WIZ_CLIENT_ID="your-client-id" \
  -e BATON_WIZ_CLIENT_SECRET="your-client-secret" \
  -e BATON_WIZ_AUTH_ENDPOINT="https://auth.wiz.io/oauth/token" \
  ghcr.io/conductorone/baton-wiz-insights:latest -f "/out/sync.c1z"

docker run --rm -v $(pwd):/out \
  ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```bash
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-wiz-insights/cmd/baton-wiz-insights@main

baton-wiz-insights \
  --wiz-api-url "https://api.wiz.io/graphql" \
  --wiz-client-id "your-client-id" \
  --wiz-client-secret "your-client-secret" \
  --wiz-auth-endpoint "https://auth.wiz.io/oauth/token"

baton resources
```

# Data Model

`baton-wiz-insights` synchronizes security insights from Wiz, filtered to issues related to identity resources:

- **Security Insights**: Wiz issues related to `USER_ACCOUNT` and `SERVICE_ACCOUNT` entity types, including issue severity, status, source rule, and the affected entity

The connector supports incremental sync via an event feed that polls for issues with updated statuses.

`baton-wiz-insights` does not support account provisioning or entitlement provisioning.

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-wiz-insights` Command Line Usage

```
baton-wiz-insights

Usage:
  baton-wiz-insights [flags]
  baton-wiz-insights [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-wiz-insights
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-wiz-insights
      --wiz-api-url string           required: The Wiz GraphQL API endpoint for your region ($BATON_WIZ_API_URL)
      --wiz-auth-endpoint string     required: OAuth2 token endpoint for authentication ($BATON_WIZ_AUTH_ENDPOINT)
      --wiz-client-id string         required: OAuth2 client ID from your Wiz service account ($BATON_WIZ_CLIENT_ID)
      --wiz-client-secret string     required: OAuth2 client secret from your Wiz service account ($BATON_WIZ_CLIENT_SECRET)

Use "baton-wiz-insights [command] --help" for more information about a command.
```
