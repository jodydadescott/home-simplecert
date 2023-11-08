default:
	@echo "building for local system; chdir into builder and run make"
	env CGO_ENABLED=0 go build -v -ldflags="-w -s" -gcflags=-trimpath=$(GOPATH)/src \
	-asmflags=-trimpath=$(GOPATH)/src -o home-simplecert-server home-simplecert-server.go

clean:
	$(RM) home-simplecert-server
	cd builder && $(MAKE) clean