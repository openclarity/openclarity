# Release

This document outlines the process for creating a new release for VMClarity using the [Go MultiMod Releaser](https://github.com/open-telemetry/opentelemetry-go-build-tools/tree/main/multimod). All code block examples provided below correspond to an update to version `v0.7.0`, please update accordingly.

## 1. Update the New Release Version

* Create a new branch for the release version update.
```sh
git checkout -b release/v0.7.0
```

* Modify the `versions.yaml` file to update the version for VMClarity's module-set. Keep in mind that the same version is applied to all modules.
```diff
  vmclarity:
-    version: v0.6.0
+    version: v0.7.0
```

* Commit the changes with a suitable message.
```sh
git add versions.yaml
git commit -m "release: update module set to version v0.7.0"
```

* Run the version verification command to check for any issues.

> [!NOTE]
> If you use `go.work` and `go.work.sum` files in the project, temporarily remove/rename them so it won't interfere with this step.

```sh
make multimod-verify
```

## 2. Bump All Dependencies to the New Release Version

* Run the following command to update all `go.mod` files to the new release version.
```sh
make multimod-prerelease
```

* Review the changes made in the last commit to ensure correctness.

* Push the branch to the GitHub repository.
```sh
git push origin release/v0.7.0
```

* Create a pull request for these changes with a title like "release: prepare version v0.7.0".

## 3. Create and Push Tags

* After the pull request is approved and merged, update your local main branch.
```sh
git checkout main
git pull origin main
```

* To trigger the release workflow, create and push to the repository a release tag for the last commit.
```sh
git tag -a v0.7.0
git push origin v0.7.0
```

Please note that the release tag is not necessarily associated with the "release: prepare version v0.7.0" commit. For example, if any bug fixes were required after this commit, they can be merged and included in the release.

## 4. Publish release

* Wait until the release workflow is completed successfully.

* Navigate to the [Releases page](https://github.com/openclarity/vmclarity/releases) and verify the draft release description as well as the assets listed.

* Once the draft release has been verified, click on `Edit` release and then on `Publish Release`.

## 5. Post release steps

* In the [docs repository](https://github.com/openclarity/docs.openclarity.io), modify the value of `latest_version` in `config/_default/config.toml`, create a pull request and merge after it was reviewed.

```toml
[params]
  latest_version = "0.7.0" # Used in some installation commands
```

* From the release page, download the AWS Cloudformation files (`aws-cloudformation-v0.7.0.tar.gz`), extract the archive locally and upload its contents to the S3 bucket used for storing them for VMClarity installation.
