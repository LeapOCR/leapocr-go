#!/bin/bash

# Script to filter OpenAPI spec to only include SDK-tagged endpoints and their dependencies
# Usage: ./filter-sdk-endpoints.sh input.json output.json

set -e

INPUT_FILE="$1"
OUTPUT_FILE="$2"

if [ -z "$INPUT_FILE" ] || [ -z "$OUTPUT_FILE" ]; then
    echo "Usage: $0 <input-openapi.json> <output-openapi.json>"
    exit 1
fi

if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' not found"
    exit 1
fi

echo "Filtering OpenAPI spec to include only SDK-tagged endpoints and their dependencies..."

# Create filtered OpenAPI spec with only SDK endpoints and referenced components
jq --indent 2 '
  # First, extract all SDK endpoints
  def sdk_paths:
    .paths | 
    to_entries | 
    map(
      select(
        .value | 
        to_entries[] | 
        .value.tags[]? == "SDK"
      )
    ) | 
    from_entries;
  
  # Function to extract all $ref references from a value recursively
  def extract_refs:
    .. | objects | to_entries[] | select(.key == "$ref") | .value;
  
  # Function to clean ref path (remove #/components/schemas/ prefix)
  def clean_ref:
    if startswith("#/components/schemas/") then
      .[21:]
    else
      .
    end;
  
  # Get all references used by SDK endpoints
  def used_refs:
    sdk_paths | [extract_refs | clean_ref] | unique;
  
  # Build the filtered spec
  {
    "openapi": .openapi,
    "info": .info,
    "servers": .servers,
    "security": .security,
    # Filter paths to only include those with SDK tag
    "paths": sdk_paths,
    # Keep only referenced components - simplified approach
    "components": {
      "schemas": (
        used_refs as $needed |
        .components.schemas | 
        with_entries(select(.key as $k | $needed | contains([$k])))
      ),
      "securitySchemes": .components.securitySchemes
    },
    # Keep only SDK tag
    "tags": [{"name": "SDK", "description": "SDK operations"}]
  }
' "$INPUT_FILE" > "$OUTPUT_FILE"

# Validate that we have some endpoints
ENDPOINT_COUNT=$(jq '.paths | length' "$OUTPUT_FILE")
SCHEMA_COUNT=$(jq '.components.schemas | length' "$OUTPUT_FILE")

if [ "$ENDPOINT_COUNT" -eq 0 ]; then
    echo "Warning: No SDK-tagged endpoints found in the OpenAPI spec!"
    echo "Available tags:"
    jq -r '.paths | to_entries[] | .value | to_entries[] | .value.tags[]?' "$INPUT_FILE" | sort -u | sed 's/^/  /'
    exit 1
fi

echo "✅ Filtered OpenAPI spec created with $ENDPOINT_COUNT SDK endpoints:"
jq -r '.paths | keys[]' "$OUTPUT_FILE" | sed 's/^/  /'

echo "✅ Included $SCHEMA_COUNT component schemas (only those referenced by SDK endpoints)"

echo "✅ Output written to: $OUTPUT_FILE"