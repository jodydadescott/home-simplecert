BUILD_NUMBER := latest
PROJECT_NAME := home-simplecert-amd64
DOCKER_REGISTRY := jodydadescott
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)

container:
	cd ../amd64 && make
	cp ../amd64/home-simplecert .
	docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	touch container

clean:
	$(RM) -r home-simplecert
	$(RM) -r container
