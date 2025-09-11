#!/bin/bash

# Script to filter OpenAPI spec by a specific tag
# Usage: ./filter-endpoints-by-tag.sh input.json output.json TAG_NAME

set -e

INPUT_FILE="$1"
OUTPUT_FILE="$2"
TAG_NAME="${3:-SDK}"

if [ -z "$INPUT_FILE" ] || [ -z "$OUTPUT_FILE" ]; then
    echo "Usage: $0 <input-openapi.json> <output-openapi.json> [tag_name]"
    echo "  tag_name defaults to 'SDK' if not provided"
    exit 1
fi

if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' not found"
    exit 1
fi

echo "Filtering OpenAPI spec to include only '$TAG_NAME'-tagged endpoints..."

# Create filtered OpenAPI spec with only specified tag endpoints
jq --indent 2 --arg tag "$TAG_NAME" '
  # Keep the base structure
  {
    "openapi": .openapi,
    "info": (.info // {}),
    "servers": (.servers // []),
    "components": (.components // {}),
    "security": (.security // []),
    # Filter paths to only include those with specified tag
    "paths": (
      .paths | 
      to_entries | 
      map(
        select(
          .value | 
          to_entries[] | 
          .value.tags[]? == $tag
        )
      ) | 
      from_entries
    ),
    # Keep tags that are related to the filtered endpoints
    "tags": (
      .tags // [] | 
      map(select(.name == $tag or .name == "OCR" or .name == "Upload"))
    )
  }
' "$INPUT_FILE" > "$OUTPUT_FILE"

# Validate that we have some endpoints
ENDPOINT_COUNT=$(jq '.paths | length' "$OUTPUT_FILE")

if [ "$ENDPOINT_COUNT" -eq 0 ]; then
    echo "Warning: No '$TAG_NAME'-tagged endpoints found in the OpenAPI spec!"
    echo "Available tags:"
    jq -r '.paths | to_entries[] | .value | to_entries[] | .value.tags[]?' "$INPUT_FILE" | sort -u | sed 's/^/  /'
    exit 1
fi

echo "✅ Filtered OpenAPI spec created with $ENDPOINT_COUNT '$TAG_NAME' endpoints:"
jq -r '.paths | keys[]' "$OUTPUT_FILE" | sed 's/^/  /'

echo "✅ Output written to: $OUTPUT_FILE"