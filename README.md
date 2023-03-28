# VMClarity

![VMClarity](https://github.com/openclarity/vmclarity/img/logos/vmclarity_logo.png "VMClarity")

VMClarity is a tool for agentless detection and management of Virtual Machine
Software Bill Of Materials (SBOM) and vulnerabilities

## Table of Contents

- [VMClarity](#vmclarity)
  - [Table of Contents](#table-of-contents)
  - [Getting Started](#getting-started)
    - [Installing on AWS](#installing-on-aws)
    - [Accessing the API](#accessing-the-api)
  - [Contributing](#contributing)
  - [Code of Conduct](#code-of-conduct)
  - [License](#license)

## Getting Started

### Installing on AWS

1. Download the cloud-formation from the VMClarity Github release
2. Go to AWS console Cloudformation for your choosen region
3. Create a stack with new resources
4. Upload the downloaded template
5. Walk through the wizard
6. Monitor install from the cloud-formation page
7. Get the VMClarity public IP address from the Outputs tab.

### Accessing the API

To access the API, a tunnel to the HTTP ports must be opened using the
VMClarity server as a bastion.

```
ssh -N -L 8888:localhost:8888 ubuntu@<VMClarity public IP address>
```

Once this has been run the VMClarity API can be access on localhost:8888. For example:

```
curl http://localhost:8888/api/scanConfigs
```

## Contributing

If you are ready to jump in and test, add code, or help with documentation,
please follow the instructions on our [contributing guide](/CONTRIBUTING.md)
for details on how to open issues, setup VMClarity for development and test.

## Code of Conduct

You can view our code of conduct [here](/CODE_OF_CONDUCT.md).

## License

[Apache License, Version 2.0](/LICENSE)
