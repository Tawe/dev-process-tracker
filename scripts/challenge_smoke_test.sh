#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ORIG_HOME="${HOME}"
TMP_HOME="${ROOT_DIR}/.tmp-home"

cleanup() {
  set +e
  if [[ -x "${ROOT_DIR}/devpt" ]]; then
    HOME="${TMP_HOME}" "${ROOT_DIR}/devpt" stop smoke-go >/dev/null 2>&1 || true
  fi
  rm -rf "${TMP_HOME}"
}
trap cleanup EXIT

rm -rf "${TMP_HOME}"
mkdir -p "${TMP_HOME}"

export GOPATH="${ORIG_HOME}/go"
export GOMODCACHE="${ORIG_HOME}/go/pkg/mod"
export GOCACHE="${ORIG_HOME}/Library/Caches/go-build"

echo "[1/9] Build"
HOME="${TMP_HOME}" go build -o "${ROOT_DIR}/devpt" ./cmd/devpt

echo "[2/9] Test"
HOME="${TMP_HOME}" go test ./...

echo "[3/9] Add service"
HOME="${TMP_HOME}" "${ROOT_DIR}/devpt" add smoke-go . "sleep 120" 3999

echo "[4/9] Start service"
HOME="${TMP_HOME}" "${ROOT_DIR}/devpt" start smoke-go
sleep 2

echo "[5/9] Status"
HOME="${TMP_HOME}" "${ROOT_DIR}/devpt" status smoke-go

echo "[6/9] Logs"
HOME="${TMP_HOME}" "${ROOT_DIR}/devpt" logs smoke-go --lines 20 || true

echo "[7/9] Restart service"
HOME="${TMP_HOME}" "${ROOT_DIR}/devpt" restart smoke-go
sleep 1

echo "[8/9] List services"
HOME="${TMP_HOME}" "${ROOT_DIR}/devpt" ls --details

echo "[9/9] Stop service"
HOME="${TMP_HOME}" "${ROOT_DIR}/devpt" stop smoke-go
HOME="${TMP_HOME}" "${ROOT_DIR}/devpt" ls --details

echo "Challenge smoke test completed."
