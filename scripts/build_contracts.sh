#!/bin/bash
set -e  # Exit on error

# Determine the current script directory
CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Determine the parent directory
PARENT_DIR="$( dirname "$CURRENT_DIR" )"

# Move to the parent directory
cd "$PARENT_DIR"

# Install Go packages in tools/schema
echo "Installing Go packages in tools/schema..."
cd tools/schema
go install

# Move back to the original script directory
cd "$CURRENT_DIR"

echo "Script completed successfully."
