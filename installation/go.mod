module github.com/openclarity/openclarity/installation

go 1.23.2

require github.com/openclarity/openclarity/utils v1.1.3

replace (
	github.com/openclarity/openclarity/core => ../core
	github.com/openclarity/openclarity/utils => ../utils
)
