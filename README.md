<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://docs.openclarity.io/img/footer-logos/OC_logo_H_1C_white.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://docs.openclarity.io/img/color-logo/logo.svg">
  <img alt="Openclarity logo" src="https://docs.openclarity.io/img/color-logo/logo.svg" width="50%">
</picture>

KubeClarity is a tool for detection and management of Software Bill Of Materials (SBOM) and vulnerabilities of container images and filesystems. It scans both runtime K8s clusters and CI/CD pipelines for enhanced software supply chain security.

![KubeClarity Dashboard screenshot](https://openclarity.io/docs/kubeclarity/dashboard.png)

KubeClarity is the tool responsible for Kubernetes Security in the [Openclarity platform](https://openclarity.io).

# Table of Contents

- [Why?](#why)
  - [SBOM & Vulnerability detection challenges](#sbom--vulnerability-detection-challenges)
  - [Solution](#solution)
- [Features](#features)
  - [Integrated SBOM generators and vulnerability scanners](#integrated-sbom-generators-and-vulnerability-scanners)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Contributing](#contributing)
- [License](#license)

# Why?
## SBOM & Vulnerability Detection Challenges

* Effective vulnerability scanning requires an accurate Software Bill Of Materials (SBOM) detection:
  * Various programming languages and package managers
  * Various OS distributions
  * Package dependency information is usually stripped upon build
* Which one is the best scanner/SBOM analyzer?
* What should we scan: Git repos, builds, container images or runtime?
* Each scanner/analyzer has its own format - how to compare the results?
* How to manage the discovered SBOM and vulnerabilities?
* How are my applications affected by a newly discovered vulnerability?

## Solution

* Separate vulnerability scanning into 2 phases:
  * Content analysis to generate SBOM
  * Scan the SBOM for vulnerabilities
* Create a pluggable infrastructure to:
  * Run several content analyzers in parallel
  * Run several vulnerability scanners in parallel
* Scan and merge results between different CI stages using KubeClarity CLI
* Runtime K8s scan to detect vulnerabilities discovered post-deployment
* Group scanned resources (images/directories) under defined applications to navigate the object tree dependencies (applications, resources, packages, vulnerabilities)

# Features

* Dashboard
  * Fixable vulnerabilities per severity
  * Top 5 vulnerable elements (applications, resources, packages)
  * New vulnerabilities trends
  * Package count per license type
  * Package count per programming language
  * General counters
* Applications
  * Automatic application detection in K8s runtime
  * Create/edit/delete applications
  * Per application, navigation to related:
    * Resources (images/directories)
    * Packages
    * Vulnerabilities
    * Licenses in use by the resources
* Application Resources (images/directories)
  * Per resource, navigation to related:
    * Applications
    * Packages
    * Vulnerabilities
* Packages
    * Per package, navigation to related:
        * Applications
        * Linkable list of resources and the detecting SBOM analyzers
        * Vulnerabilities
* Vulnerabilities
    * Per vulnerability, navigation to related:
        * Applications
        * Resources
        * List of detecting scanners
* K8s Runtime scan
  * On-demand or scheduled scanning
  * Automatic detection of target namespaces
  * Scan progress and result navigation per affected element (applications, resources, packages, vulnerabilities)
  * CIS Docker benchmark
* CLI (CI/CD)
  * SBOM generation using multiple integrated content analyzers (Syft, cyclonedx-gomod)
  * SBOM/image/directory vulnerability scanning using multiple integrated scanners (Grype, Dependency-track)
  * Merging of SBOM and vulnerabilities across different CI/CD stages
  * Export results to KubeClarity backend
* API
  * The API for KubeClarity can be found [here](https://github.com/openclarity/kubeclarity/blob/master/api/swagger.yaml)

## Integrated SBOM generators and vulnerability scanners
KubeClarity content analyzer integrates with the following SBOM generators:
* [Syft](https://github.com/anchore/syft)
* [Cyclonedx-gomod](https://github.com/CycloneDX/cyclonedx-gomod)
* [Trivy](https://github.com/aquasecurity/trivy)

KubeClarity vulnerability scanner integrates with the following scanners:
* [Grype](https://github.com/anchore/grype)
* [Dependency-Track](https://github.com/DependencyTrack/dependency-track)
* [Trivy](https://github.com/aquasecurity/trivy)

# Architecture

![](images/architecture.png)

# Getting Started

To get started, see the [KubeClarity documentation on the Openclarity site](https://openclarity.io/docs/kubeclarity/getting-started/).

# Contributing

Pull requests and bug reports are welcome.

For larger changes please create an Issue in GitHub first to discuss your
proposed changes and possible implications.

For more details, please see the [Contribution guidelines for this project](https://openclarity.io/docs/contributing/).

## License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)
