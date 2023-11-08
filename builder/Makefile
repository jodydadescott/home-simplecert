default:
	@echo "build what? client, server, or all"

all:
	cd client && $(MAKE)
	cd server && $(MAKE)

clean:
	cd client && $(MAKE) clean
	cd server && $(MAKE) clean

push:
	cd server && $(MAKE) push