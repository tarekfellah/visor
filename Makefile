build: fmt compile

compile:
	go build
	go build -o bin/visor ./cli

fmt:
	find . -name "*.go" -exec gofmt -w {} \;
