#!/usr/bin/env bash

ROOT_MODULE="github.com/peertechde/argon"
ROOT_DIR=$(go list -f '{{.Dir}}' "${ROOT_MODULE}")

GIT_SHA=$(git rev-parse --short HEAD || echo "GitNotFound")
if [[ -n "$FAILPOINTS" ]]; then
  GIT_SHA="$GIT_SHA"-FAILPOINTS
fi

GO_BUILD_ENV=("CGO_ENABLED=0" "GO_BUILD_FLAGS=${GO_BUILD_FLAGS}" "GOOS=${GOOS}" "GOARCH=${GOARCH}")

COLOR_RED='\033[0;31m'
COLOR_ORANGE='\033[0;33m'
COLOR_GREEN='\033[0;32m'
COLOR_LIGHTCYAN='\033[0;36m'
COLOR_BLUE='\033[0;94m'
COLOR_MAGENTA='\033[95m'
COLOR_BOLD='\033[1m'
COLOR_NONE='\033[0m' # No Color

function log_error {
  >&2 echo -n -e "${COLOR_BOLD}${COLOR_RED}"
  >&2 echo "$@"
  >&2 echo -n -e "${COLOR_NONE}"
}

function log_warning {
  >&2 echo -n -e "${COLOR_ORANGE}"
  >&2 echo "$@"
  >&2 echo -n -e "${COLOR_NONE}"
}

function log_callout {
  >&2 echo -n -e "${COLOR_LIGHTCYAN}"
  >&2 echo "$@"
  >&2 echo -n -e "${COLOR_NONE}"
}

function log_cmd {
  >&2 echo -n -e "${COLOR_BLUE}"
  >&2 echo "$@"
  >&2 echo -n -e "${COLOR_NONE}"
}

function log_success {
  >&2 echo -n -e "${COLOR_GREEN}"
  >&2 echo "$@"
  >&2 echo -n -e "${COLOR_NONE}"
}

function log_info {
  >&2 echo -n -e "${COLOR_NONE}"
  >&2 echo "$@"
  >&2 echo -n -e "${COLOR_NONE}"
}

# From http://stackoverflow.com/a/12498485
function relativePath {
  # both $1 and $2 are absolute paths beginning with /
  # returns relative path to $2 from $1
  local source=$1
  local target=$2

  local commonPart=$source
  local result=""

  while [[ "${target#$commonPart}" == "${target}" ]]; do
    # no match, means that candidate common part is not correct
    # go up one level (reduce common part)
    commonPart="$(dirname "$commonPart")"
    # and record that we went back, with correct / handling
    if [[ -z $result ]]; then
      result=".."
    else
      result="../$result"
    fi
  done

  if [[ $commonPart == "/" ]]; then
    # special case for root (no common path)
    result="$result/"
  fi

  # since we now have identified the common part,
  # compute the non-common part
  local forwardPart="${target#$commonPart}"

  # and now stick all parts together
  if [[ -n $result ]] && [[ -n $forwardPart ]]; then
    result="$result$forwardPart"
  elif [[ -n $forwardPart ]]; then
    # extra slash removal
    result="${forwardPart:1}"
  fi

  echo "$result"
}

# Prints subdirectory (from the repo root) for the current module.
function module_subdir {
  relativePath "${ROOT_DIR}" "${PWD}"
}

# run [command...] - runs given command, printing it first and
# again if it failed (in RED). Use to wrap important test commands
# that user might want to re-execute to shorten the feedback loop when fixing
# the test.
function run {
  local rpath
  local command
  rpath=$(module_subdir)
  # Quoting all components as the commands are fully copy-parsable:
  command=("${@}")
  command=("${command[@]@Q}")
  if [[ "${rpath}" != "." && "${rpath}" != "" ]]; then
    repro="(cd ${rpath} && ${command[*]})"
  else 
    repro="${command[*]}"
  fi

  log_cmd "% ${repro}"
  "${@}" 2> >(while read -r line; do echo -e "${COLOR_NONE}stderr: ${COLOR_MAGENTA}${line}${COLOR_NONE}">&2; done)
  local error_code=$?
  if [ ${error_code} -ne 0 ]; then
    log_error -e "FAIL: (code:${error_code}):\\n  % ${repro}"
    return ${error_code}
  fi
}

build() {
  out="bin"

  run rm -f "${out}/argon"
  (
    cd ./cmd/argon
    # shellcheck disable=SC2086
    run env "${GO_BUILD_ENV[@]}" go build $GO_BUILD_FLAGS \
      -installsuffix=cgo \
      -o="../../${out}/argon" . || return 2
  ) || return 2
}

# only build when called directly, not sourced
if echo "$0" | grep -E "build(.sh)?$" >/dev/null; then
  if build ; then
    log_success "SUCCESS: build (GOARCH=${GOARCH})"
  else
    log_error "FAIL: build (GOARCH=${GOARCH})"
    exit 2
  fi
fi
