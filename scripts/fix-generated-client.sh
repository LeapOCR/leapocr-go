#!/bin/bash

# Fix generated client.go to only include existing API services
# This script removes references to API services that weren't generated

CLIENT_FILE="gen/client.go"

if [ ! -f "$CLIENT_FILE" ]; then
    echo "❌ Client file not found: $CLIENT_FILE"
    exit 1
fi

echo "🔧 Fixing generated client.go to remove undefined API services..."

# Check which API services actually exist by looking for their definitions
HAVE_SDKAPI=false
HAVE_OCRAPI=false
HAVE_UPLOADAPI=false

if grep -q "type SDKAPI\|SDKAPIService" gen/api_*.go 2>/dev/null; then
    HAVE_SDKAPI=true
    echo "📋 Found SDKAPI service"
fi
if grep -q "type OCRAPI\|OCRAPIService" gen/api_*.go 2>/dev/null; then
    HAVE_OCRAPI=true
    echo "📋 Found OCRAPI service"
fi
if grep -q "type UploadAPI\|UploadAPIService" gen/api_*.go 2>/dev/null; then
    HAVE_UPLOADAPI=true
    echo "📋 Found UploadAPI service"
fi

# Create backup
cp "$CLIENT_FILE" "$CLIENT_FILE.backup"

# Remove or comment out API services that don't exist
if [ "$HAVE_OCRAPI" = false ]; then
    echo "🔧 Removing OCRAPI references"
    # Comment out the OCRAPI field in struct
    sed -i '' 's/^[[:space:]]*OCRAPI OCRAPI[[:space:]]*$/\/\/ & - removed (not generated)/' "$CLIENT_FILE"
    # Comment out the OCRAPI initialization
    sed -i '' 's/^[[:space:]]*c\.OCRAPI.*OCRAPIService.*$/\/\/ & - removed (not generated)/' "$CLIENT_FILE"
fi

if [ "$HAVE_UPLOADAPI" = false ]; then
    echo "🔧 Removing UploadAPI references"
    # Comment out the UploadAPI field in struct
    sed -i '' 's/^[[:space:]]*UploadAPI UploadAPI[[:space:]]*$/\/\/ & - removed (not generated)/' "$CLIENT_FILE"
    # Comment out the UploadAPI initialization
    sed -i '' 's/^[[:space:]]*c\.UploadAPI.*UploadAPIService.*$/\/\/ & - removed (not generated)/' "$CLIENT_FILE"
fi

if [ "$HAVE_SDKAPI" = false ]; then
    echo "🔧 Removing SDKAPI references"
    # Comment out the SDKAPI field in struct
    sed -i '' 's/^[[:space:]]*SDKAPI SDKAPI[[:space:]]*$/\/\/ & - removed (not generated)/' "$CLIENT_FILE"
    # Comment out the SDKAPI initialization
    sed -i '' 's/^[[:space:]]*c\.SDKAPI.*SDKAPIService.*$/\/\/ & - removed (not generated)/' "$CLIENT_FILE"
fi

# Clean up backup if successful
if [ $? -eq 0 ]; then
    rm "$CLIENT_FILE.backup"
    echo "✅ Fixed generated client.go"
else
    echo "❌ Error occurred, restoring backup"
    mv "$CLIENT_FILE.backup" "$CLIENT_FILE"
    exit 1
fi

# Format the file
if go fmt "$CLIENT_FILE" 2>/dev/null; then
    echo "✅ Formatted client.go"
else
    echo "⚠️  Warning: Could not format client.go"
fi