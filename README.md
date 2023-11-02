<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./img/logos/VMClarity-logo-dark-bg-horizontal@4x.png">
  <source media="(prefers-color-scheme: light)" srcset="./img/logos/VMClarity-logo-light-bg-horizontal@4x.png">
  <img alt="VMClarity Logo" src="./img/logos/VMClarity-logo-light-bg-horizontal@4x.png">
</picture>

[![Slack Invite](https://img.shields.io/badge/Slack-Join-blue?logo=slack)](https://outshift.slack.com/messages/vmclarity)
[![Go Reference](https://pkg.go.dev/badge/github.com/openclarity/vmclarity.svg)](https://pkg.go.dev/github.com/openclarity/vmclarity)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/openclarity/vmclarity/main-merge.yml?style=flat-square&branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/openclarity/vmclarity)](https://goreportcard.com/report/github.com/openclarity/vmclarity)

VMClarity is an open source tool for agentless detection and management of Virtual Machine
Software Bill Of Materials (SBOM) and security threats such as vulnerabilities, exploits, malware, rootkits, misconfigurations and leaked secrets.

<img src="./img/vmclarity_demo.gif" alt="VMClarity demo" />

Join [VMClarity's Slack channel](https://outshift.slack.com/messages/vmclarity) to hear about the latest announcements and upcoming activities. We would love to get your feedback!

# Table of Contents<!-- omit in toc -->

- [Why VMClarity?](#why-vmclarity)
- [Quick Start](#quick-start)
- [Overview](#overview)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Code of Conduct](#code-of-conduct)
- [License](#license)

# Why VMClarity?

Virtual machines (VMs) are the most used service across all hyperscalers. AWS,
Azure, GCP, and others have virtual computing services that are used not only
as standalone VM services but also as the most popular method for hosting
containers (e.g., Docker, Kubernetes).

VMs are vulnerable to multiple threats:
- Software vulnerabilities
- Leaked Secrets/Passwords
- Malware
- System Misconfiguration
- Rootkits

There are many very good open source and commercial-based solutions for
providing threat detection for VMs, manifesting the different threat categories above.

However, there are challenges with assembling and managing these tools yourself:
- Complex installation, configuration, and reporting
- Integration with deployment automation
- Siloed reporting and visualization

The VMClarity project is focused on unifying detection and management of VM security threats in an agentless manner.

# Quick start

## Install VMClarity

### AWS

1. Start the CloudFormation [wizard](https://console.aws.amazon.com/cloudformation/home#/stacks/create/review?stackName=VMClarity&templateURL=https://s3.eu-west-2.amazonaws.com/vmclarity-v0.4.0/VmClarity.cfn), or upload the [latest](https://github.com/openclarity/vmclarity/releases/latest) CloudFormation template 
2. Specify the SSH key to be used to connect to VMClarity under 'KeyName'
3. Once deployed, copy VmClarity SSH Address from the "Outputs" tab

For a detailed installation guide, please see [AWS](installation/aws/README.md).

### Azure

1. Click the [![Deploy To Azure](https://docs.microsoft.com/en-us/azure/templates/media/deploy-to-azure.svg)](https://portal.azure.com/#blade/Microsoft_Azure_CreateUIDef/CustomDeploymentBlade/uri/https%3A%2F%2Fraw.githubusercontent.com%2Fopenclarity%2Fvmclarity%2Fmain%2Finstallation%2Fazure%2Fvmclarity.json/uiFormDefinitionUri/https%3A%2F%2Fraw.githubusercontent.com%2Fopenclarity%2Fvmclarity%2Fmain%2Finstallation%2Fazure%2Fvmclarity-UI.json) button.
2. Fill out the required fields in the wizard
3. Once deployed, copy the VMClarity SSH address from the Outputs tab

### GCP

1. Change directory to `installation/gcp/dm`
2. Copy `vmclarity-config.example.yaml` to `vmclarity-config.yaml`, update with required values.
3. Deploy vmclarity using GCP deployment manager
   ```
   gcloud deployment-manager deployments create <vmclarity deployment name> --config vmclarity-config.yaml
   ```
4. Once deployed, copy the VMClarity SSH IP address from the CLI output.

### Kubernetes

1. helm install -n vmclarity --create-namespace vmclarity ./vmclarity

## Access VMClarity UI

1. Open connection to VMClarity API Gateway either:

   * On AWS, Azure or GCP, open an SSH tunnel to VMClarity VM server
     ```
     ssh -N -L 8080:localhost:80 -i  "<Path to the SSH key specified during install>" ubuntu@<VmClarity SSH Address copied during install>
     ```

   * On Kubernetes port-forward vmclarity-gateway service:
     ```
     kubectl port-forward -n vmclarity service/vmclarity-gateway 8080:80
     ```

2. Access VMClarity UI in the browser: http://localhost:8080/
3. Access the [API](api/openapi.yaml) via http://localhost:8080/api

For a detailed UI tour, please see [tour](TOUR.md).

# Overview

VMClarity uses a pluggable scanning infrastructure to provide:
- SBOM analysis
- Package and OS vulnerability detection
- Exploit detection
- Leaked secret detection
- Malware detection
- Misconfiguration detection
- Rootkit detection

The pluggable scanning infrastructure uses several tools that can be
enabled/disabled on an individual basis. VMClarity normalizes, merges and
provides a robust visualization of the results from these various tools.

These tools include:
- SBOM Generation and Analysis
  - [Syft](https://github.com/anchore/syft)
  - [Trivy](https://github.com/aquasecurity/trivy)
  - [Cyclonedx-gomod](https://github.com/CycloneDX/cyclonedx-gomod)
- Vulnerability detection
  - [Grype](https://github.com/anchore/grype)
  - [Trivy](https://github.com/aquasecurity/trivy)
  - [Dependency-Track](https://github.com/DependencyTrack/dependency-track)
- Exploits
  - [Go exploit db](https://github.com/vulsio/go-exploitdb)
- Secrets
  - [gitleaks](https://github.com/gitleaks/gitleaks)
- Malware
  - [ClamAV](https://github.com/Cisco-Talos/clamav)
- Misconfiguration
  - [Lynis](https://github.com/CISOfy/lynis)
- Rootkits
  - [Chkrootkit](https://github.com/Magentron/chkrootkit)

A high-level architecture overview is available [here](ARCHITECTURE.md)

# Roadmap
VMClarity project roadmap is available [here](https://github.com/orgs/openclarity/projects/5/views/5).

# Contributing

If you are ready to jump in and test, add code, or help with documentation,
please follow the instructions on our [contributing guide](CONTRIBUTING.md)
for details on how to open issues, setup VMClarity for development and test.

# Code of Conduct

You can view our code of conduct [here](CODE_OF_CONDUCT.md).

# License

[Apache License, Version 2.0](LICENSE)
