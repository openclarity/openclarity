module github.com/openclarity/openclarity/installation

go 1.22.4

require github.com/openclarity/openclarity/utils v0.7.2

replace (
	github.com/openclarity/openclarity/core => ../core
	github.com/openclarity/openclarity/utils => ../utils
)
