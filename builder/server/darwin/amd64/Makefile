home-simplecert-server:
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -v -trimpath -o home-simplecert-server ../../../../server/home-simplecert-server.go

clean:
	$(RM) home-simplecert-server