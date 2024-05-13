# Documentation available at: https://docs.docker.com/build/bake/

# Docker build args
variable "DOCKER_REGISTRY" {default = "ghcr.io/openclarity"}
variable "DOCKER_TAG" {default = "latest"}
variable "SUFFIX" {default = ""}

# Golang build args
variable "VERSION" {default = null}
variable "BUILD_TIMESTAMP" {default = null}
variable "COMMIT_HASH" {default = null}
variable "VMCLARITY_TOOLS_BASE" {default = null}
variable "BUILD_OPTS" {default = null}

function "get_tag" {
  params = [name]
  result = ["${DOCKER_REGISTRY}/${name}${SUFFIX}:${DOCKER_TAG}"]
}

group "default" {
	targets = [
		"vmclarity-apiserver",
		"vmclarity-cli",
		"vmclarity-cr-discovery-server",
		"vmclarity-orchestrator",
		"vmclarity-ui",
		"vmclarity-ui-backend",
		"vmclarity-plugin-kics",
	]
}

group "vmclarity-scanner-plugins" {
	targets = [
		"vmclarity-plugin-kics",
		"vmclarity-plugin-example-go",
		"vmclarity-plugin-example-python"
	]
}

target "_common" {
	labels = {
		"org.opencontainers.image.url" = "https://github.com/openclarity/vmclarity"
		"org.opencontainers.image.licenses" = "Apache-2.0"
	}
	output = ["type=image"]
}

target "_common_args_for_go" {
	args = {
		VERSION = "${VERSION}"
		BUILD_TIMESTAMP = "${BUILD_TIMESTAMP}"
		COMMIT_HASH = "${COMMIT_HASH}"
		BUILD_OPTS = "${BUILD_OPTS}"
	}
}

target "vmclarity-apiserver" {
	context = "."
	dockerfile = "Dockerfile.apiserver"
	tags = get_tag("${target.vmclarity-apiserver.name}")
	inherits = ["_common", "_common_args_for_go"]
	labels = {
		"org.opencontainers.image.title" = "VMClarity API Server"
		"org.opencontainers.image.description" = "The VMClarity API for managing all objects in the VMClarity system."
	}
}

target "vmclarity-cli" {
	context = "."
	dockerfile = "Dockerfile.cli"
	tags = get_tag("${target.vmclarity-cli.name}")
	inherits = ["_common", "_common_args_for_go"]
	args = {
		VMCLARITY_TOOLS_BASE = "${VMCLARITY_TOOLS_BASE}"
	}
	labels = {
		"org.opencontainers.image.title" = "VMClarity CLI"
		"org.opencontainers.image.description" = "The VMClarity CLI for managing all objects in the VMClarity system."
	}
}

target "vmclarity-cr-discovery-server" {
	context = "."
	dockerfile = "Dockerfile.cr-discovery-server"
	tags = get_tag("${target.vmclarity-cr-discovery-server.name}")
	inherits = ["_common", "_common_args_for_go"]
	labels = {
		"org.opencontainers.image.title" = "Container Runtime Discovery Server"
		"org.opencontainers.image.description" = "Container Runtime Discovery Server for VMClarity."
	}
}

target "vmclarity-orchestrator" {
	context = "."
	dockerfile = "Dockerfile.orchestrator"
	tags = get_tag("${target.vmclarity-orchestrator.name}")
	inherits = ["_common", "_common_args_for_go"]
	labels = {
		"org.opencontainers.image.title" = "VMClarity Orchestrator"
		"org.opencontainers.image.description" = "Orchestrates and manages the life cycle of VMClarity scan configs, scans and asset scans."
	}
}

target "vmclarity-ui" {
	context = "."
	dockerfile = "Dockerfile.ui"
	tags = get_tag("${target.vmclarity-ui.name}")
	inherits = ["_common"]
	labels = {
		"org.opencontainers.image.title" = "VMClarity UI"
		"org.opencontainers.image.description" = "A server serving the UI static files."
	}
}

target "vmclarity-ui-backend" {
	context = "."
	dockerfile = "Dockerfile.uibackend"
	tags = get_tag("${target.vmclarity-ui-backend.name}")
	inherits = ["_common", "_common_args_for_go"]
	labels = {
		"org.opencontainers.image.title" = "VMClarity UI Backend"
		"org.opencontainers.image.description" = "A separate backend API which offloads some processing from the browser to the infrastructure to process and filter data closer to the source."
	}
}

target "vmclarity-plugin-kics" {
	context = "."
	dockerfile = "./plugins/store/kics/Dockerfile"
	tags = get_tag("${target.vmclarity-plugin-kics.name}")
	inherits = ["_common", "_common_args_for_go"]
	labels = {
		"org.opencontainers.image.title" = "VMClarity KICS Scanner"
		"org.opencontainers.image.description" = ""
	}
}

target "vmclarity-plugin-example-go" {
	context = "."
	dockerfile = "./plugins/sdk-go/example/Dockerfile"
	tags = get_tag("${target.vmclarity-plugin-example-go.name}")
	labels = {
		"org.opencontainers.image.title" = "VMClarity Scanner Go Example"
		"org.opencontainers.image.description" = ""
	}
}

target "vmclarity-plugin-example-python" {
	context = "."
	dockerfile = "./plugins/sdk-python/example/Dockerfile.test"
	tags = get_tag("${target.vmclarity-plugin-example-python.name}")
	labels = {
		"org.opencontainers.image.title" = "VMClarity Scanner Python Example"
		"org.opencontainers.image.description" = ""
	}
}
