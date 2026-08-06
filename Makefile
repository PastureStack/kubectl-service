TARGETS := $(shell ls scripts)

DAPPER_IMAGE ?= pasturestack-kubectl-service-dapper:ubuntu26
DAPPER_HOST_ARCH ?= amd64
DOCKER_BUILDKIT ?= 1
DAPPER_TRIVY_CACHE ?= $(HOME)/.cache/trivy
export DOCKER_BUILDKIT

.PHONY: $(TARGETS) deps trash trash-keep dapper-image dapper-trivy-cache

dapper-image:
	docker build \
		$(if $(DOCKER_BUILD_NETWORK),--network $(DOCKER_BUILD_NETWORK),) \
		--build-arg DAPPER_HOST_ARCH=$(DAPPER_HOST_ARCH) \
		-t $(DAPPER_IMAGE) \
		-f Dockerfile.dapper .

dapper-trivy-cache:
	@test -r "$(DAPPER_TRIVY_CACHE)/db/trivy.db" -a -r "$(DAPPER_TRIVY_CACHE)/db/metadata.json" || { \
		echo "A pre-fetched Trivy vulnerability database is required at $(DAPPER_TRIVY_CACHE)/db" >&2; \
		exit 1; \
	}

ci package release: dapper-trivy-cache

$(TARGETS): dapper-image
	docker run --rm \
		--user "$$(id -u):$$(id -g)" \
		--group-add "$$(stat -c '%g' /var/run/docker.sock)" \
		-v $(CURDIR):/go/src/github.com/PastureStack/kubectl-service \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v "$(DAPPER_TRIVY_CACHE):/tmp/trivy-cache" \
		-e HOME=/tmp/pasturestack-dapper-home-$$(id -u) \
		-e GOCACHE=/tmp/go-build-cache-$$(id -u) \
		-e XDG_CONFIG_HOME=/tmp/go-config-$$(id -u) \
		-e GIT_CONFIG_GLOBAL=/tmp/gitconfig-$$(id -u) \
		-e GOMAXPROCS=2 \
		-e TRIVY_CACHE_DIR=/tmp/trivy-cache \
		-e DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) \
		-e ARCH=$(DAPPER_HOST_ARCH) \
		-e IMAGE_NAME \
		-e TILLER_IMAGE_NAME \
		-e TILLER_TAG \
		-e HELM_MIGRATION_IMAGE_NAME \
		-e HELM_MIGRATION_TAG \
		-e TAG \
		-e VERSION_OVERRIDE \
		-e SOURCE_REVISION \
		-e DOCKER_BUILD_NETWORK \
		-e GO_VERSION \
		-e GO_SHA256_amd64 \
		-e KUBERNETES_SOURCE_VERSION \
		-e KUBERNETES_SERVER_SHA512 \
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
