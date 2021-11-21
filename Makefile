build_macos:
	- cd app && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ../dist/cli