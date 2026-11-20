#!/usr/bin/env bash
set -euo pipefail

case ${1:-} in
  ""|--push) ;;
  *) echo "usage: $0 [--push]" >&2; exit 2 ;;
esac

: "${AWS_ECR_REGISTRY:?set the ECR registry hostname}"
: "${RELEASE_TAG:?set the release tag}"
version=${RELEASE_TAG#v}
if [[ ! $version =~ ^[0-9][A-Za-z0-9_.-]*$ ]]; then
  echo "invalid release tag: $RELEASE_TAG" >&2
  exit 2
fi

image="$AWS_ECR_REGISTRY/gotenberg-fulll"
versioned_tag="$image:$version-cloudrun"
latest_tag="$image:latest"

docker buildx build --platform linux/amd64 --target gotenberg-cloudrun --load \
  --build-arg "GOTENBERG_VERSION=$version" --build-arg "APT_SNAPSHOT=$(date -u +%G-W%V)" \
  -f build/Dockerfile -t "$versioned_tag" -t "$latest_tag" .

printf 'Cloud Run image tags: %s %s\n' "$versioned_tag" "$latest_tag"
if [[ ${1:-} == --push ]]; then
  docker push "$versioned_tag"
  docker push "$latest_tag"
fi
