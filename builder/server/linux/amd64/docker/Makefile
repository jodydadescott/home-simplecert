BUILD_NUMBER := latest
PROJECT_NAME := home-simplecert-server-amd64
DOCKER_REGISTRY := jodydadescott
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)

home-simplecert-server:
	cd ../bin && $(MAKE)
	cp ../bin/home-simplecert-server .
	docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

push:
	$(MAKE) home-simplecert-server
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

clean:
	$(RM) home-simplecert-server