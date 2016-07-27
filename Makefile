#BUILD_VERBOSE :=
BUILD_VERBOSE := -v

TEST_VERBOSE :=
TEST_VERBOSE := -v

DOCKER_IMAGE_NAME := $$USER/coredns

all: coredns

coredns:
	GOOS=linux go build -a -tags netgo -installsuffix netgo
	#CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo

.PHONY: docker
docker: coredns
	docker build -t $(DOCKER_IMAGE_NAME) .

.PHONY: deps
deps:
	go get ${BUILD_VERBOSE}

.PHONY: test
test:
	go test $(TEST_VERBOSE) ./...

.PHONY: testk8s
testk8s:
	go test $(TEST_VERBOSE) -tags k8s -race -run 'TestK8sIntegration' ./test

.PHONY: clean
clean:
	go clean
	# Remove docker image
	if [ -n `docker images -q $(DOCKER_IMAGE_NAME)` ]; then docker rmi $(DOCKER_IMAGE_NAME) ; fi
