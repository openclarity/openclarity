---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''

---

<!-- Please use this template while reporting a bug and provide as much info as possible.
-->


#### What happened:

#### What you expected to happen:

#### How to reproduce it (as minimally and precisely as possible):

#### Are there any error messages in KubeClarity logs?
(e.g. `kubectl logs -n kubeclarity --selector=app=kubeclarity`)

#### Anything else we need to know?:

#### Environment:
- Kubernetes version (use `kubectl version --short`):
- Helm version (use `helm version`):
- KubeClarity version (use `kubectl -n kubeclarity exec deploy/kubeclarity -- ./backend version`)
- KubeClarity Helm Chart version (use `helm -n kubeclarity list`)
- Cloud provider or hardware configuration:
- Others:
