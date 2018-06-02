install:
	go install -v

build:
	go build -v ./...

lint:
	golint ./...
	go vet ./...

test:
	go test -v ./... --cover

deps:
	go get -u github.com/golang/lint/golint
	go get -u github.com/stretchr/testify/assert

clean:
	go clean
