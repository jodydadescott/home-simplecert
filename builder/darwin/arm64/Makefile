home-simplecert:
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -v -trimpath -o home-simplecert ../../../home-simplecert.go

clean:
	$(RM) -r home-simplecert