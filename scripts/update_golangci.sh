#!/bin/bash

# URL of the Gist containing the .golangci.yml file
GIST_URL="https://gist.githubusercontent.com/maratori/47a4d00457a92aa426dbd48a18776322/raw"

# Download the content of the Gist
curl -o .golangci.yml "$GIST_URL"

echo "Downloaded .golangci.yml successfully."

