#!/bin/bash
# Build backup container image

set -euo pipefail

VERSION="${1:-latest}"
REGISTRY="${REGISTRY:-docker.io/trading-engine}"

echo "Building backup container image..."
echo "Version: $VERSION"
echo "Registry: $REGISTRY"

# Build image
docker build \
    -t "$REGISTRY/backup:$VERSION" \
    -t "$REGISTRY/backup:latest" \
    -f Dockerfile \
    ../..

echo "✓ Image built successfully"

# Push to registry
if [[ "${PUSH:-false}" == "true" ]]; then
    echo "Pushing to registry..."
    docker push "$REGISTRY/backup:$VERSION"
    docker push "$REGISTRY/backup:latest"
    echo "✓ Image pushed successfully"
fi

echo ""
echo "Run container:"
echo "  docker run -it --rm $REGISTRY/backup:$VERSION /scripts/backup-full.sh"
