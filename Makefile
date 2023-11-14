default:
	@echo "change directory into what you want to build or run make all"
	exit 2

all:
	cd builder && $(MAKE)

clean:
	cd builder && $(MAKE) clean