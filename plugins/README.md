# Overview

**Plugins** provide additional **scanning capabilities** to VMClarity ecosystem.
They are executed as standalone containers.
Project structure is defined as:

- **runner** - Provides necessary logic to execute scanner plugins. Used as a library in VMClarity.
- **sdk** - Language-specific libraries, templates, and examples to help with the implementation of scanner plugins.
- **store** - Collection of implemented containerized scanner plugins.

## Architecture
