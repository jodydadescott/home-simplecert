default:
	@echo "build what? darwin, linux, or all"

all:
	cd darwin && $(MAKE)
	cd linux && $(MAKE)

clean:
	cd darwin && $(MAKE) clean
	cd linux && $(MAKE) clean

push:
	cd linux && $(MAKE) push