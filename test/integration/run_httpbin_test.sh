#!/usr/bin/env bash
# Copyright 2026-present matstech
# SPDX-License-Identifier: GPL-3.0-only


set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
WORK_DIR="$(mktemp -d)"
RUN_ID="${RUN_ID:-$(date +%s)-$$}"

HTTPBIN_IMAGE="${HTTPBIN_IMAGE:-ghcr.io/mccutchen/go-httpbin}"
HTTPBIN_CONTAINER_NAME="${HTTPBIN_CONTAINER_NAME:-aegis-httpbin-itest-${RUN_ID}}"
AEGIS_IMAGE="${AEGIS_IMAGE:-aegis-integration:${RUN_ID}}"
AEGIS_CONTAINER_NAME="${AEGIS_CONTAINER_NAME:-aegis-itest-${RUN_ID}}"
DOCKER_NETWORK_NAME="${DOCKER_NETWORK_NAME:-aegis-itest-net-${RUN_ID}}"
HTTPBIN_PORT="${HTTPBIN_PORT:-$((30000 + RANDOM % 2000))}"
AEGIS_PORT="${AEGIS_PORT:-$((33000 + RANDOM % 2000))}"
AEGIS_PROBES_PORT="${AEGIS_PROBES_PORT:-$((36000 + RANDOM % 2000))}"

ACCESSKEY_TEST="${ACCESSKEY_TEST:-integration-secret}"
UPSTREAM_AUTHORIZATION="${UPSTREAM_AUTHORIZATION:-Bearer upstream-secret}"
AUTH_KID="test"
AUTH_HEADERS="Content-Type;X-Drop-Me;Authorization"
CORRELATION_ID="itest-$(date +%s)"
RESPONSE_FILE="${WORK_DIR}/response.json"
CONFIG_DIR="${WORK_DIR}/config"

mkdir -p "${CONFIG_DIR}"

cleanup() {
  status=$?

  if [[ ${status} -ne 0 ]]; then
    echo "aegis log:" >&2
    docker logs "${AEGIS_CONTAINER_NAME}" >&2 || true
  fi

  if [[ ${status} -ne 0 && -f "${RESPONSE_FILE}" ]]; then
    echo "httpbin response:" >&2
    cat "${RESPONSE_FILE}" >&2
  fi

  docker rm -f "${AEGIS_CONTAINER_NAME}" >/dev/null 2>&1 || true
  docker rm -f "${HTTPBIN_CONTAINER_NAME}" >/dev/null 2>&1 || true
  docker network rm "${DOCKER_NETWORK_NAME}" >/dev/null 2>&1 || true
  rm -rf "${WORK_DIR}"
  exit ${status}
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

wait_for_http() {
  local url="$1"
  local attempts="${2:-30}"

  for ((i = 1; i <= attempts; i++)); do
    if curl --silent --show-error --fail "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  echo "service did not become ready: ${url}" >&2
  return 1
}

assert_response() {
  jq -e --arg expected_auth "${UPSTREAM_AUTHORIZATION}" '
    .headers.Authorization == [$expected_auth] and
    .headers["X-Aegis-Proxy"] == ["true"] and
    (.headers["X-Drop-Me"] == null) and
    (.headers["Auth-Kid"] == null) and
    (.headers["Auth-Headers"] == null) and
    (.headers["Signature"] == null) and
    (.headers["Auth-Correlationid"] != null) and
    .json.message == "integration-test"
  ' "${RESPONSE_FILE}" >/dev/null
}

trap cleanup EXIT

require_command docker
require_command curl
require_command jq
require_command go

cat > "${CONFIG_DIR}/config.json" <<EOF
{
  "ginmode": "release",
  "loglevel": "debug",
  "server": {
    "mode": "PLAIN",
    "port": 8080,
    "probesport": 2112,
    "upstream": "${HTTPBIN_CONTAINER_NAME}:8080",
    "dropHeaders": ["Authorization", "X-Drop-Me"],
    "injectHeaders": [
      {
        "name": "X-Aegis-Proxy",
        "value": "true"
      },
      {
        "name": "Authorization",
        "valueFromEnv": "UPSTREAM_AUTHORIZATION"
      }
    ]
  },
  "kids": ["${AUTH_KID}"]
}
EOF

docker build --tag "${AEGIS_IMAGE}" "${ROOT_DIR}" >/dev/null

docker network create "${DOCKER_NETWORK_NAME}" >/dev/null

docker rm -f "${HTTPBIN_CONTAINER_NAME}" >/dev/null 2>&1 || true
docker run --detach --rm \
  --name "${HTTPBIN_CONTAINER_NAME}" \
  --network "${DOCKER_NETWORK_NAME}" \
  --publish "${HTTPBIN_PORT}:8080" \
  "${HTTPBIN_IMAGE}" >/dev/null

wait_for_http "http://127.0.0.1:${HTTPBIN_PORT}/status/200"

docker rm -f "${AEGIS_CONTAINER_NAME}" >/dev/null 2>&1 || true
docker run --detach --rm \
  --name "${AEGIS_CONTAINER_NAME}" \
  --network "${DOCKER_NETWORK_NAME}" \
  --publish "${AEGIS_PORT}:8080" \
  --publish "${AEGIS_PROBES_PORT}:2112" \
  --volume "${CONFIG_DIR}:/etc/config:ro" \
  --env CONFIG_PATH=/etc/config/ \
  --env ACCESSKEY_TEST="${ACCESSKEY_TEST}" \
  --env UPSTREAM_AUTHORIZATION="${UPSTREAM_AUTHORIZATION}" \
  "${AEGIS_IMAGE}" >/dev/null

wait_for_http "http://127.0.0.1:${AEGIS_PROBES_PORT}/readiness"

SIGNATURE="$(cd "${ROOT_DIR}" && go run ./test/integration/sign_request.go \
  --kid "${AUTH_KID}" \
  --secret "${ACCESSKEY_TEST}" \
  --correlation-id "${CORRELATION_ID}" \
  --auth-headers "${AUTH_HEADERS}" \
  --body-file "${ROOT_DIR}/test/integration/request.json" \
  --header "Content-Type: application/json" \
  --header "X-Drop-Me: drop-me" \
  --header "Authorization: Bearer client-token")"

curl --silent --show-error --fail \
  --request POST "http://127.0.0.1:${AEGIS_PORT}/anything" \
  --header "Auth-CorrelationId: ${CORRELATION_ID}" \
  --header "Auth-Kid: ${AUTH_KID}" \
  --header "Auth-Headers: ${AUTH_HEADERS}" \
  --header "Signature: ${SIGNATURE}" \
  --header "Content-Type: application/json" \
  --header "X-Drop-Me: drop-me" \
  --header "Authorization: Bearer client-token" \
  --data-binary "@${ROOT_DIR}/test/integration/request.json" \
  > "${RESPONSE_FILE}"

assert_response

echo "integration test passed"
