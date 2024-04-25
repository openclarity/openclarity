<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./assets/logos/VMClarity-logo-dark-bg-horizontal@4x.png">
  <source media="(prefers-color-scheme: light)" srcset="./assets/logos/VMClarity-logo-light-bg-horizontal@4x.png">
  <img alt="VMClarity Logo" src="./assets/logos/VMClarity-logo-light-bg-horizontal@4x.png">
</picture>

[![Slack Invite](https://img.shields.io/badge/Slack-Join-blue?logo=slack)](https://outshift.slack.com/messages/vmclarity)
[![Go Reference](https://pkg.go.dev/badge/github.com/openclarity/vmclarity.svg)](https://pkg.go.dev/github.com/openclarity/vmclarity)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/openclarity/vmclarity/main-merge.yml?style=flat-square&branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/openclarity/vmclarity)](https://goreportcard.com/report/github.com/openclarity/vmclarity)

VMClarity is an open source tool for agentless detection and management of Virtual Machine
Software Bill Of Materials (SBOM) and security threats such as vulnerabilities, exploits, malware, rootkits, misconfigurations and leaked secrets.

<img src="./assets/vmclarity_demo.gif" alt="VMClarity demo" />

Join [VMClarity's Slack channel](https://outshift.slack.com/messages/vmclarity) to hear about the latest announcements and upcoming activities. We would love to get your feedback!

# Table of Contents<!-- omit in toc -->

- [Why VMClarity?](#why-vmclarity)
- [Getting started](#getting-started)
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

# Getting started

For step-by-step guidance on how to deploy VMClarity across different environments, including AWS, Azure, GCP, and Docker, click on [this link](https://openclarity.io/docs/vmclarity/getting-started/) and choose your preferred provider for detailed deployment instructions.

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
  - [Windows Registry](cli/analyzer/windows)
- Vulnerability detection
  - [Grype](https://github.com/anchore/grype)
  - [Trivy](https://github.com/aquasecurity/trivy)
- Exploits
  - [Go exploit db](https://github.com/vulsio/go-exploitdb)
- Secrets
  - [gitleaks](https://github.com/gitleaks/gitleaks)
- Malware
  - [ClamAV](https://github.com/Cisco-Talos/clamav)
- Misconfiguration
  - [Lynis](https://github.com/CISOfy/lynis)
  - [CIS Docker Benchmark](https://github.com/goodwithtech/dockle)
- Rootkits
  - [Chkrootkit](https://github.com/Magentron/chkrootkit)
- Security scanning plugins
  - [Plugins](plugins)

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
