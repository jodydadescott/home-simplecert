home-simplecert-client:
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -v -trimpath -o home-simplecert-client ../../../../client/home-simplecert-client.go

clean:
	$(RM) home-simplecert-client
