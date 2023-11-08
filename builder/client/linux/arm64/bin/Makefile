home-simplecert-client:
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -v -trimpath -o home-simplecert-client ../../../../../client/home-simplecert-client.go

clean:
	$(RM) -f home-simplecert-client
