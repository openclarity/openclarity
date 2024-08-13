#!/bin/bash
set -euo pipefail

## Arguments
if [ $# -ne 1 ]
then
    echo "No branch name passed to script."
    echo "Usage: sync-branch <branch-name>"
    exit 1
fi
VMCLARITY_IMPORT_BRANCH="$1"

## Resolved vars
VMCLARITY_REPO="openclarity/vmclarity"
OPENCLARITY_REPO="openclarity/openclarity"
OPENCLARITY_BACKPORT_FEATURE_BRANCH="vmclarity/$VMCLARITY_IMPORT_BRANCH"
OPENCLARITY_UNIFICATION_BASE_BRANCH="unification/base"

###################################################################################################
## Function to pull dependencies, OpenClarity codebase, and VMClarity remote references.
function setup() {
  # Download dependencies
  curl https://raw.githubusercontent.com/newren/git-filter-repo/main/git-filter-repo > /tmp/git-filter-repo

  # Clone OpenClarity repo and navigate into
  rm -rf /tmp/openclarity
  git clone "https://github.com/$OPENCLARITY_REPO" /tmp/openclarity
  cd /tmp/openclarity
  git checkout $OPENCLARITY_UNIFICATION_BASE_BRANCH

  # Add reference to VMClarity
  git remote add --fetch vmclarity-repo "https://github.com/$VMCLARITY_REPO"
}

function cleanup() {
  # Remove openclarity code
  rm -rf /tmp/openclarity
}

trap cleanup EXIT

###################################################################################################
## This function imports VMClarity into an existing unification base branch.
## It also updates the created branch to include historical data from VMClarity.
##
##  Usage:
##    import_vmclarity_branch <branch to import from VMClarity> <branch in OpenClarity to import code in>
##
##  Notes:
##    - Requires setup function
##    - If the imported branch already exists, this will fail
##
function import_vmclarity_branch() {
  local branch_to_import=${1}
  local imported_branch=${2}

  ## Switch to import
  git checkout -b $imported_branch $OPENCLARITY_UNIFICATION_BASE_BRANCH

  ## Merge remote vmclarity branch to our import branch
  git merge --no-commit --allow-unrelated-histories vmclarity-repo/$branch_to_import

  ## Push code
  git commit -m "chore(unification): sync remote branch openclarity/vmclarity/$branch_to_import"
  git push --set-upstream origin $imported_branch

  ## Overwrite history to include vmclarity references
  python3 /tmp/git-filter-repo --force --refs $imported_branch --message-callback '
      if message.startswith(b"kubeclarity: "):
        return message

      if b"vmclarity#" in message:
        return message

      message = re.sub(b"(#\\d+)", b"vmclarity\\1", message, flags=re.MULTILINE)
      return message
  '

  ## Update history
  git push -f
}

###################################################################################################
function openclarity_branch_exists() {
  local branch=${1}

  if [[ -z $(git branch --list ${branch}) ]]; then
      echo 0
  else
      echo 1
  fi
}

function vmclarity_branch_exists() {
  local branch=${1}

  git ls-remote --exit-code --heads vmclarity-repo $branch >/dev/null 2>&1
  EXIT_CODE=$?

  if [[ $EXIT_CODE == '0' ]]; then
    echo 1
  elif [[ $EXIT_CODE == '2' ]]; then
    echo 0
  fi
}

function notify() {
  echo
  echo "====> ${1}"
  echo
}

###################################################################################################
## This function sync VMClarity backport branch in OpenClarity with changes from VMClarity branch.
##
##  Usage:
##    sync_backport_branch <backport branch> <import branch>
##
function sync_backport_branch() {
  local backport_branch=${1}
  local imported_branch=${2}

  ## Switch or create backport branch
  git checkout $backport_branch || git checkout -b $backport_branch $imported_branch && \
                                            git push --set-upstream origin $backport_branch

  ## Rebase backport against imported branch
  git rebase $imported_branch
  git push -f
}

###################################################################################################
#################################### SYNC CODEBASE ################################################
###################################################################################################

## Setup
notify "Setting up environment and pulling dependencies..."
setup

## Check if valid branch on VMClarity
if [[ "$(vmclarity_branch_exists $VMCLARITY_IMPORT_BRANCH)" == "0" ]]; then
    echo "ERROR: Branch $VMCLARITY_IMPORT_BRANCH does not exist on VMClarity"
    exit 1
fi

## Create temporary branch (remove if exists)
tmp_import_branch="tmp/$OPENCLARITY_BACKPORT_FEATURE_BRANCH"

# Remove if exists since its only used as an intermediary step for syncing
if [[ "$(openclarity_branch_exists $tmp_import_branch)" == "1" ]]; then
    notify "Temporary branch $tmp_import_branch already exists, removing..."

    git branch -d $tmp_import_branch || echo "not found locally"
    git push origin --delete $tmp_import_branch || echo "not found on remote"
fi

## Import VMClarity code into temporary branch
notify "Importing branch $VMCLARITY_IMPORT_BRANCH from VMClarity..."
import_vmclarity_branch $VMCLARITY_IMPORT_BRANCH $tmp_import_branch

## Sync imported code with our backport
notify "Syncing branch $VMCLARITY_IMPORT_BRANCH from VMClarity to $OPENCLARITY_BACKPORT_FEATURE_BRANCH..."
sync_backport_branch $OPENCLARITY_BACKPORT_FEATURE_BRANCH $tmp_import_branch

## Remove temporary branch
git branch -d $tmp_import_branch || echo "not found locally"
git push origin --delete $tmp_import_branch || echo "not found on remote"

## Print success
notify "SUCCESS! Branch '$VMCLARITY_IMPORT_BRANCH' from VMClartiy synced into '$OPENCLARITY_BACKPORT_FEATURE_BRANCH'"
