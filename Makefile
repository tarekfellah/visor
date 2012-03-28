build: fmt compile

compile:
	go build
	go build -o bin/visor ./cmd

fmt:
	find . -name "*.go" -exec gofmt -w {} \;
