.PHONY: build
build:
	go build -o bin/syncweather ./cmd/syncweather
	chmod a+x bin/syncweather

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/syncweather_linux_amd64 ./cmd/syncweather
	chmod a+x bin/syncweather_linux_amd64
