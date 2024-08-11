#!/bin/bash

## TODO(ramizpolic): Make script easier to work for Makefile

set -euox pipefail

## Required vars
OPENCLARITY_BASE_BRANCH=${OPENCLARITY_BASE_BRANCH:-unification/base}
VMCLARITY_BRANCH_TO_PORT=${VMCLARITY_BRANCH_TO_PORT}

## Start backporting
OPENCLARITY_BACKPORTED_BRANCH_DEST="backport-vmclarity/${VMCLARITY_BRANCH_TO_PORT}"
echo "Backporting branch ${VMCLARITY_BRANCH_TO_PORT} from VMClarity..."

# Switch to base branch
git checkout ${OPENCLARITY_BASE_BRANCH}

# Import VMClarity remote
git remote add --fetch vmclarity-repo https://github.com/openclarity/vmclarity

# Checkout backport feature branch
git checkout -b ${OPENCLARITY_BACKPORTED_BRANCH_DEST} ${OPENCLARITY_BASE_BRANCH}

# Merge remote to current branch
git merge --no-commit --allow-unrelated-histories vmclarity-repo/${VMCLARITY_BRANCH_TO_PORT}
git commit -m "chore: merge remote branch openclarity/vmclarity/${VMCLARITY_BRANCH_TO_PORT}"
git push --set-upstream origin ${OPENCLARITY_BACKPORTED_BRANCH_DEST}

# Overwrite history to include vmclarity references
python3 ../git-filter-repo --force --refs "${OPENCLARITY_BACKPORTED_BRANCH_DEST}" --message-callback '
  if message.startswith(b"kubeclarity: "):
      return message
  if message.startswith(b"vmclarity: "):
      return message

  message = re.sub(b"(#\d+)", b"vmclarity\\1", message, flags=re.MULTILINE)
  return b"vmclarity: "+message
'

## Do a force push after update
echo "Once you have checked the changes, force push them"

## After force push, rebase against main :)
