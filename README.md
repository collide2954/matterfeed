# matterfeed

`matterfeed` is a Go-based RSS/ATOM/XML feed reader that scans for new articles and sends them to Mattermost (or Slack) channels using webhooks.

## Features

- Reads multiple RSS/ATOM/XML feeds
- Sends new articles to Mattermost using a webhook
- Utilizes SQLite for information storage
- Opens a health-check endpoint for status changes

## Development

Refer to the Makefile for available commands:

- `make brew`- Ensure the software dependencies for development
- `make build`- Build the binary
- `make run`  - Run the application
- `make clean`- Remove the binary
- `make lint` - Run linters with --fix flag for automatic fixes
- `make test` - Run tests
- `make vuln` - Check for vulnerabilities in dependencies
