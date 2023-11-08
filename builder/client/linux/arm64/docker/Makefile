BUILD_NUMBER := latest
PROJECT_NAME := home-simplecert-client-arm64
DOCKER_REGISTRY := jodydadescott
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)

home-simplecert-client:
	cd ../bin && $(MAKE)
	cp ../bin/home-simplecert-client .
	docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

push:
	$(MAKE) home-simplecert-client
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

clean:
	$(RM) home-simplecert-client