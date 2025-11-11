#!/bin/bash
# Helper script to setup Java PATH for OpenAPI generator
# Checks common Homebrew locations and exports PATH if Java is not already available

if ! command -v java >/dev/null 2>&1; then
    if [ -d /opt/homebrew/opt/openjdk@21/bin ]; then
        export PATH="/opt/homebrew/opt/openjdk@21/bin:$PATH"
    elif [ -d /opt/homebrew/opt/openjdk@17/bin ]; then
        export PATH="/opt/homebrew/opt/openjdk@17/bin:$PATH"
    elif [ -d /opt/homebrew/opt/openjdk/bin ]; then
        export PATH="/opt/homebrew/opt/openjdk/bin:$PATH"
    elif [ -d /usr/local/opt/openjdk@21/bin ]; then
        export PATH="/usr/local/opt/openjdk@21/bin:$PATH"
    elif [ -d /usr/local/opt/openjdk@17/bin ]; then
        export PATH="/usr/local/opt/openjdk@17/bin:$PATH"
    elif [ -d /usr/local/opt/openjdk/bin ]; then
        export PATH="/usr/local/opt/openjdk/bin:$PATH"
    fi
fi

# Verify Java is available
if ! command -v java >/dev/null 2>&1; then
    echo "❌ Java not found. Please install Java:" >&2
    echo "   brew install openjdk@21" >&2
    echo "   Or visit: https://www.java.com/download/" >&2
    exit 1
fi

# Export PATH so it's available to parent shell
export PATH

