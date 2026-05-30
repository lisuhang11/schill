#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
TAG="${1:-latest}"

build_image() {
  image="$1"
  dockerfile="$2"

  echo "Building ${image}:${TAG}"
  docker build -f "$REPO_ROOT/$dockerfile" -t "${image}:${TAG}" "$REPO_ROOT"
}

build_image schill-user-rpc service/user/rpc/Dockerfile
build_image schill-content-rpc service/content/rpc/Dockerfile
build_image schill-feed-rpc service/feed/rpc/Dockerfile
build_image schill-comment-rpc service/comment/rpc/Dockerfile
build_image schill-interaction-rpc service/interaction/rpc/Dockerfile
build_image schill-relation-rpc service/relation/rpc/Dockerfile
build_image schill-search-api service/search/api/Dockerfile
build_image schill-canal service/canal/Dockerfile
build_image schill-gateway service/gateway/Dockerfile

echo "Build complete."
echo "Next: cd \"$SCRIPT_DIR\" && docker compose -f docker-compose.prod.yml up -d"
