In order to use CRI-O runtime with OpenClarity, you need to:
- install `CRIU` on Kubernetes nodes
- enable the `ContainerCheckpoint` feature gate on you cluster
- in OpenClarity's Helm chart enable containerSecurityContext for `crDiscoveryServer`, with the following configuration:
  ```
  containerSecurityContext:
    enabled: true
    privileged: true
    readOnlyRootFilesystem: false

  ```