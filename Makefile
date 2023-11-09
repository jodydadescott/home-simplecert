default:
	@echo "build what? (local, darwin-amd64, darwin-arm64, linux-amd64, linux-arm64)"
	exit 2

home-simplecert:
	env CGO_ENABLED=0 go build -v -trimpath -o home-simplecert home-simplecert.go

local:
	$(MAKE) home-simplecert

darwin-amd64:
	$(MAKE) build/darwin/amd64/home-simplecert

build/darwin/amd64/home-simplecert:
	mkdir -p build/darwin/amd64
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -v -trimpath -o build/darwin/amd64/home-simplecert home-simplecert.go

darwin-arm64:
	$(MAKE) build/darwin/arm64/home-simplecert

build/darwin/arm64/home-simplecert:
	mkdir -p build/darwin/arm64
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -v -trimpath -o build/darwin/arm64/home-simplecert home-simplecert.go

linux-amd64:
	$(MAKE) build/linux/amd64/home-simplecert

build/linux/amd64/home-simplecert:
	mkdir -p build/linux/amd64
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -trimpath -o build/linux/amd64/home-simplecert home-simplecert.go

linux-arm64:
	$(MAKE) build/linux/arm64/home-simplecert

build/linux/arm64/home-simplecert:
	mkdir -p build/linux/arm64
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -v -trimpath -o build/linux/arm64/home-simplecert home-simplecert.go

clean:
	$(RM) home-simplecert
	$(RM) -r build