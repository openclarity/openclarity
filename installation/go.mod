module github.com/openclarity/openclarity/installation

go 1.22.6

require github.com/openclarity/openclarity/utils v1.0.0

replace (
	github.com/openclarity/openclarity/core => ../core
	github.com/openclarity/openclarity/utils => ../utils
)
