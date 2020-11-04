all: build

TAG?=0.0.1-rc23
FLAGS=
ENVVAR=
GOOS?=darwin
REGISTRY?=686244538589.dkr.ecr.us-east-2.amazonaws.com/release
BASEIMAGE?=alpine:3.9
#BUILD_NUMBER=$$(date +'%Y%m%d-%H%M%S')
BUILD_NUMBER := $(shell bash -c 'echo $$(date +'%Y%m%d-%H%M%S')')

build: clean
	$(ENVVAR) GOOS=$(GOOS) go build -o migrator

clean:
	rm -f migrator

run: build
	./migrator

.PHONY: build
docker-build-image:  build
	 docker build -t migrator:$(TAG) .

.PHONY: build, all, wire, clean, run, set-docker-build-env, docker-build-push, orchestrator,
docker-build-push: docker-build-image
	docker tag migrator:${TAG}  ${REGISTRY}/migrator:${TAG}
	docker push ${REGISTRY}/migrator:${TAG}




