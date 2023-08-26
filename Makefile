.PHONY: build
build:
	go build -o bin/syncweather ./cmd/syncweather
	chmod a+x bin/syncweather

.PHONY: build-aws-lambda
build-aws-lambda:
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bin/bootstrap ./cmd/syncweather
	chmod a+x bin/bootstrap
