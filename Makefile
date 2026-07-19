TARGETS := $(shell ls scripts)

DAPPER_IMAGE ?= pasturestack-kubectl-service-dapper:ubuntu26
DAPPER_HOST_ARCH ?= amd64
DOCKER_VERSION ?= 29.5.3
UBUNTU_MIRROR ?= http://archive.ubuntu.com/ubuntu

.PHONY: $(TARGETS) deps trash trash-keep dapper-image

dapper-image:
	docker build \
		$(if $(DOCKER_BUILD_NETWORK),--network $(DOCKER_BUILD_NETWORK),) \
		--build-arg DAPPER_HOST_ARCH=$(DAPPER_HOST_ARCH) \
		--build-arg DOCKER_VERSION=$(DOCKER_VERSION) \
		--build-arg UBUNTU_MIRROR=$(UBUNTU_MIRROR) \
		-t $(DAPPER_IMAGE) \
		-f Dockerfile.dapper .

$(TARGETS): dapper-image
	docker run --rm \
		-v $(CURDIR):/go/src/github.com/PastureStack/kubectl-service \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-e DAPPER_UID=$$(id -u) \
		-e DAPPER_GID=$$(id -g) \
		-e ARCH=$(DAPPER_HOST_ARCH) \
		-e IMAGE_NAME \
		-e TAG \
		-e VERSION_OVERRIDE \
		-e DOCKER_BUILD_NETWORK \
		-e UBUNTU_MIRROR=$(UBUNTU_MIRROR) \
		-e GO_VERSION \
		-e GO_SHA256_amd64 \
		-e KUBERNETES_VERSION \
		-e KUBERNETES_SERVER_SHA256 \
		-e HELM_VERSION \
		-e HELM_GIT_COMMIT \
		-e HELM_PATCH_SHA256 \
		-e GO111MODULE=off \
		$(DAPPER_IMAGE) $@

trash:
	@echo "Dependencies are vendored; no legacy trash dependency sync is required."

trash-keep: trash

deps: trash

.DEFAULT_GOAL := ci
