#!/bin/sh
set -e

OPENFGA_URL="http://openfga:8080"

echo "=== OpenFGA Setup Script ==="
echo "Waiting for OpenFGA to be ready..."

# Wait for OpenFGA
until curl -s "$OPENFGA_URL/healthz" > /dev/null 2>&1; do
    echo "Waiting for OpenFGA..."
    sleep 2
done

echo "OpenFGA is ready!"

# Create store
echo "Creating store..."
STORE_RESPONSE=$(curl -s -X POST "$OPENFGA_URL/stores" \
    -H "Content-Type: application/json" \
    -d '{"name": "examples-store"}')

STORE_ID=$(echo "$STORE_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$STORE_ID" ]; then
    echo "Failed to create store. Response: $STORE_RESPONSE"
    exit 1
fi

echo "Store created: $STORE_ID"

# Save store ID to shared volume
echo "$STORE_ID" > /shared/openfga-store-id

# Create authorization model
echo "Creating authorization model..."
MODEL=$(cat /deploy/model.json)

MODEL_RESPONSE=$(curl -s -X POST "$OPENFGA_URL/stores/$STORE_ID/authorization-models" \
    -H "Content-Type: application/json" \
    -d "$MODEL")

MODEL_ID=$(echo "$MODEL_RESPONSE" | grep -o '"authorization_model_id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$MODEL_ID" ]; then
    echo "Failed to create model. Response: $MODEL_RESPONSE"
    exit 1
fi

echo "Authorization model created: $MODEL_ID"

# Write sample tuples
echo "Writing sample tuples..."

# Sample workspace
curl -s -X POST "$OPENFGA_URL/stores/$STORE_ID/write" \
    -H "Content-Type: application/json" \
    -d '{
        "writes": {
            "tuple_keys": [
                {"user": "user:user-1", "relation": "admin", "object": "workspace:workspace-1"},
                {"user": "user:user-2", "relation": "member", "object": "workspace:workspace-1"},
                {"user": "user:user-3", "relation": "viewer", "object": "workspace:workspace-1"}
            ]
        }
    }' > /dev/null

echo "Created workspace tuples"

# Sample documents with different access levels
curl -s -X POST "$OPENFGA_URL/stores/$STORE_ID/write" \
    -H "Content-Type: application/json" \
    -d '{
        "writes": {
            "tuple_keys": [
                {"user": "workspace:workspace-1", "relation": "workspace", "object": "document:doc-1"},
                {"user": "user:user-1", "relation": "owner", "object": "document:doc-1"},

                {"user": "workspace:workspace-1", "relation": "workspace", "object": "document:doc-2"},
                {"user": "user:user-2", "relation": "owner", "object": "document:doc-2"},
                {"user": "user:user-1", "relation": "editor", "object": "document:doc-2"},

                {"user": "workspace:workspace-1", "relation": "workspace", "object": "document:doc-3"},
                {"user": "user:user-3", "relation": "owner", "object": "document:doc-3"},
                {"user": "user:user-1", "relation": "viewer", "object": "document:doc-3"},
                {"user": "user:user-2", "relation": "viewer", "object": "document:doc-3"}
            ]
        }
    }' > /dev/null

echo "Created document tuples"

# Sample projects
curl -s -X POST "$OPENFGA_URL/stores/$STORE_ID/write" \
    -H "Content-Type: application/json" \
    -d '{
        "writes": {
            "tuple_keys": [
                {"user": "workspace:workspace-1", "relation": "workspace", "object": "project:proj-1"},
                {"user": "user:user-1", "relation": "owner", "object": "project:proj-1"},

                {"user": "workspace:workspace-1", "relation": "workspace", "object": "project:proj-2"},
                {"user": "user:user-2", "relation": "owner", "object": "project:proj-2"},
                {"user": "user:user-1", "relation": "admin", "object": "project:proj-2"}
            ]
        }
    }' > /dev/null

echo "Created project tuples"

echo ""
echo "=== Setup Complete ==="
echo "Store ID: $STORE_ID"
echo "Model ID: $MODEL_ID"
echo ""
echo "Sample data created:"
echo "  - Workspace: workspace-1"
echo "    - user-1: admin"
echo "    - user-2: member"
echo "    - user-3: viewer"
echo ""
echo "  - Documents:"
echo "    - doc-1: owned by user-1"
echo "    - doc-2: owned by user-2, user-1 is editor"
echo "    - doc-3: owned by user-3, user-1 and user-2 are viewers"
echo ""
echo "  - Projects:"
echo "    - proj-1: owned by user-1"
echo "    - proj-2: owned by user-2, user-1 is admin"
echo ""
echo "OpenFGA Playground: http://localhost:3001"
echo "OpenFGA API: http://localhost:8081"
