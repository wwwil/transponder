VERSION?="0.0.1"

DOCKER_IMAGE?=wwwil/transponder
DOCKER_IMAGE_TAG?=$(DOCKER_IMAGE):$(VERSION)

build:
	go build

install:
	go install

docker-build:
	docker build --tag $(DOCKER_IMAGE_TAG) .
	docker image prune --force --filter label=transponder=docker-build