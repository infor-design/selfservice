PROJECT=selfservice
IMAGE=selfservice
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist
CLI_NAME=selfservice
BIN_NAME=selfservice
DEV_IMAGE?=false

DOCKER_PUSH?=false
IMAGE_NAMESPACE?=

GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)

ifdef IMAGE_NAMESPACE
IMAGE_PREFIX=${IMAGE_NAMESPACE}/
endif

ifneq (${GIT_TAG},)
IMAGE_TAG=${GIT_TAG}
LDFLAGS += -X ${PACKAGE}.gitTag=${GIT_TAG}
else
IMAGE_TAG?=latest
endif

.PHONY: all
all: image

.PHONY: build
build:
	docker build --cache-from docker.io/$(PROJECT)/$(IMAGE):latest \
		-t docker.io/$(PROJECT)/$(IMAGE):$(VERSION) \
		-t docker.io/$(PROJECT)/$(IMAGE):latest -f Dockerfile .

.PHONY: push
push:
	docker push docker.io/$(PROJECT)/$(IMAGE):latest
	docker push docker.io/$(PROJECT)/$(IMAGE):$(VERSION)

.PHONY: clean-debug
clean-debug:
	-find ${CURRENT_DIR} -name debug.test -exec rm -f {} +

.PHONY: selfservice-all
selfservice-all: clean-debug
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${BIN_NAME} ./cmd

.PHONY: build-ui
build-ui:
	DOCKER_BUILDKIT=1 docker build -t selfservice-ui --target selfservice-ui .
	find ./ui/build -type f -not -name gitkeep -delete
	docker run -v ${CURRENT_DIR}/ui/build:/tmp/app --rm -t selfservice-ui sh -c 'cp -r ./build/* /tmp/app/'

.PHONY: image
ifeq ($(DEV_IMAGE), true)
# The "dev" image builds the binaries from the users desktop environment (instead of in Docker)
# which speeds up builds. Dockerfile.dev needs to be copied into dist to perform the build, since
# the dist directory is under .dockerignore.
IMAGE_TAG="dev-$(shell git describe --always --dirty)"
image:
	DOCKER_BUILDKIT=1 docker build --platform=linux/amd64 -t selfservice-base --target selfservice-base .
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/selfservice ./cmd
	ln -sfn ${DIST_DIR}/selfservice ${DIST_DIR}/selfservice-server
	ln -sfn ${DIST_DIR}/selfservice ${DIST_DIR}/selfservice-reposerver
	ln -sfn ${DIST_DIR}/selfservice ${DIST_DIR}/selfservice-wsserver
	cp Dockerfile.dev dist
	DOCKER_BUILDKIT=1 docker build --platform=linux/amd64 -t $(IMAGE_PREFIX)selfservice:$(IMAGE_TAG) -f dist/Dockerfile.dev dist
else
image:
	DOCKER_BUILDKIT=1 docker build -t $(IMAGE_PREFIX)argocd:$(IMAGE_TAG) .
endif
	@if [ "$(DOCKER_PUSH)" = "true" ] ; then docker push $(IMAGE_PREFIX)argocd:$(IMAGE_TAG) ; fi