#!/usr/bin/env bash

set -euo pipefail

# Defines the git repository the cache entry belongs to.
# Example: openclarity/vmclarity
declare -rx REPO="${REPO:?}"
# Defines the git reference the cache entry belongs to.
declare -rx REF="${REF:-}"
# Defines the operation mode for the script where `ref` cleans up cache keys based belongs to a specific git reference
# while `scheduled` performs cleanup based on the time the cache was last used.
# Values: by-ref|by-age
# Default: ref
declare -rx OPERATION_MODE="${OPERATION_MODE:-by-ref}"
# If set to true the script runs in dry-run mode.
# Values: true|false
# Default: true
declare -rx DRY_RUN="${DRY_RUN:-true}"
# Cleanup cache entries with last used time older than this value.
# Default: 7 days ago
# Example: 1 hour ago
declare -rx OLDER_THAN="${OLDER_THAN:-7 days ago}"

cleanup_by_ref () {
  while IFS=$'\t' read -r cache_key size ref last_used; do
    if [[ "${DRY_RUN}" == "true" ]]; then
      printf "INFO: skip deleting cache due to dry-run: key: %s | size: %s | ref: %s | last used: %s\n" \
          "${cache_key}" "${size}" "${ref}" "${last_used}"
      continue
    fi

    printf "INFO: deleting cache: key: %s | size: %s | ref: %s | last used: %s\n" \
        "${cache_key}" "${size}" "${ref}" "${last_used}"
    gh actions-cache delete --repo "${REPO}" --branch "${REF}" --confirm "${cache_key}" \
    || printf "ERROR: failed to delete cache-key: %s\n" "${cache_key}"
 done < <(gh actions-cache list --repo "${REPO}" --branch "${REF}" --limit 100)
}

cleanup_by_age () {
  while IFS=$'\t' read -r cache_key size ref last_used; do
      local used older_than

      # GNU date command returns parse error if the last_used variable has the "an hour ago"
      # or "a few seconds ago" string value.
      used="${last_used//an hour/a hour}"
      used="${used//a few/2}"
      used="$(date -d "${used}" +'%s')"
      older_than="$(date -d "${OLDER_THAN}" +'%s')"

      if (( "${used}" > "${older_than}" )); then
        printf "INFO: skip deleting cache due to it's age: key: %s | size: %s | ref: %s | last used: %s\n" \
            "${cache_key}" "${size}" "${ref}" "${last_used}"
        continue
      fi

      if [[ "${DRY_RUN}" == "true" ]]; then
        printf "INFO: skip deleting cache due to dry-run: key: %s | size: %s | ref: %s | last used: %s\n" \
            "${cache_key}" "${size}" "${ref}" "${last_used}"
        continue
      fi

      printf "INFO: deleting cache: key: %s | size: %s | ref: %s | last used: %s\n" \
          "${cache_key}" "${size}" "${ref}" "${last_used}"
      gh actions-cache delete --repo "${REPO}" --branch "${ref}" --confirm "${cache_key}" \
      || printf "ERROR: failed to delete cache-key: %s\n" "${cache_key}"
  done < <(gh actions-cache list --repo "${REPO}" --limit 100 --sort last-used --order asc)
}

main () {
  # Install extension if it is not available
  if ! gh actions-cache list --repo "${REPO}" --branch "${REF}" --limit 1 &> /dev/null; then
    gh extension install actions/gh-actions-cache
  fi

  case "${OPERATION_MODE}" in
  by-ref)
    cleanup_by_ref
    ;;
  by-age)
    cleanup_by_age
    ;;
  *)
    printf "ERROR: %s\n" "invalid OPERATION_MODE"
    exit 1
    ;;
  esac
}

main
