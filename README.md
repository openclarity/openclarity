<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./assets/logos/OpenClarity-logo-dark-bg.png">
  <source media="(prefers-color-scheme: light)" srcset="./assets/logos/OpenClarity-logo-light-bg.png">
  <img alt="OpenClarity Logo" src="./assets/logos/OpenClarity-logo-light-bg.png">
  <br/><br/><br/>
</picture>

[![Slack Invite](https://img.shields.io/badge/Slack-Join-blue?logo=slack)](https://outshift.slack.com/messages/vmclarity)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/openclarity/openclarity/main-merge.yml?style=flat-square&branch=main)

<!--TODO: Uncomment these once we have the first tagged version-->
<!--[![Go Reference](https://pkg.go.dev/badge/github.com/openclarity/openclarity.svg)](https://pkg.go.dev/github.com/openclarity/openclarity)-->
<!--[![Go Report Card](https://goreportcard.com/badge/github.com/openclarity/openclarity)](https://goreportcard.com/report/github.com/openclarity/openclarity)-->

OpenClarity is an open source tool for agentless detection and management of Virtual Machine
Software Bill Of Materials (SBOM) and security threats such as vulnerabilities, exploits, malware, rootkits, misconfigurations and leaked secrets.

<img src="./assets/OpenClarity-demo.gif" alt="OpenClarity demo" />

Join [OpenClarity's Slack channel](https://outshift.slack.com/messages/vmclarity) to hear about the latest announcements and upcoming activities. We would love to get your feedback!

# Table of Contents<!-- omit in toc -->

- [Why OpenClarity?](#why-openclarity)
- [Getting started](#getting-started)
- [Overview](#overview)
  - [Usage modes](#usage-modes)
    - [1. OpenClarity stack](#1-openclarity-stack)
    - [2. CLI](#2-cli)
    - [3. Go module](#3-go-module)
  - [Asset discovery](#asset-discovery)
  - [Supported filesystems](#supported-filesystems)
- [Architecture](#architecture)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Code of Conduct](#code-of-conduct)
- [License](#license)

# Why OpenClarity?

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

The OpenClarity project is focused on unifying detection and management of VM security threats in an agentless manner.

# Getting started

For step-by-step guidance on how to deploy OpenClarity across different environments, including AWS, Azure, GCP, and Docker, click on [this link](https://openclarity.io/docs/getting-started/) and choose your preferred provider for detailed deployment instructions.

# Overview

OpenClarity uses a pluggable scanning infrastructure to provide:

- SBOM analysis
- Package and OS vulnerability detection
- Exploit detection
- Leaked secret detection
- Malware detection
- Misconfiguration detection
- Rootkit detection

The pluggable scanning infrastructure uses several tools that can be
enabled/disabled on an individual basis. OpenClarity normalizes, merges and
provides a robust visualization of the results from these various tools.

These tools include:

- SBOM Generation and Analysis
  - [Syft](https://github.com/anchore/syft)
  - [Trivy](https://github.com/aquasecurity/trivy)
  - [Windows Registry](scanner/families/sbom/windows)\*
  - [Cyclonedx-gomod](https://github.com/CycloneDX/cyclonedx-gomod)
- Vulnerability detection
  - [Grype](https://github.com/anchore/grype)
  - [Trivy](https://github.com/aquasecurity/trivy)
- Exploits
  - [Go exploit db](https://github.com/vulsio/go-exploitdb)
- Secrets
  - [gitleaks](https://github.com/gitleaks/gitleaks)
- Malware
  - [ClamAV](https://github.com/Cisco-Talos/clamav)
  - [YARA](https://github.com/virustotal/yara)
- Misconfiguration
  - [Lynis](https://github.com/CISOfy/lynis)\*\*
  - [CIS Docker Benchmark](https://github.com/goodwithtech/dockle)
  - [KICS](https://github.com/Checkmarx/kics)
- Rootkits
  - [Chkrootkit](https://github.com/Magentron/chkrootkit)\*\*

\* Windows only\
\*\* Linux and MacOS only

## Usage modes

OpenClarity can be used multiple ways to fit different needs:

### 1. OpenClarity stack

As a complete stack, OpenClarity provides an integrated solution to

- discover assets in your environment,
- manage scan configurations, schedule and execute scans,
- visualize the results on a dashboard.

For the deployment instructions visit this page: [Getting started](https://openclarity.io/docs/getting-started/).

### 2. CLI

OpenClarity can be used as a standalone command line tool to run the supported scanner tools.

1. Download `openclarity-cli` from the [GitHub releases page](https://github.com/openclarity/openclarity/releases/).
2. Create a configuration file, make sure to enable the scanner families you need. An example can be found here: [.families.yaml](https://github.com/openclarity/openclarity/blob/main/.families.yaml)
3. Execute the following command:

   ```bash
   openclarity-cli scan --config .families.yaml
   ```

### 3. Go module

Import the `github.com/openclarity/openclarity/scanner` package to run a scan with OpenClarity’s family manager from your code.

Example: [scan.go](https://github.com/openclarity/openclarity/blob/94c46f830838416706c2deef71ecce095d706e6a/cli/cmd/scan/scan.go#L121)

## Asset discovery

OpenClarity stack supports the automatic discovery of assets in the following providers:

| Provider   | Asset types                      | Scope                 |
| ---------- | -------------------------------- | --------------------- |
| Docker     | Docker containers and images     | Local Docker daemon   |
| Kubernetes | Docker containers and images     | Cluster               |
| AWS        | Virtual machines (EC2 instances) | Account (all regions) |
| Azure      | Virtual machines                 | Subscription          |
| GCP        | Virtual machines                 | Project               |

## Supported filesystems

The following filesystem operations are supported on different host types:

| Host    | List block devices | Mount Ext2, Ext3, Ext4 | Mount XFS     | Mount NTFS    |
| ------- | ------------------ | ---------------------- | ------------- | ------------- |
| Linux   | Supported          | Supported              | Supported     | Supported     |
| Darwin  | Supported          | Supported              | Supported     | Supported     |
| Windows | Not supported      | Not supported          | Not supported | Not supported |

# Architecture

A high-level architecture overview is available [here](ARCHITECTURE.md).

# Roadmap

OpenClarity project roadmap is available [here](https://github.com/orgs/openclarity/projects/5/views/5).

# Contributing

If you are ready to jump in and test, add code, or help with documentation,
please follow the instructions on our [contributing guide](CONTRIBUTING.md)
for details on how to open issues, setup OpenClarity for development and test.

# Code of Conduct

You can view our code of conduct [here](CODE_OF_CONDUCT.md).

# License

[Apache License, Version 2.0](LICENSE)
