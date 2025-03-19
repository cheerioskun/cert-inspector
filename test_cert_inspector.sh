#!/bin/bash

# Build cert-inspector if needed
# Uncomment the following line if you want to build from source
# go build -o cert-inspector ./...

echo "Running cert-inspector search on testdata directory..."
./cert-inspector search ./testdata

echo -e "\nRunning cert-inspector tree to visualize the certificates..."
./cert-inspector tree

echo -e "\nTest completed. Check the output above for the certificate information."
