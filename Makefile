BUILD_NUMBER := latest
PROJECT_NAME := home-simplecert
DOCKER_REGISTRY := jodydadescott
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)

default:
	@echo "build what? (darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, linux-amd64-container, linux-arm64-container, all-bin, all-container)"
	exit 2

build/darwin/amd64/home-simplecert:
	mkdir -p build/darwin/amd64
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -v -trimpath -o build/darwin/amd64/home-simplecert home-simplecert.go

darwin-amd64:
	@$(MAKE) build/darwin/amd64/home-simplecert

build/darwin/arm64/home-simplecert:
	mkdir -p build/darwin/arm64
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -v -trimpath -o build/darwin/arm64/home-simplecert home-simplecert.go

darwin-arm64:
	@$(MAKE) build/darwin/arm64/home-simplecert

build/linux/amd64/home-simplecert:
	mkdir -p build/linux/amd64
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -trimpath -o build/linux/amd64/home-simplecert home-simplecert.go

linux-amd64:
	@$(MAKE) build/linux/amd64/home-simplecert

build/linux/arm64/home-simplecert:
	mkdir -p build/linux/arm64
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -v -trimpath -o build/linux/arm64/home-simplecert home-simplecert.go

linux-arm64:
	@$(MAKE) build/linux/arm64/home-simplecert

linux-amd64-container:
	$(MAKE) linux-amd64
	mkdir -p build/linux/container/amd64
	cp build/linux/amd64/home-simplecert build/linux/container/amd64
	cat Dockerfile | sed 's#FROM image.*#FROM fedora:37#g' > build/linux/container/amd64/Dockerfile
	cd build/linux/container/amd64 && docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME)-amd64:$(DOCKER_IMAGE_TAG) .

linux-arm64-container:
	$(MAKE) linux-arm64
	mkdir -p build/linux/container/arm64
	cp build/linux/arm64/home-simplecert build/linux/container/arm64
	cat Dockerfile | sed 's#FROM image.*#FROM arm64v8/fedora:37#g' > build/linux/container/arm64/Dockerfile
	cd build/linux/container/arm64 && docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME)-arm64:$(DOCKER_IMAGE_TAG) .

all-bin:
	$(MAKE) darwin-amd64
	$(MAKE) darwin-arm64
	$(MAKE) linux-amd64
	$(MAKE) linux-arm64

all-container:
	$(MAKE) linux-amd64-container
	$(MAKE) linux-arm64-container

clean:
	$(RM) -r build
