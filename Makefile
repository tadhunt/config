all:
	go mod tidy
	go vet
	staticcheck
	go build

test:
	go test -failfast -count=1 -v ./...
