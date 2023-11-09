BUILD_NUMBER := latest
PROJECT_NAME := home-simplecert
DOCKER_REGISTRY := jodydadescott
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)

default:
	@echo "build what? (darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, linux-amd64-container, linux-arm64-container)"
	exit 2

darwin-amd64:
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -v -trimpath -o home-simplecert-darwin-amd64 home-simplecert.go

darwin-arm64:
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -v -trimpath -o home-simplecert-darwin-arm64 home-simplecert.go

linux-amd64:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -trimpath -o home-simplecert-linux-amd64 home-simplecert.go

linux-arm64:
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -v -trimpath -o home-simplecert-linux-arm64 home-simplecert.go

linux-amd64-container:
	$(MAKE) linux-amd64
	mkdir -p containers/linux-amd64
	cp home-simplecert-linux-amd64 containers/linux-amd64/home-simplecert
	cat Dockerfile | sed 's/FROM fedora.*/FROM fedora:37/g' > containers/linux-amd64/Dockerfile
	cat Dockerfile | sed 's/FROM image.*/FROM fedora:37/g'
	cd containers/linux-amd64 && docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME)-amd64:$(DOCKER_IMAGE_TAG) .

linux-arm64-container:
	$(MAKE) linux-arm64
	mkdir -p containers/linux-arm64
	cp home-simplecert-linux-arm64 containers/linux-arm64/home-simplecert
	cat Dockerfile | sed 's/FROM image.*/FROM arm64v8/fedora:37/g'
	cd containers/linux-arm64 && docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME)-arm64:$(DOCKER_IMAGE_TAG) .

clean:
	$(RM) home-simplecert-darwin-amd64
	$(RM) home-simplecert-darwin-arm64
	$(RM) home-simplecert-linux-amd64
	$(RM) home-simplecert-linux-arm64
	$(RM) -r containers
