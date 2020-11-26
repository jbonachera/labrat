VERSION = $(shell git rev-parse HEAD)
DOCKER_BUILD_ARGS = --network host --build-arg https_proxy=${https_proxy} --build-arg BUILT_VERSION=${VERSION}

push:: build
	docker push jbonachera/labrat
build::
	docker build ${DOCKER_BUILD_ARGS} -t jbonachera/labrat .
